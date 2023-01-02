// Copyright 2022 The Sigstore Authors.
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

package v1alpha1

import (
	"context"
	"fmt"

	"github.com/sigstore/policy-controller/pkg/apis/policy/v1beta1"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/ptr"
)

var _ apis.Convertible = (*ClusterImagePolicy)(nil)

// ConvertTo implements api.Convertible
func (c *ClusterImagePolicy) ConvertTo(ctx context.Context, obj apis.Convertible) error {
	switch sink := obj.(type) {
	case *v1beta1.ClusterImagePolicy:
		sink.ObjectMeta = c.ObjectMeta
		return c.Spec.ConvertTo(ctx, &sink.Spec)
	default:
		return fmt.Errorf("unknown version, got: %T", sink)
	}
}

// ConvertFrom implements api.Convertible
func (c *ClusterImagePolicy) ConvertFrom(ctx context.Context, obj apis.Convertible) error {
	switch source := obj.(type) {
	case *v1beta1.ClusterImagePolicy:
		c.ObjectMeta = source.ObjectMeta
		return c.Spec.ConvertFrom(ctx, &source.Spec)
	default:
		return fmt.Errorf("unknown version, got: %T", c)
	}
}

func (spec *ClusterImagePolicySpec) ConvertTo(ctx context.Context, sink *v1beta1.ClusterImagePolicySpec) error {
	for _, image := range spec.Images {
		sink.Images = append(sink.Images, v1beta1.ImagePattern{Glob: image.Glob})
	}
	for _, authority := range spec.Authorities {
		v1beta1Authority := v1beta1.Authority{}
		err := authority.ConvertTo(ctx, &v1beta1Authority)
		if err != nil {
			return err
		}
		sink.Authorities = append(sink.Authorities, v1beta1Authority)
	}
	for _, m := range spec.Match {
		v1beta1Match := v1beta1.MatchResource{}
		err := m.ConvertTo(ctx, &v1beta1Match)
		if err != nil {
			return err
		}
		sink.Match = append(sink.Match, v1beta1Match)
	}
	if spec.Policy != nil {
		sink.Policy = &v1beta1.Policy{}
		spec.Policy.ConvertTo(ctx, sink.Policy)
	}
	sink.Mode = spec.Mode
	return nil
}

func (matchResource *MatchResource) ConvertTo(ctx context.Context, sink *v1beta1.MatchResource) error {
	sink.GroupVersionResource = *matchResource.GroupVersionResource.DeepCopy()
	if matchResource.ResourceSelector != nil {
		sink.ResourceSelector = matchResource.ResourceSelector.DeepCopy()
	}

	return nil
}

func (authority *Authority) ConvertTo(ctx context.Context, sink *v1beta1.Authority) error {
	sink.Name = authority.Name
	if authority.CTLog != nil && authority.CTLog.URL != nil {
		sink.CTLog = &v1beta1.TLog{
			URL:          authority.CTLog.URL.DeepCopy(),
			TrustRootRef: authority.CTLog.TrustRootRef,
		}
	}
	if authority.RFC3161Timestamp != nil && authority.RFC3161Timestamp.TrustRootRef != "" {
		sink.RFC3161Timestamp = &v1beta1.RFC3161Timestamp{}
		sink.RFC3161Timestamp.TrustRootRef = authority.RFC3161Timestamp.TrustRootRef
	}
	for _, source := range authority.Sources {
		v1beta1Source := v1beta1.Source{}
		v1beta1Source.OCI = source.OCI
		for _, sps := range source.SignaturePullSecrets {
			v1beta1Source.SignaturePullSecrets = append(v1beta1Source.SignaturePullSecrets, v1.LocalObjectReference{Name: sps.Name})
		}
		sink.Sources = append(sink.Sources, v1beta1Source)
	}
	for _, att := range authority.Attestations {
		v1beta1Att := v1beta1.Attestation{}
		v1beta1Att.Name = att.Name
		v1beta1Att.PredicateType = att.PredicateType
		if att.Policy != nil {
			v1beta1Att.Policy = &v1beta1.Policy{}
			att.Policy.ConvertTo(ctx, v1beta1Att.Policy)
		}
		sink.Attestations = append(sink.Attestations, v1beta1Att)
	}
	if authority.Key != nil {
		sink.Key = &v1beta1.KeyRef{}
		authority.Key.ConvertTo(ctx, sink.Key)
	}
	if authority.Keyless != nil {
		sink.Keyless = &v1beta1.KeylessRef{
			URL:          authority.Keyless.URL.DeepCopy(),
			TrustRootRef: authority.Keyless.TrustRootRef,
		}
		for _, id := range authority.Keyless.Identities {
			sink.Keyless.Identities = append(sink.Keyless.Identities, v1beta1.Identity{Issuer: id.Issuer, Subject: id.Subject, IssuerRegExp: id.IssuerRegExp, SubjectRegExp: id.SubjectRegExp})
		}
		if authority.Keyless.CACert != nil {
			sink.Keyless.CACert = &v1beta1.KeyRef{}
			authority.Keyless.CACert.ConvertTo(ctx, sink.Keyless.CACert)
		}
	}
	if authority.Static != nil {
		sink.Static = &v1beta1.StaticRef{
			Action: authority.Static.Action,
		}
	}
	return nil
}

func (p *Policy) ConvertTo(ctx context.Context, sink *v1beta1.Policy) {
	sink.Type = p.Type
	sink.Data = p.Data
	if p.URL != nil {
		sink.URL = p.URL.DeepCopy()
	}
	if p.ConfigMapRef != nil {
		sink.ConfigMapRef = &v1beta1.ConfigMapReference{
			Name:      p.ConfigMapRef.Name,
			Namespace: p.ConfigMapRef.Namespace,
			Key:       p.ConfigMapRef.Key,
		}
	}
	if p.FetchConfigFile != nil {
		sink.FetchConfigFile = ptr.Bool(*p.FetchConfigFile)
	}
	if p.IncludeSpec != nil {
		sink.IncludeSpec = ptr.Bool(*p.IncludeSpec)
	}
	if p.IncludeObjectMeta != nil {
		sink.IncludeObjectMeta = ptr.Bool(*p.IncludeObjectMeta)
	}
	if p.IncludeTypeMeta != nil {
		sink.IncludeTypeMeta = ptr.Bool(*p.IncludeTypeMeta)
	}
}

func (p *Policy) ConvertFrom(ctx context.Context, source *v1beta1.Policy) {
	p.Type = source.Type
	p.Data = source.Data
	if source.URL != nil {
		p.URL = source.URL.DeepCopy()
	}
	if source.ConfigMapRef != nil {
		p.ConfigMapRef = &ConfigMapReference{
			Name:      source.ConfigMapRef.Name,
			Namespace: source.ConfigMapRef.Namespace,
			Key:       source.ConfigMapRef.Key,
		}
	}
	if source.FetchConfigFile != nil {
		p.FetchConfigFile = ptr.Bool(*source.FetchConfigFile)
	}
	if source.IncludeSpec != nil {
		p.IncludeSpec = ptr.Bool(*source.IncludeSpec)
	}
	if source.IncludeObjectMeta != nil {
		p.IncludeObjectMeta = ptr.Bool(*source.IncludeObjectMeta)
	}
	if source.IncludeTypeMeta != nil {
		p.IncludeTypeMeta = ptr.Bool(*source.IncludeTypeMeta)
	}
}

func (key *KeyRef) ConvertTo(ctx context.Context, sink *v1beta1.KeyRef) {
	sink.SecretRef = key.SecretRef.DeepCopy()
	sink.Data = key.Data
	sink.KMS = key.KMS
	sink.HashAlgorithm = key.HashAlgorithm
}

func (spec *ClusterImagePolicySpec) ConvertFrom(ctx context.Context, source *v1beta1.ClusterImagePolicySpec) error {
	for _, image := range source.Images {
		spec.Images = append(spec.Images, ImagePattern{Glob: image.Glob})
	}
	for i := range source.Authorities {
		authority := Authority{}
		err := authority.ConvertFrom(ctx, &source.Authorities[i])
		if err != nil {
			return err
		}
		spec.Authorities = append(spec.Authorities, authority)
	}
	for i := range source.Match {
		matchResource := MatchResource{}
		err := matchResource.ConvertFrom(ctx, &source.Match[i])
		if err != nil {
			return err
		}
		spec.Match = append(spec.Match, matchResource)
	}
	spec.Mode = source.Mode
	if source.Policy != nil {
		spec.Policy = &Policy{}
		spec.Policy.ConvertFrom(ctx, source.Policy)
	}
	return nil
}

func (authority *Authority) ConvertFrom(ctx context.Context, source *v1beta1.Authority) error {
	authority.Name = source.Name
	if source.CTLog != nil && source.CTLog.URL != nil {
		authority.CTLog = &TLog{
			URL:          source.CTLog.URL.DeepCopy(),
			TrustRootRef: source.CTLog.TrustRootRef,
		}
	}
	if source.RFC3161Timestamp != nil && source.RFC3161Timestamp.TrustRootRef != "" {
		authority.RFC3161Timestamp = &RFC3161Timestamp{}
		authority.RFC3161Timestamp.TrustRootRef = source.RFC3161Timestamp.TrustRootRef
	}
	for _, s := range source.Sources {
		src := Source{}
		src.OCI = s.OCI
		for _, sps := range s.SignaturePullSecrets {
			src.SignaturePullSecrets = append(src.SignaturePullSecrets, v1.LocalObjectReference{Name: sps.Name})
		}
		authority.Sources = append(authority.Sources, src)
	}
	for _, att := range source.Attestations {
		attestation := Attestation{}
		attestation.Name = att.Name
		attestation.PredicateType = att.PredicateType
		if att.Policy != nil {
			attestation.Policy = &Policy{}
			attestation.Policy.ConvertFrom(ctx, att.Policy)
		}
		authority.Attestations = append(authority.Attestations, attestation)
	}
	if source.Key != nil {
		authority.Key = &KeyRef{}
		authority.Key.ConvertFrom(ctx, source.Key)
	}
	if source.Keyless != nil {
		authority.Keyless = &KeylessRef{
			URL:          source.Keyless.URL.DeepCopy(),
			TrustRootRef: source.Keyless.TrustRootRef,
		}
		for _, id := range source.Keyless.Identities {
			authority.Keyless.Identities = append(authority.Keyless.Identities, Identity{Issuer: id.Issuer, Subject: id.Subject, IssuerRegExp: id.IssuerRegExp, SubjectRegExp: id.SubjectRegExp})
		}
		if source.Keyless.CACert != nil {
			authority.Keyless.CACert = &KeyRef{}
			authority.Keyless.CACert.ConvertFrom(ctx, source.Keyless.CACert)
		}
	}
	if source.Static != nil {
		authority.Static = &StaticRef{Action: source.Static.Action}
	}
	return nil
}

func (key *KeyRef) ConvertFrom(ctx context.Context, source *v1beta1.KeyRef) {
	key.SecretRef = source.SecretRef.DeepCopy()
	key.Data = source.Data
	key.KMS = source.KMS
	key.HashAlgorithm = source.HashAlgorithm
}

func (matchResource *MatchResource) ConvertFrom(ctx context.Context, source *v1beta1.MatchResource) error {
	matchResource.GroupVersionResource = *source.GroupVersionResource.DeepCopy()
	if source.ResourceSelector != nil {
		matchResource.ResourceSelector = source.ResourceSelector.DeepCopy()
	}
	return nil
}
