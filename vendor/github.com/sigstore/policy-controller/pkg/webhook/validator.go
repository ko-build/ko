//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/oci"
	ociremote "github.com/sigstore/cosign/v2/pkg/oci/remote"
	"github.com/sigstore/cosign/v2/pkg/policy"
	csigs "github.com/sigstore/cosign/v2/pkg/signature"
	"github.com/sigstore/policy-controller/pkg/apis/config"
	policyduckv1beta1 "github.com/sigstore/policy-controller/pkg/apis/duck/v1beta1"
	"github.com/sigstore/policy-controller/pkg/apis/policy/v1alpha1"
	policycontrollerconfig "github.com/sigstore/policy-controller/pkg/config"
	webhookcip "github.com/sigstore/policy-controller/pkg/webhook/clusterimagepolicy"
	rekor "github.com/sigstore/rekor/pkg/client"
	"github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/fulcioroots"
	"github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/tuf"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
)

type Validator struct{}

func NewValidator(ctx context.Context) *Validator {
	return &Validator{}
}

// isDeletedOrStatusUpdate returns true if the resource in question is being
// deleted, is already deleted or Status is being updated. In any of those
// cases, we do not validate the resource
func isDeletedOrStatusUpdate(ctx context.Context, deletionTimestamp *metav1.Time) bool {
	return apis.IsInDelete(ctx) || deletionTimestamp != nil || apis.IsInStatusUpdate(ctx)
}

// This is attached to contexts passed to webhook methods so that if the
// user wants to get the Spec for the PolicyResult we can attach it.
type includeSpecKey struct{}

// IncludeSpec adds the spec to context so it's later available for
// inclusion in PolicyResult. This is safe to call multiple times, first
// one "wins". This is on purpose so that since we call down the various
// levels and we want the highest resource level to be available, otherwise
// everything boils down to PodSpec and it's lossy then.
func IncludeSpec(ctx context.Context, spec interface{}) context.Context {
	if GetIncludeSpec(ctx) == nil {
		return context.WithValue(ctx, includeSpecKey{}, spec)
	}
	return ctx
}

// GetIncludeSpec returns the highest level spec for a resource possible.
// For example, for Deployment it would return Deployment.Spec
func GetIncludeSpec(ctx context.Context) interface{} {
	return ctx.Value(includeSpecKey{})
}

// This is attached to contexts passed to webhook methods so that if the
// user wants to get the ObjectMeta for the PolicyResult we can attach it.
type includeObjectMetaKey struct{}

// This is attached to contexts passed to webhook methods so that if the
// user wants to get the TypeMeta for the PolicyResult we can attach it.
type includeTypeMetaKey struct{}

// IncludeObjectMeta adds the ObjectMeta to context so it's later available for
// inclusion in PolicyResult. This is safe to call multiple times, first
// one "wins". This is on purpose so that since we call down the various
// levels and we want the highest resource level to be available, otherwise
// everything boils down to PodSpec and it's lossy then.
func IncludeObjectMeta(ctx context.Context, meta interface{}) context.Context {
	if GetIncludeObjectMeta(ctx) == nil {
		return context.WithValue(ctx, includeObjectMetaKey{}, meta)
	}
	return ctx
}

// GetIncludeObjectMeta returns the highest level ObjectMeta for a resource
// possible. For example, for Deployment it would return Deployment.Spec
func GetIncludeObjectMeta(ctx context.Context) interface{} {
	return ctx.Value(includeObjectMetaKey{})
}

// IncludeTypeMeta adds the TypeMeta to context so it's later available for
// inclusion in PolicyResult. This is safe to call multiple times, first
// one "wins". This is on purpose so that since we call down the various
// levels and we want the highest resource level to be available, otherwise
// everything boils down to PodSpec and it's lossy then.
func IncludeTypeMeta(ctx context.Context, meta interface{}) context.Context {
	if GetIncludeTypeMeta(ctx) == nil {
		return context.WithValue(ctx, includeTypeMetaKey{}, meta)
	}
	return ctx
}

// GetIncludeTypeMeta returns the highest level TypeMeta for a resource
// possible. For example, for Deployment it would return:
// apiVersion: apps/v1
// kind: Deployment
func GetIncludeTypeMeta(ctx context.Context) interface{} {
	return ctx.Value(includeTypeMetaKey{})
}

// ValidatePodScalable implements policyduckv1beta1.PodScalableValidator
// It is very similar to ValidatePodSpecable, but allows for spec.replicas
// to be decremented. This allows for scaling down pods with non-compliant
// images that would otherwise be forbidden.
func (v *Validator) ValidatePodScalable(ctx context.Context, ps *policyduckv1beta1.PodScalable) *apis.FieldError {
	// If we are deleting (or already deleted) or updating status, don't block.
	if isDeletedOrStatusUpdate(ctx, ps.DeletionTimestamp) {
		return nil
	}

	// If we are being scaled down don't block it.
	if ps.IsScalingDown(ctx) {
		logging.FromContext(ctx).Debugf("Skipping validations due to scale down request %s/%s", &ps.ObjectMeta.Name, &ps.ObjectMeta.Namespace)
		return nil
	}

	// Attach the spec for down the line to be attached if it's required by
	// policy to be included in the PolicyResult.
	ctx = IncludeSpec(ctx, ps.Spec)
	ctx = IncludeObjectMeta(ctx, ps.ObjectMeta)
	ctx = IncludeTypeMeta(ctx, ps.TypeMeta)

	imagePullSecrets := make([]string, 0, len(ps.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range ps.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	ns := getNamespace(ctx, ps.Namespace)
	opt := k8schain.Options{
		Namespace:          ns,
		ServiceAccountName: ps.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}

	return v.validatePodSpec(ctx, ns, ps.Kind, ps.APIVersion, ps.ObjectMeta.Labels, &ps.Spec.Template.Spec, opt).ViaField("spec.template.spec")
}

// ValidatePodSpecable implements duckv1.PodSpecValidator
func (v *Validator) ValidatePodSpecable(ctx context.Context, wp *duckv1.WithPod) *apis.FieldError {
	// If we are deleting (or already deleted) or updating status, don't block.
	if isDeletedOrStatusUpdate(ctx, wp.DeletionTimestamp) {
		return nil
	}

	// Attach the spec/metadata for down the line to be attached if it's
	// required by policy to be included in the PolicyResult.
	ctx = IncludeSpec(ctx, wp.Spec)
	ctx = IncludeObjectMeta(ctx, wp.ObjectMeta)
	ctx = IncludeTypeMeta(ctx, wp.TypeMeta)

	imagePullSecrets := make([]string, 0, len(wp.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range wp.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	ns := getNamespace(ctx, wp.Namespace)
	opt := k8schain.Options{
		Namespace:          ns,
		ServiceAccountName: wp.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}
	return v.validatePodSpec(ctx, ns, wp.Kind, wp.APIVersion, wp.ObjectMeta.Labels, &wp.Spec.Template.Spec, opt).ViaField("spec.template.spec")
}

// ValidatePod implements duckv1.PodValidator
func (v *Validator) ValidatePod(ctx context.Context, p *duckv1.Pod) *apis.FieldError {
	// If we are deleting (or already deleted) or updating status, don't block.
	if isDeletedOrStatusUpdate(ctx, p.DeletionTimestamp) {
		return nil
	}

	// Attach the spec/metadata for down the line to be attached if it's
	// required by policy to be included in the PolicyResult.
	ctx = IncludeSpec(ctx, p.Spec)
	ctx = IncludeObjectMeta(ctx, p.ObjectMeta)

	imagePullSecrets := make([]string, 0, len(p.Spec.ImagePullSecrets))
	for _, s := range p.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	ns := getNamespace(ctx, p.Namespace)
	opt := k8schain.Options{
		Namespace:          ns,
		ServiceAccountName: p.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}
	return v.validatePodSpec(ctx, ns, p.Kind, p.APIVersion, p.ObjectMeta.Labels, &p.Spec, opt).ViaField("spec")
}

// ValidateCronJob implements duckv1.CronJobValidator
func (v *Validator) ValidateCronJob(ctx context.Context, c *duckv1.CronJob) *apis.FieldError {
	// If we are deleting (or already deleted) or updating status, don't block.
	if isDeletedOrStatusUpdate(ctx, c.DeletionTimestamp) {
		return nil
	}

	// Attach the spec/metadata for down the line to be attached if it's
	// required by policy to be included in the PolicyResult.
	ctx = IncludeSpec(ctx, c.Spec)
	ctx = IncludeObjectMeta(ctx, c.ObjectMeta)
	ctx = IncludeTypeMeta(ctx, c.TypeMeta)

	imagePullSecrets := make([]string, 0, len(c.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range c.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	ns := getNamespace(ctx, c.Namespace)
	opt := k8schain.Options{
		Namespace:          ns,
		ServiceAccountName: c.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}

	return v.validatePodSpec(ctx, ns, c.Kind, c.APIVersion, c.ObjectMeta.Labels, &c.Spec.JobTemplate.Spec.Template.Spec, opt).ViaField("spec.jobTemplate.spec.template.spec")
}

func (v *Validator) validatePodSpec(ctx context.Context, namespace, kind, apiVersion string, labels map[string]string, ps *corev1.PodSpec, opt k8schain.Options) (errs *apis.FieldError) {
	kc, err := k8schain.New(ctx, kubeclient.Get(ctx), opt)
	if err != nil {
		logging.FromContext(ctx).Warnf("Unable to build k8schain: %v", err)
		return apis.ErrGeneric(err.Error(), apis.CurrentField)
	}

	type containerCheckResult struct {
		index                int
		containerCheckResult *apis.FieldError
	}
	checkContainers := func(cs []corev1.Container, field string) {
		results := make(chan containerCheckResult, len(cs))
		wg := new(sync.WaitGroup)
		for i, c := range cs {
			i := i
			c := c
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Require digests, otherwise the validation is meaningless
				// since the tag can move.
				fe := refOrFieldError(c.Image, field, i)
				if fe != nil {
					results <- containerCheckResult{index: i, containerCheckResult: fe}
					return
				}

				containerErrors := v.validateContainerImage(ctx, c.Image, namespace, field, i, kind, apiVersion, labels, kc, ociremote.WithRemoteOptions(
					remote.WithContext(ctx),
					remote.WithAuthFromKeychain(kc),
				))
				results <- containerCheckResult{index: i, containerCheckResult: containerErrors}
			}()
		}
		for i := 0; i < len(cs); i++ {
			select {
			case <-ctx.Done():
				errs = errs.Also(apis.ErrGeneric("context was canceled before validation completed"))
			case result, ok := <-results:
				if !ok {
					errs = errs.Also(apis.ErrGeneric("results channel failed to produce a result"))
				} else {
					errs = errs.Also(result.containerCheckResult)
				}
			}
		}
		wg.Wait()
	}

	checkEphemeralContainers := func(cs []corev1.EphemeralContainer, field string) {
		results := make(chan containerCheckResult, len(cs))
		wg := new(sync.WaitGroup)
		for i, c := range cs {
			i := i
			c := c
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Require digests, otherwise the validation is meaningless
				// since the tag can move.
				fe := refOrFieldError(c.Image, field, i)
				if fe != nil {
					results <- containerCheckResult{index: i, containerCheckResult: fe}
					return
				}

				containerErrors := v.validateContainerImage(ctx, c.Image, namespace, field, i, kind, apiVersion, labels, kc, ociremote.WithRemoteOptions(
					remote.WithContext(ctx),
					remote.WithAuthFromKeychain(kc),
				))
				results <- containerCheckResult{index: i, containerCheckResult: containerErrors}
			}()
		}
		for i := 0; i < len(cs); i++ {
			select {
			case <-ctx.Done():
				errs = errs.Also(apis.ErrGeneric("context was canceled before validation completed"))
			case result, ok := <-results:
				if !ok {
					errs = errs.Also(apis.ErrGeneric("results channel failed to produce a result"))
				} else {
					errs = errs.Also(result.containerCheckResult)
				}
			}
		}
		wg.Wait()
	}

	checkContainers(ps.InitContainers, "initContainers")
	checkContainers(ps.Containers, "containers")
	checkEphemeralContainers(ps.EphemeralContainers, "ephemeralContainers")

	return errs
}

// setNoMatchingPoliciesError returns nil if the no matching policies behaviour
// has been set to allow or has not been set. Otherwise returns either a warning
// or error based on the NoMatchPolicy.
func setNoMatchingPoliciesError(ctx context.Context, image, field string, index int) *apis.FieldError {
	// Check what the configuration is and act accordingly.
	pcConfig := policycontrollerconfig.FromContextOrDefaults(ctx)

	noMatchingPolicyError := apis.ErrGeneric("no matching policies", "image").ViaFieldIndex(field, index)
	noMatchingPolicyError.Details = image
	if pcConfig == nil {
		// This should not happen, but handle it as fail close
		return noMatchingPolicyError
	}
	switch pcConfig.NoMatchPolicy {
	case policycontrollerconfig.AllowAll:
		// Allow it through, nothing to do.
		return nil
	case policycontrollerconfig.DenyAll:
		return noMatchingPolicyError
	case policycontrollerconfig.WarnAll:
		return noMatchingPolicyError.At(apis.WarningLevel)
	default:
		// Fail closed.
		return noMatchingPolicyError
	}
}

// validatePolicies will go through all the matching Policies and their
// Authorities for a given image. Returns the map of policy=>Validated
// signatures. From the map you can see the number of matched policies along
// with the signatures that were verified.
// If there's a policy that did not match, it will be returned in the errors map
// along with all the errors that caused it to fail.
// Note that if an image does not match any policies, it's perfectly
// reasonable that the return value is 0, nil since there were no errors, but
// the image was not validated against any matching policy and hence authority.
func validatePolicies(ctx context.Context, namespace string, ref name.Reference, policies map[string]webhookcip.ClusterImagePolicy, kc authn.Keychain, remoteOpts ...ociremote.Option) (map[string]*PolicyResult, map[string][]error) {
	type retChannelType struct {
		name         string
		policyResult *PolicyResult
		errors       []error
	}
	results := make(chan retChannelType, len(policies))

	wg := new(sync.WaitGroup)

	// For each matching policy it must validate at least one Authority within
	// it.
	// From the Design document, the part about multiple Policies matching:
	// "If multiple policies match a particular image, then ALL of those
	// policies must be satisfied for the image to be admitted."
	// If none of the Authorities for a given policy pass the checks, gather
	// the errors here. If one passes, do not return the errors.
	for cipName, cip := range policies {
		// Due to running in gofunc
		cipName := cipName
		cip := cip
		logging.FromContext(ctx).Debugf("Checking Policy: %s", cipName)
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := retChannelType{name: cipName}

			result.policyResult, result.errors = ValidatePolicy(ctx, namespace, ref, cip, kc, remoteOpts...)
			// Cache the result.
			FromContext(ctx).Set(ctx, ref.Name(), cipName, string(cip.UID), cip.ResourceVersion, &CacheResult{
				PolicyResult: result.policyResult,
				Errors:       result.errors,
			})
			results <- result
		}()
	}
	// Gather all validated policies here.
	policyResults := make(map[string]*PolicyResult)
	// For a policy that does not pass at least one authority, gather errors
	// here so that we can give meaningful errors to the user.
	ret := map[string][]error{}

	for i := 0; i < len(policies); i++ {
		select {
		case <-ctx.Done():
			ret["internalerror"] = append(ret["internalerror"], fmt.Errorf("context was canceled before validation completed"))
		case result, ok := <-results:
			if !ok {
				ret["internalerror"] = append(ret["internalerror"], fmt.Errorf("results channel failed to produce a result"))
				continue
			}
			switch {
			// Return AuthorityMatches before errors, since even if there
			// are errors, if there are 0 or more authorities that match,
			// it will pass the Policy. Of course, a CIP level policy can
			// override this behaviour, but that has been checked above and
			// if it failed, it will nil out the policyResult.
			case result.policyResult != nil:
				policyResults[result.name] = result.policyResult
			case len(result.errors) > 0:
				ret[result.name] = append(ret[result.name], result.errors...)
			default:
				ret[result.name] = append(ret[result.name], fmt.Errorf("failed to process policy: %s", result.name))
			}
		}
	}
	wg.Wait()
	return policyResults, ret
}

func asFieldError(warn bool, err error) *apis.FieldError {
	r := &apis.FieldError{Message: err.Error()}
	if warn {
		return r.At(apis.WarningLevel)
	}
	return r.At(apis.ErrorLevel)
}

// ValidatePolicy will go through all the Authorities for a given image/policy
// and return validated authorities if at least one of the Authorities
// validated the signatures OR attestations if atttestations were specified.
// Returns PolicyResult if one or more authorities matched, otherwise nil.
// In any case returns all errors encountered if none of the authorities
// passed.
// kc is the Keychain to use for fetching ConfigFile that's independent of the
// signatures / attestations.
func ValidatePolicy(ctx context.Context, namespace string, ref name.Reference, cip webhookcip.ClusterImagePolicy, kc authn.Keychain, remoteOpts ...ociremote.Option) (*PolicyResult, []error) {
	// Check the cache and return if hit, otherwise, check the policy
	cacheResult := FromContext(ctx).Get(ctx, ref.String(), string(cip.UID), cip.ResourceVersion)
	if cacheResult != nil {
		return cacheResult.PolicyResult, cacheResult.Errors
	}

	// Each gofunc creates and puts one of these into a results channel.
	// Once each gofunc finishes, we go through the channel and pull out
	// the results.
	type retChannelType struct {
		name         string
		static       bool
		attestations map[string][]PolicyAttestation
		signatures   []PolicySignature
		err          error
	}
	wg := new(sync.WaitGroup)

	results := make(chan retChannelType, len(cip.Authorities))
	for _, authority := range cip.Authorities {
		authority := authority // due to gofunc
		logging.FromContext(ctx).Debugf("Checking Authority: %s", authority.Name)

		wg.Add(1)
		go func() {
			defer wg.Done()
			result := retChannelType{name: authority.Name}
			// Assignment for appendAssign lint error
			authorityRemoteOpts := remoteOpts
			authorityRemoteOpts = append(authorityRemoteOpts, authority.RemoteOpts...)

			signaturePullSecretsOpts, err := authority.SourceSignaturePullSecretsOpts(ctx, namespace)
			if err != nil {
				result.err = err
				results <- result
				return
			}
			authorityRemoteOpts = append(authorityRemoteOpts, signaturePullSecretsOpts...)

			switch {
			case authority.Static != nil:
				if authority.Static.Action == "fail" {
					result.err = cosign.NewVerificationError("disallowed by static policy")
					results <- result
					return
				}
				result.static = true

			case len(authority.Attestations) > 0:
				// We're doing the verify-attestations path, so validate (.att)
				result.attestations, result.err = ValidatePolicyAttestationsForAuthority(ctx, ref, authority, authorityRemoteOpts...)

			default:
				result.signatures, result.err = ValidatePolicySignaturesForAuthority(ctx, ref, authority, authorityRemoteOpts...)
			}
			results <- result
		}()
	}

	// If none of the Authorities for a given policy pass the checks, gather
	// the errors here. Even if there are errors, return the matched
	// authoritypolicies.
	authorityErrors := make([]error, 0, len(cip.Authorities))
	// We collect all the successfully satisfied Authorities into this and
	// return it.
	policyResult := &PolicyResult{
		AuthorityMatches: make(map[string]AuthorityMatch, len(cip.Authorities)),
	}
	for range cip.Authorities {
		select {
		case <-ctx.Done():
			authorityErrors = append(authorityErrors, fmt.Errorf("%w before validation completed", ctx.Err()))

		case result, ok := <-results:
			if !ok {
				authorityErrors = append(authorityErrors, errors.New("results channel closed before all results were sent"))
				continue
			}
			switch {
			case result.err != nil:
				// We only wrap actual policy failures as FieldErrors with the
				// possibly Warn level. Other things imho should be still
				// be considered errors.
				authorityErrors = append(authorityErrors, asFieldError(cip.Mode == "warn", result.err))

			case len(result.signatures) > 0:
				policyResult.AuthorityMatches[result.name] = AuthorityMatch{Signatures: result.signatures}

			case len(result.attestations) > 0:
				policyResult.AuthorityMatches[result.name] = AuthorityMatch{Attestations: result.attestations}

			case result.static:
				// This happens when we encounter a policy with:
				//   static:
				//     action: "pass"
				policyResult.AuthorityMatches[result.name] = AuthorityMatch{
					Static: true,
				}

			default:
				authorityErrors = append(authorityErrors, fmt.Errorf("failed to process authority: %s", result.name))
			}
		}
	}
	wg.Wait()
	// Even if there are errors, return the policies, since as per the
	// spec, we just need one authority to pass checks. If more than
	// one are required, that is enforced at the CIP policy level.
	// If however there are no authorityMatches, return nil so we don't have
	// to keep checking the length on the returned calls.
	if len(policyResult.AuthorityMatches) == 0 {
		return nil, authorityErrors
	}
	// Ok, there's at least one valid authority that matched. If there's a CIP
	// level policy, validate it here before returning.
	if cip.Policy != nil {
		if cip.Policy.FetchConfigFile != nil && *cip.Policy.FetchConfigFile {
			logging.FromContext(ctx).Debug("Fetching ConfigFiles")
			// It's unfortunate that we have to keep having the kc here. It
			// would be nice if we could just unwrap/generate the ggcr remote
			// options from the oci remote options, but for now this is how
			// we're rolling.
			rOpts := []remote.Option{
				remote.WithContext(ctx),
				remote.WithAuthFromKeychain(kc),
			}
			configFiles, errs := getConfigs(ctx, ref, rOpts...)
			if len(errs) > 0 {
				for _, e := range errs {
					authorityErrors = append(authorityErrors, asFieldError(cip.Mode == "warn", e))
				}
				return nil, authorityErrors
			}
			policyResult.Config = configFiles
		}
		if cip.Policy.IncludeSpec != nil && *cip.Policy.IncludeSpec {
			policyResult.Spec = GetIncludeSpec(ctx)
		}
		if cip.Policy.IncludeObjectMeta != nil && *cip.Policy.IncludeObjectMeta {
			policyResult.ObjectMeta = GetIncludeObjectMeta(ctx)
		}
		if cip.Policy.IncludeTypeMeta != nil && *cip.Policy.IncludeTypeMeta {
			policyResult.TypeMeta = GetIncludeTypeMeta(ctx)
		}

		logging.FromContext(ctx).Info("Validating CIP level policy")
		policyJSON, err := json.Marshal(policyResult)
		if err != nil {
			return nil, append(authorityErrors, err)
		}
		logging.FromContext(ctx).Infof("CIP level policy: %s", string(policyJSON))
		err = policy.EvaluatePolicyAgainstJSON(ctx, "ClusterImagePolicy", cip.Policy.Type, cip.Policy.Data, policyJSON)
		if err != nil {
			logging.FromContext(ctx).Warnf("Failed to validate CIP level policy against %s", string(policyJSON))
			return nil, append(authorityErrors, asFieldError(cip.Mode == "warn", err))
		}
	}
	return policyResult, authorityErrors
}

func ociSignatureToPolicySignature(ctx context.Context, sigs []oci.Signature) []PolicySignature {
	ret := make([]PolicySignature, 0, len(sigs))
	for _, ociSig := range sigs {
		logging.FromContext(ctx).Debugf("Converting signature %+v", ociSig)

		id, err := ociSig.Digest()
		if err != nil {
			logging.FromContext(ctx).Debugf("Error fetching signature digest %+v", err)
			continue
		}

		if cert, err := ociSig.Cert(); err == nil && cert != nil {
			ce := cosign.CertExtensions{
				Cert: cert,
			}
			ret = append(ret, PolicySignature{
				ID:      id.Hex,
				Subject: csigs.CertSubject(cert),
				Issuer:  ce.GetIssuer(),
				GithubExtensions: GithubExtensions{
					WorkflowTrigger: ce.GetCertExtensionGithubWorkflowTrigger(),
					WorkflowSHA:     ce.GetExtensionGithubWorkflowSha(),
					WorkflowName:    ce.GetCertExtensionGithubWorkflowName(),
					WorkflowRepo:    ce.GetCertExtensionGithubWorkflowRepository(),
					WorkflowRef:     ce.GetCertExtensionGithubWorkflowRef(),
				},
			})
		} else {
			ret = append(ret, PolicySignature{
				ID: id.Hex,
				// TODO(mattmoor): Is there anything we should encode for key-based?
			})
		}
	}
	return ret
}

// attestation is used to accumulate the signature along with extracted and
// validated metadata during validation to construct a list of
// PolicyAttestations upon completion without needing to refetch any of the
// parts.
type attestation struct {
	oci.Signature

	PredicateType string
	Payload       []byte
}

func attestationToPolicyAttestations(ctx context.Context, atts []attestation) []PolicyAttestation {
	ret := make([]PolicyAttestation, 0, len(atts))
	for _, att := range atts {
		logging.FromContext(ctx).Debugf("Converting attestation %+v", att)

		id, err := att.Digest()
		if err != nil {
			logging.FromContext(ctx).Debugf("Error fetching attestation digest %+v", err)
			continue
		}

		if cert, err := att.Cert(); err == nil && cert != nil {
			ce := cosign.CertExtensions{
				Cert: cert,
			}
			ret = append(ret, PolicyAttestation{
				PolicySignature: PolicySignature{
					ID:      id.Hex,
					Subject: csigs.CertSubject(cert),
					Issuer:  ce.GetIssuer(),
					GithubExtensions: GithubExtensions{
						WorkflowTrigger: ce.GetCertExtensionGithubWorkflowTrigger(),
						WorkflowSHA:     ce.GetExtensionGithubWorkflowSha(),
						WorkflowName:    ce.GetCertExtensionGithubWorkflowName(),
						WorkflowRepo:    ce.GetCertExtensionGithubWorkflowRepository(),
						WorkflowRef:     ce.GetCertExtensionGithubWorkflowRef(),
					},
				},
				PredicateType: att.PredicateType,
				Payload:       att.Payload,
			})
		} else {
			ret = append(ret, PolicyAttestation{
				PolicySignature: PolicySignature{
					ID: id.Hex,
					// TODO(mattmoor): Is there anything we should encode for key-based?
				},
				PredicateType: att.PredicateType,
				Payload:       att.Payload,
			})
		}
	}
	return ret
}

// ValidatePolicySignaturesForAuthority takes the Authority and tries to
// verify a signature against it.
func ValidatePolicySignaturesForAuthority(ctx context.Context, ref name.Reference, authority webhookcip.Authority, remoteOpts ...ociremote.Option) ([]PolicySignature, error) {
	name := authority.Name

	checkOpts, err := checkOptsFromAuthority(ctx, authority, remoteOpts...)
	if err != nil {
		logging.FromContext(ctx).Errorf("failed constructing checkOpts for %s: +v", name, err)
		return nil, fmt.Errorf("constructing checkOpts for %s: %w", name, err)
	}
	switch {
	case authority.Key != nil:
		if len(authority.Key.PublicKeys) == 0 {
			return nil, fmt.Errorf("there are no public keys for authority %s", name)
		}
		// TODO(vaikas): What should happen if there are multiple keys
		// Is it even allowed? 'valid' returns success if any key
		// matches.
		// https://github.com/sigstore/policy-controller/issues/1652
		sps, err := valid(ctx, ref, authority.Key.PublicKeys, authority.Key.HashAlgorithmCode, checkOpts)
		if err != nil {
			return nil, fmt.Errorf("signature key validation failed for authority %s for %s: %w", name, ref.Name(), err)
		}
		logging.FromContext(ctx).Debugf("validated signature for %s for authority %s got %d signatures", ref.Name(), authority.Name, len(sps))
		return ociSignatureToPolicySignature(ctx, sps), nil

	case authority.Keyless != nil:
		if authority.Keyless.URL != nil {
			sps, err := validSignatures(ctx, ref, checkOpts)
			if err != nil {
				logging.FromContext(ctx).Errorf("failed validSignatures for authority %s with fulcio for %s: %v", name, ref.Name(), err)
				return nil, fmt.Errorf("signature keyless validation failed for authority %s for %s: %w", name, ref.Name(), err)
			}
			logging.FromContext(ctx).Debugf("validated signature for %s, got %d signatures", ref.Name(), len(sps))
			return ociSignatureToPolicySignature(ctx, sps), nil
		}
		return nil, fmt.Errorf("no Keyless URL specified")
	case authority.RFC3161Timestamp != nil:
		sps, err := validSignatures(ctx, ref, checkOpts)
		if err != nil {
			logging.FromContext(ctx).Errorf("failed validSignatures for authority %s with fulcio for %s: %v", name, ref.Name(), err)
			return nil, fmt.Errorf("signature TSA validation failed for authority %s for %s: %w", name, ref.Name(), err)
		}
		logging.FromContext(ctx).Debugf("validated TSA signature for %s, got %d signatures", ref.Name(), len(sps))
		return ociSignatureToPolicySignature(ctx, sps), nil
	}

	// This should never happen because authority has to have been validated to
	// be either having a Key, Keyless, or Static (handled elsewhere)
	return nil, errors.New("authority has neither key, keyless, or static specified")
}

// ValidatePolicyAttestationsForAuthority takes the Authority and tries to
// verify attestations against it.
func ValidatePolicyAttestationsForAuthority(ctx context.Context, ref name.Reference, authority webhookcip.Authority, remoteOpts ...ociremote.Option) (map[string][]PolicyAttestation, error) {
	name := authority.Name
	checkOpts, err := checkOptsFromAuthority(ctx, authority, remoteOpts...)
	if err != nil {
		logging.FromContext(ctx).Errorf("failed creating checkopts client: %v", err)
		return nil, fmt.Errorf("creating CheckOpts: %w", err)
	}

	verifiedAttestations := []oci.Signature{}
	switch {
	case authority.Key != nil && len(authority.Key.PublicKeys) > 0:
		for _, k := range authority.Key.PublicKeys {
			verifier, err := signature.LoadVerifier(k, authority.Key.HashAlgorithmCode)
			if err != nil {
				logging.FromContext(ctx).Errorf("error creating verifier: %v", err)
				return nil, fmt.Errorf("creating verifier: %w", err)
			}
			checkOpts.SigVerifier = verifier
			va, err := validAttestations(ctx, ref, checkOpts)
			if err != nil {
				logging.FromContext(ctx).Errorf("error validating attestations: %v", err)
				return nil, fmt.Errorf("attestation key validation failed for authority %s for %s: %w", name, ref.Name(), err)
			}
			verifiedAttestations = append(verifiedAttestations, va...)
		}

	case authority.Keyless != nil:
		if authority.Keyless != nil && authority.Keyless.URL != nil {
			va, err := validAttestations(ctx, ref, checkOpts)
			if err != nil {
				logging.FromContext(ctx).Errorf("failed validAttestationsWithFulcio for authority %s with fulcio for %s: %v", name, ref.Name(), err)
				return nil, fmt.Errorf("attestation keyless validation failed for authority %s for %s: %w", name, ref.Name(), err)
			}
			verifiedAttestations = append(verifiedAttestations, va...)
		}
	case authority.RFC3161Timestamp != nil:
		va, err := validAttestations(ctx, ref, checkOpts)
		if err != nil {
			logging.FromContext(ctx).Errorf("failed validAttestations for authority %s with fulcio for %s: %v", name, ref.Name(), err)
			return nil, fmt.Errorf("signature TSA validAttestations failed for authority %s for %s: %w", name, ref.Name(), err)
		}
		logging.FromContext(ctx).Debugf("validated TSA signature for %s, got %d signatures", ref.Name(), len(va))
		verifiedAttestations = append(verifiedAttestations, va...)
	}

	// If we didn't get any verified attestations either from the Key or Keyless
	// path, then error out
	if len(verifiedAttestations) == 0 {
		logging.FromContext(ctx).Errorf("no valid attestations found for authority %s for %s", name, ref.Name())
		return nil, fmt.Errorf("%w for authority %s for %s", cosign.ErrNoMatchingAttestations, name, ref.Name())
	}
	logging.FromContext(ctx).Debugf("Found %d valid attestations, validating policies for them", len(verifiedAttestations))

	// Now spin through the Attestations that the user specified and validate
	// them.
	// TODO(vaikas): Pretty inefficient here, figure out a better way if
	// possible.
	ret := make(map[string][]PolicyAttestation, len(authority.Attestations))

	for _, wantedAttestation := range authority.Attestations {
		// Since there can be multiple verified attestations that matched, for
		// example multiple 'custom' attestations. We keep the first error that
		// we encounter here but do not exit on it, in case another attestation
		// satisfies the policy.
		var reterror error
		// There's a particular type, so we need to go through all the verified
		// attestations and make sure that our particular one is satisfied.
		checkedAttestations := make([]attestation, 0, len(verifiedAttestations))
		for _, va := range verifiedAttestations {
			attBytes, err := policy.AttestationToPayloadJSON(ctx, wantedAttestation.PredicateType, va)
			if err != nil {
				if reterror == nil {
					// Only stash the first error
					reterror = err
				}
				logging.FromContext(ctx).Warnf("failed to convert attestation payload to json: %v", err)
				continue
			}
			if attBytes == nil {
				// This happens when we ask for a predicate type that this
				// attestation is not for. It's not an error, so we skip it.
				continue
			}
			if wantedAttestation.Type != "" {
				if err := policy.EvaluatePolicyAgainstJSON(ctx, wantedAttestation.Name, wantedAttestation.Type, wantedAttestation.Data, attBytes); err != nil {
					if reterror == nil {
						// Only stash the first error
						reterror = err
					}
					logging.FromContext(ctx).Warnf("failed policy validation for %s: %v", wantedAttestation.Name, err)
					continue
				}
			}
			// Ok, so this passed aok, jot it down to our result set as
			// verified attestation with the predicate type match
			checkedAttestations = append(checkedAttestations, attestation{
				Signature:     va,
				PredicateType: wantedAttestation.PredicateType,
				Payload:       attBytes,
			})
		}
		if len(checkedAttestations) == 0 {
			if reterror != nil {
				// If there was a matching policy, but it failed to be validated
				// then return that more specific error instead of the more
				// generic 'no matching attestations'.
				return nil, reterror
			}
			return nil, fmt.Errorf("%w with type %s", cosign.ErrNoMatchingAttestations, wantedAttestation.PredicateType)
		}
		ret[wantedAttestation.Name] = attestationToPolicyAttestations(ctx, checkedAttestations)
	}
	return ret, nil
}

// ResolvePodScalable implements policyduckv1beta1.PodScalableValidator
func (v *Validator) ResolvePodScalable(ctx context.Context, ps *policyduckv1beta1.PodScalable) {
	// Don't mess with things that are being deleted or already deleted or
	// if status is being updated
	if isDeletedOrStatusUpdate(ctx, ps.DeletionTimestamp) {
		return
	}

	if ps.IsScalingDown(ctx) {
		logging.FromContext(ctx).Debugf("Skipping validations due to scale down request %s/%s", &ps.ObjectMeta.Name, &ps.ObjectMeta.Namespace)
		return
	}

	imagePullSecrets := make([]string, 0, len(ps.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range ps.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	opt := k8schain.Options{
		Namespace:          getNamespace(ctx, ps.Namespace),
		ServiceAccountName: ps.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}
	v.resolvePodSpec(ctx, &ps.Spec.Template.Spec, opt)
}

// ResolvePodSpecable implements duckv1.PodSpecValidator
func (v *Validator) ResolvePodSpecable(ctx context.Context, wp *duckv1.WithPod) {
	// Don't mess with things that are being deleted or already deleted or
	// status update.
	if isDeletedOrStatusUpdate(ctx, wp.DeletionTimestamp) {
		return
	}

	imagePullSecrets := make([]string, 0, len(wp.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range wp.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	opt := k8schain.Options{
		Namespace:          getNamespace(ctx, wp.Namespace),
		ServiceAccountName: wp.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}
	v.resolvePodSpec(ctx, &wp.Spec.Template.Spec, opt)
}

// ResolvePod implements duckv1.PodValidator
func (v *Validator) ResolvePod(ctx context.Context, p *duckv1.Pod) {
	// Don't mess with things that are being deleted or already deleted or
	// status update.
	if isDeletedOrStatusUpdate(ctx, p.DeletionTimestamp) {
		return
	}
	imagePullSecrets := make([]string, 0, len(p.Spec.ImagePullSecrets))
	for _, s := range p.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	opt := k8schain.Options{
		Namespace:          getNamespace(ctx, p.Namespace),
		ServiceAccountName: p.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}
	v.resolvePodSpec(ctx, &p.Spec, opt)
}

// ResolveCronJob implements duckv1.CronJobValidator
func (v *Validator) ResolveCronJob(ctx context.Context, c *duckv1.CronJob) {
	// Don't mess with things that are being deleted or already deleted or
	// status update.
	if isDeletedOrStatusUpdate(ctx, c.DeletionTimestamp) {
		return
	}

	imagePullSecrets := make([]string, 0, len(c.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets))
	for _, s := range c.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, s.Name)
	}
	opt := k8schain.Options{
		Namespace:          getNamespace(ctx, c.Namespace),
		ServiceAccountName: c.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName,
		ImagePullSecrets:   imagePullSecrets,
	}
	v.resolvePodSpec(ctx, &c.Spec.JobTemplate.Spec.Template.Spec, opt)
}

// For testing
var remoteResolveDigest = ociremote.ResolveDigest

func (v *Validator) resolvePodSpec(ctx context.Context, ps *corev1.PodSpec, opt k8schain.Options) {
	kc, err := k8schain.New(ctx, kubeclient.Get(ctx), opt)
	if err != nil {
		logging.FromContext(ctx).Warnf("Unable to build k8schain: %v", err)
		return
	}

	resolveContainers := func(cs []corev1.Container) {
		for i, c := range cs {
			ref, err := name.ParseReference(c.Image)
			if err != nil {
				logging.FromContext(ctx).Debugf("Unable to parse reference: %v", err)
				continue
			}

			// If we are in the context of a mutating webhook, then resolve the tag to a digest.
			switch {
			case apis.IsInCreate(ctx), apis.IsInUpdate(ctx):
				digest, err := remoteResolveDigest(ref, ociremote.WithRemoteOptions(
					remote.WithContext(ctx),
					remote.WithAuthFromKeychain(kc),
				))
				if err != nil {
					logging.FromContext(ctx).Debugf("Unable to resolve digest %q: %v", ref.String(), err)
					continue
				}
				cs[i].Image = digest.String()
			}
		}
	}

	resolveEphemeralContainers := func(cs []corev1.EphemeralContainer) {
		for i, c := range cs {
			ref, err := name.ParseReference(c.Image)
			if err != nil {
				logging.FromContext(ctx).Debugf("Unable to parse reference: %v", err)
				continue
			}

			// If we are in the context of a mutating webhook, then resolve the tag to a digest.
			switch {
			case apis.IsInCreate(ctx), apis.IsInUpdate(ctx):
				digest, err := remoteResolveDigest(ref, ociremote.WithRemoteOptions(
					remote.WithContext(ctx),
					remote.WithAuthFromKeychain(kc),
				))
				if err != nil {
					logging.FromContext(ctx).Debugf("Unable to resolve digest %q: %v", ref.String(), err)
					continue
				}
				cs[i].Image = digest.String()
			}
		}
	}

	resolveContainers(ps.InitContainers)
	resolveContainers(ps.Containers)
	resolveEphemeralContainers(ps.EphemeralContainers)
}

// getNamespace tries to extract the namespace from the HTTPRequest
// if the namespace passed as argument is empty. This is a workaround
// for a bug in k8s <= 1.24.
func getNamespace(ctx context.Context, namespace string) string {
	if namespace == "" {
		r := apis.GetHTTPRequest(ctx)
		if r != nil && r.Body != nil {
			var review admissionv1.AdmissionReview
			if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
				logging.FromContext(ctx).Errorf("could not decode body: %v", err)
				return ""
			}
			return review.Request.Namespace
		}
	}
	return namespace
}

// validateContainer will validate the container image, and any errors will use
// field & index to craft the meaningful error message.
// field is necessary because higher level resources come here from different
// contexts and the container could be nested at different levels in the
// resource
// index is the number in the containers array from the said context.
//
// Returns any encountered errors, or nil in two cases:
// All the matched policies were validated, or
// no matching policies were found, but the PolicyControllerConfig has been
// configured to allow images not matching any policies.
func (v *Validator) validateContainerImage(ctx context.Context, containerImage string, namespace, field string, index int, kind, apiVersion string, labels map[string]string, kc authn.Keychain, ociRemoteOpts ...ociremote.Option) *apis.FieldError {
	ref, err := name.ParseReference(containerImage)
	if err != nil {
		return apis.ErrGeneric(err.Error(), "image").ViaFieldIndex(field, index)
	}
	config := config.FromContext(ctx)

	if config != nil {
		policies, err := config.ImagePolicyConfig.GetMatchingPolicies(ref.Name(), kind, apiVersion, labels)
		if err != nil {
			errorField := apis.ErrGeneric(err.Error(), "image").ViaFieldIndex(field, index)
			errorField.Details = containerImage
			return errorField
		}

		// If there is at least one policy that matches, that means it
		// has to be satisfied.
		if len(policies) > 0 {
			signatures, fieldErrors := validatePolicies(ctx, namespace, ref, policies, kc, ociRemoteOpts...)
			if len(signatures) != len(policies) {
				logging.FromContext(ctx).Warnf("Failed to validate at least one policy for %s wanted %d policies, only validated %d", ref.Name(), len(policies), len(signatures))
			} else {
				logging.FromContext(ctx).Infof("Validated %d policies for image %s", len(signatures), containerImage)
			}
			return errorsToFieldErrors(containerImage, field, index, fieldErrors)
		}
		// Container matched no policies, so return based on the configured
		// NoMatchPolicy.
		return setNoMatchingPoliciesError(ctx, containerImage, field, index)
	}
	return nil
}

func errorsToFieldErrors(image, field string, index int, fieldErrors map[string][]error) (errs *apis.FieldError) {
	// Do we really want to add all the error details here?
	// Seems like we can just say which policy failed, so
	// doing that for now.
	// Split the errors and warnings to their own
	// error levels.
	hasWarnings := false
	hasErrors := false
	for failingPolicy, policyErrs := range fieldErrors {
		errDetails := image
		warnDetails := image
		for _, policyErr := range policyErrs {
			var fe *apis.FieldError
			if errors.As(policyErr, &fe) {
				if fe.Filter(apis.WarningLevel) != nil {
					warnDetails = warnDetails + " " + fe.Message
					hasWarnings = true
				} else {
					errDetails = errDetails + " " + fe.Message
					hasErrors = true
				}
			} else {
				// Just a regular error.
				errDetails = errDetails + " " + policyErr.Error()
			}
		}
		if hasWarnings {
			warnField := apis.ErrGeneric(fmt.Sprintf("failed policy: %s", failingPolicy), "image").ViaFieldIndex(field, index)
			warnField.Details = warnDetails
			errs = errs.Also(warnField).At(apis.WarningLevel)
		}
		if hasErrors {
			errorField := apis.ErrGeneric(fmt.Sprintf("failed policy: %s", failingPolicy), "image").ViaFieldIndex(field, index)
			errorField.Details = errDetails
			errs = errs.Also(errorField)
		}
	}
	return
}

// refOrFieldError parses the given image into a name.Reference, or returns
// a properly constructed FieldError for a given field/index in the resource
// spec.
func refOrFieldError(image, field string, index int) *apis.FieldError {
	ref, err := name.ParseReference(image)
	if err != nil {
		return apis.ErrGeneric(err.Error(), "image").ViaFieldIndex(field, index)
	}
	if _, ok := ref.(name.Digest); !ok {
		return apis.ErrInvalidValue(
			fmt.Sprintf("%s must be an image digest", image),
			"image",
		).ViaFieldIndex(field, index)
	}
	return nil
}

// configFileResult is used to communicate results from gofuncs that fetch
// ConfigFiles for a given image.
// Because this can be recursive (say, multi-arch image), returns a map where
// key is the architecture of the container image.
type configFileResult struct {
	ret  map[string]*v1.ConfigFile
	errs []error
}

// getConfigs will fetch ConfigFile(s) for a given image. In case the image
// is an index, we'll fetch the arch images recursively.
func getConfigs(ctx context.Context, ref name.Reference, options ...remote.Option) (map[string]*v1.ConfigFile, []error) {
	descriptor, err := remote.Get(ref, options...)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get ref %s : %w", ref.String(), err)}
	}
	switch descriptor.MediaType {
	case types.OCIImageIndex, types.DockerManifestList:
		ii, err := descriptor.ImageIndex()
		if err != nil {
			return nil, []error{fmt.Errorf("getting ImageIndex for %s : %w", ref.String(), err)}
		}
		im, err := ii.IndexManifest()
		if err != nil {
			return nil, []error{fmt.Errorf("getting IndexManifest for %s : %w", ref.String(), err)}
		}
		wg := new(sync.WaitGroup)

		results := make(chan configFileResult, len(im.Manifests))
		for _, manifest := range im.Manifests {
			manifest := manifest
			wg.Add(1)
			go func() {
				defer wg.Done()
				newRefString := ref.Context().Digest(manifest.Digest.String()).String()
				newRef, err := name.ParseReference(newRefString)
				if err != nil {
					results <- configFileResult{ret: nil, errs: []error{fmt.Errorf("failed to ParseReference for: %s: %w", newRefString, err)}}
					return
				}

				newRefConfigs, errs := getConfigs(ctx, newRef, options...)
				results <- configFileResult{ret: newRefConfigs, errs: errs}
			}()
		}
		errs := []error{}
		ret := make(map[string]*v1.ConfigFile, len(im.Manifests))
		for i := 0; i < len(im.Manifests); i++ {
			select {
			case <-ctx.Done():
				errs = append(errs, errors.New("context canceled"))
			case result, ok := <-results:
				if !ok {
					errs = append(errs, errors.New("channel closed before all results were gathered"))
				} else {
					if len(result.errs) != 0 {
						errs = append(errs, fmt.Errorf("failed to get a ConfigFile: %v", result.errs))
					} else {
						for k, v := range result.ret {
							ret[k] = v
						}
					}
				}
			}
		}
		wg.Wait()
		if len(errs) > 0 {
			return nil, errs
		}
		return ret, nil
	case types.OCIManifestSchema1, types.DockerManifestSchema2:
		// This is an Image, so just return it.
		image, err := descriptor.Image()
		if err != nil {
			return nil, []error{fmt.Errorf("getting Image for %s: %w", ref.String(), err)}
		}
		cf, err := image.ConfigFile()
		if err != nil {
			return nil, []error{fmt.Errorf("getting ConfigFile for %s: %w", ref.String(), err)}
		}
		return map[string]*v1.ConfigFile{normalizeArchitecture(cf): cf}, nil
	default:
		return nil, []error{fmt.Errorf("unknown mime type for %s: %v", ref.String(), descriptor.MediaType)}
	}
}

// normalizeArchitecture normalizes the os/architecture/variant to:
// {OS}/{Architecture}[/{Variant}]
//
// Some examples are:
// linux/arm64
// linux/arm/v7
// linux/arm/v6
func normalizeArchitecture(cf *v1.ConfigFile) string {
	return v1.Platform{
		Architecture: cf.Architecture,
		OS:           cf.OS,
		OSVersion:    cf.OSVersion,
		Variant:      cf.Variant,
	}.String()
}

// checkOptsFromAuthority creates the necessary options for calling Cosign
// verify functions (signatures and attestations).
func checkOptsFromAuthority(ctx context.Context, authority webhookcip.Authority, remoteOpts ...ociremote.Option) (*cosign.CheckOpts, error) {
	ret := &cosign.CheckOpts{
		RegistryClientOpts: remoteOpts,
	}

	// Add in the identities for verification purposes, as well as Fulcio URL
	// and certificates
	if authority.Keyless != nil {
		for _, id := range authority.Keyless.Identities {
			ret.Identities = append(ret.Identities,
				cosign.Identity{
					Issuer:        id.Issuer,
					Subject:       id.Subject,
					IssuerRegExp:  id.IssuerRegExp,
					SubjectRegExp: id.SubjectRegExp})
		}
		fulcioRoots, fulcioIntermediates, ctlogKeys, err := fulcioCertsFromAuthority(ctx, authority.Keyless)
		if err != nil {
			return nil, fmt.Errorf("getting Fulcio certs: %s: %w", authority.Name, err)
		}
		ret.RootCerts = fulcioRoots
		ret.IntermediateCerts = fulcioIntermediates
		ret.CTLogPubKeys = ctlogKeys
	}
	rekorClient, rekorPubKeys, err := rekorClientAndKeysFromAuthority(ctx, authority.CTLog)
	if err != nil {
		return nil, fmt.Errorf("getting Rekor public keys: %s: %w", authority.Name, err)
	}
	ret.RekorClient = rekorClient
	ret.RekorPubKeys = rekorPubKeys
	// Skip the TLog verification if we have no client or keys to validate
	// against.
	if ret.RekorClient == nil {
		if ret.RekorPubKeys == nil {
			ret.SkipTlogVerify = true
		} else {
			// If there's keys however, use offline for verification.
			ret.Offline = true
		}
	}
	if authority.RFC3161Timestamp != nil && authority.RFC3161Timestamp.TrustRootRef != "" {
		logging.FromContext(ctx).Debug("Using RFC3161Timestamp...")
		// TODO: By default, we disable any tlog verification when using the RFC3161Timestamp validation.
		// There are use cases when the validation is only handled by TSA, and there isn't any TLog involved.
		ret.SkipTlogVerify = true

		sigstoreKeys, err := sigstoreKeysFromContext(ctx, authority.RFC3161Timestamp.TrustRootRef)
		if err != nil {
			return nil, err
		}
		sk, ok := sigstoreKeys.SigstoreKeys[authority.RFC3161Timestamp.TrustRootRef]
		if !ok {
			return nil, fmt.Errorf("trustRootRef %s not found", authority.RFC3161Timestamp.TrustRootRef)
		}
		for _, timestampAuthority := range sk.TimeStampAuthorities {
			leaves, intermediates, roots, err := splitPEMCertificateChain(timestampAuthority.CertChain)
			if err != nil {
				return nil, fmt.Errorf("error splitting certificates: %w", err)
			}
			if len(leaves) > 1 {
				return nil, fmt.Errorf("certificate chain must contain at most one TSA certificate")
			}
			if len(leaves) == 1 {
				ret.TSACertificate = leaves[0]
			}
			ret.TSAIntermediateCertificates = intermediates
			ret.TSARootCertificates = roots
		}
	}
	return ret, nil
}

func sigstoreKeysFromContext(ctx context.Context, trustRootRef string) (*config.SigstoreKeysMap, error) {
	config := config.FromContext(ctx)
	if config == nil {
		// No config, can't fetch certificates, bail.
		return nil, fmt.Errorf("trustRootRef %s not found, config missing", trustRootRef)
	}
	if config.SigstoreKeysConfig == nil {
		// No config, can't fetch keys, bail.
		return nil, fmt.Errorf("trustRootRef %s not found, SigstoreKeys missing", trustRootRef)
	}
	return config.SigstoreKeysConfig, nil
}

// fulcioCertsFromAuthority gets the necessary Fulcio certificates, this is
// rootPool and an optional intermediatePool. Additionally fetches the CTLog
// public keys.
// Preference is given to TrustRoot if specified, from which the certificates
// are fetched and returned. If there's no TrustRoot, the certificates are
// fetched from embedded or cached TUF root.
func fulcioCertsFromAuthority(ctx context.Context, keylessRef *webhookcip.KeylessRef) (*x509.CertPool, *x509.CertPool, *cosign.TrustedTransparencyLogPubKeys, error) {
	// If this is not Keyless, there's no Fulcio, so just return
	if keylessRef.TrustRootRef == "" {
		roots, err := fulcioroots.Get()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to fetch Fulcio roots: %w", err)
		}
		intermediates, err := fulcioroots.GetIntermediates()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to fetch Fulcio intermediates: %w", err)
		}
		ctPubs, err := cosign.GetCTLogPubs(ctx)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to fetch CTLog public keys: %w", err)
		}
		return roots, intermediates, ctPubs, nil
	}

	// There's TrustRootRef, so fetch it
	trustRootRef := keylessRef.TrustRootRef
	sigstoreKeys, err := sigstoreKeysFromContext(ctx, trustRootRef)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting SigstoreKeys: %w", err)
	}
	rootCertsPool := x509.NewCertPool()
	intermediateCertsPool := x509.NewCertPool()

	sk, ok := sigstoreKeys.SigstoreKeys[trustRootRef]
	if !ok {
		return nil, nil, nil, fmt.Errorf("trustRootRef %s not found", trustRootRef)
	}
	for _, ca := range sk.CertificateAuthorities {
		certs, err := cryptoutils.UnmarshalCertificatesFromPEM(ca.CertChain)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error unmarshalling certificates: %w", err)
		}
		for _, cert := range certs {
			// root certificates are self-signed
			if bytes.Equal(cert.RawSubject, cert.RawIssuer) {
				rootCertsPool.AddCert(cert)
			} else {
				intermediateCertsPool.AddCert(cert)
			}
		}
	}

	ctlogKeys := &cosign.TrustedTransparencyLogPubKeys{
		Keys: make(map[string]cosign.TransparencyLogPubKey, len(sk.CTLogs)),
	}
	for i, ctlog := range sk.CTLogs {
		pk, err := cryptoutils.UnmarshalPEMToPublicKey(ctlog.PublicKey)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unmarshaling public key %d failed: %w", i, err)
		}
		ctlogKeys.Keys[ctlog.LogID] = cosign.TransparencyLogPubKey{
			PubKey: pk,
			Status: tuf.Active,
		}
	}
	if len(ctlogKeys.Keys) == 0 {
		// if keys are empty just return a nil map to make easier for the caller
		// to see if it's empty.
		ctlogKeys = nil
	}
	return rootCertsPool, intermediateCertsPool, ctlogKeys, nil
}

// rekorClientAndKeysFromAuthority creates a Rekor client that should be used
// and public keys to go with it.
// Note that if Rekor is not specified, it's not an error and nil will be
// returned for it.
// Preference is given to TrustRoot if specified, from which the URL and public
// keys are fetched and returned. If there's no TrustRoot but a URL, then
// a Rekor client is returned and the keys from the embedded or cached TUF root.
func rekorClientAndKeysFromAuthority(ctx context.Context, tlog *v1alpha1.TLog) (*client.Rekor, *cosign.TrustedTransparencyLogPubKeys, error) {
	if tlog == nil {
		return nil, nil, nil
	}
	if tlog.TrustRootRef != "" {
		trustRootRef := tlog.TrustRootRef
		rekorPubKeys, rekorURL, err := rekorKeysFromTrustRef(ctx, trustRootRef)
		if err != nil {
			return nil, nil, fmt.Errorf("fetching keys for trustRootRef: %w", err)
		}
		if rekorURL == "" && tlog.URL != nil {
			// Pull this from the tlog entry in this case.
			rekorURL = tlog.URL.String()
		}
		rekorClient, err := rekor.GetRekorClient(rekorURL)
		if err != nil {
			logging.FromContext(ctx).Errorf("failed creating rekor client: %v", err)
			return nil, nil, fmt.Errorf("creating Rekor client: %w", err)
		}
		return rekorClient, rekorPubKeys, nil
	}

	// No TrustRoot, so see if there's one specified in the authority and if
	// not just return that no Rekor is to be used.
	if tlog.URL == nil {
		return nil, nil, nil
	}
	rekorClient, err := rekor.GetRekorClient(tlog.URL.String())
	if err != nil {
		logging.FromContext(ctx).Errorf("failed creating rekor client: %v", err)
		return nil, nil, fmt.Errorf("creating Rekor client: %w", err)
	}
	rekorPubKeys, err := cosign.GetRekorPubs(ctx)
	if err != nil {
		logging.FromContext(ctx).Errorf("failed getting rekor public keys: %v", err)
		return nil, nil, fmt.Errorf("getting Rekor public keys: %w", err)
	}
	return rekorClient, rekorPubKeys, nil
}

func rekorKeysFromTrustRef(ctx context.Context, trustRootRef string) (*cosign.TrustedTransparencyLogPubKeys, string, error) {
	sigstoreKeys, err := sigstoreKeysFromContext(ctx, trustRootRef)
	if err != nil {
		return nil, "", fmt.Errorf("getting SigstoreKeys: %w", err)
	}

	if sk, ok := sigstoreKeys.SigstoreKeys[trustRootRef]; ok {
		retKeys := &cosign.TrustedTransparencyLogPubKeys{
			Keys: make(map[string]cosign.TransparencyLogPubKey, len(sk.TLogs)),
		}
		rekorURL := ""
		for i, tlog := range sk.TLogs {
			pk, err := cryptoutils.UnmarshalPEMToPublicKey(tlog.PublicKey)
			if err != nil {
				return nil, "", fmt.Errorf("unmarshaling public key %d failed: %w", i, err)
			}
			// This needs to be ecdsa instead of crypto.PublicKey
			// https://github.com/sigstore/cosign/issues/2540
			pkecdsa, ok := pk.(*ecdsa.PublicKey)
			if !ok {
				return nil, "", fmt.Errorf("public key %d is not ecdsa.PublicKey", i)
			}
			retKeys.Keys[tlog.LogID] = cosign.TransparencyLogPubKey{
				PubKey: pkecdsa,
				Status: tuf.Active,
			}
			rekorURL = tlog.BaseURL.String()
		}
		return retKeys, rekorURL, nil
	}
	return nil, "", fmt.Errorf("trustRootRef %s not found", trustRootRef)
}

// splitPEMCertificateChain returns a list of leaf (non-CA) certificates, a certificate pool for
// intermediate CA certificates, and a certificate pool for root CA certificates
func splitPEMCertificateChain(pem []byte) (leaves, intermediates, roots []*x509.Certificate, err error) {
	certs, err := cryptoutils.UnmarshalCertificatesFromPEM(pem)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, cert := range certs {
		if !cert.IsCA {
			leaves = append(leaves, cert)
		} else {
			// root certificates are self-signed
			if bytes.Equal(cert.RawSubject, cert.RawIssuer) {
				roots = append(roots, cert)
			} else {
				intermediates = append(intermediates, cert)
			}
		}
	}

	return leaves, intermediates, roots, nil
}
