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

package cosign

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"os"
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/transparency-dev/merkle/proof"
	"github.com/transparency-dev/merkle/rfc6962"

	"github.com/sigstore/cosign/v2/pkg/cosign/bundle"
	"github.com/sigstore/cosign/v2/pkg/cosign/env"
	"github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/rekor/pkg/types"
	hashedrekord_v001 "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	"github.com/sigstore/rekor/pkg/types/intoto"
	intoto_v001 "github.com/sigstore/rekor/pkg/types/intoto/v0.0.1"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/tuf"
)

// This is the rekor transparency log public key target name
var rekorTargetStr = `rekor.pub`

// TransparencyLogPubKey contains the ECDSA verification key and the current status
// of the key according to TUF metadata, whether it's active or expired.
type TransparencyLogPubKey struct {
	PubKey crypto.PublicKey
	Status tuf.StatusKind
}

// This is a map of TransparencyLog public keys indexed by log ID that's used
// in verification.
type TrustedTransparencyLogPubKeys struct {
	// A map of keys indexed by log ID
	Keys map[string]TransparencyLogPubKey
}

const treeIDHexStringLen = 16
const uuidHexStringLen = 64
const entryIDHexStringLen = treeIDHexStringLen + uuidHexStringLen

// GetTransparencyLogID generates a SHA256 hash of a DER-encoded public key.
// (see RFC 6962 S3.2)
// In CT V1 the log id is a hash of the public key.
func GetTransparencyLogID(pub crypto.PublicKey) (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(pubBytes)
	return hex.EncodeToString(digest[:]), nil
}

func intotoEntry(ctx context.Context, signature, pubKey []byte) (models.ProposedEntry, error) {
	var pubKeyBytes [][]byte

	if len(pubKey) == 0 {
		return nil, errors.New("none of the Rekor public keys have been found")
	}

	pubKeyBytes = append(pubKeyBytes, pubKey)

	return types.NewProposedEntry(ctx, intoto.KIND, intoto_v001.APIVERSION, types.ArtifactProperties{
		ArtifactBytes:  signature,
		PublicKeyBytes: pubKeyBytes,
	})
}

// GetRekorPubs retrieves trusted Rekor public keys from the embedded or cached
// TUF root. If expired, makes a network call to retrieve the updated targets.
// There are two Env variable that can be used to override this behaviour:
// SIGSTORE_REKOR_PUBLIC_KEY - If specified, location of the file that contains
// the Rekor Public Key on local filesystem
func GetRekorPubs(ctx context.Context) (*TrustedTransparencyLogPubKeys, error) {
	publicKeys := NewTrustedTransparencyLogPubKeys()
	altRekorPub := env.Getenv(env.VariableSigstoreRekorPublicKey)

	if altRekorPub != "" {
		raw, err := os.ReadFile(altRekorPub)
		if err != nil {
			return nil, fmt.Errorf("error reading alternate Rekor public key file: %w", err)
		}
		if err := publicKeys.AddTransparencyLogPubKey(raw, tuf.Active); err != nil {
			return nil, fmt.Errorf("AddRekorPubKey: %w", err)
		}
	} else {
		tufClient, err := tuf.NewFromEnv(ctx)
		if err != nil {
			return nil, err
		}
		targets, err := tufClient.GetTargetsByMeta(tuf.Rekor, []string{rekorTargetStr})
		if err != nil {
			return nil, err
		}
		for _, t := range targets {
			if err := publicKeys.AddTransparencyLogPubKey(t.Target, t.Status); err != nil {
				return nil, fmt.Errorf("AddRekorPubKey: %w", err)
			}
		}
	}

	if len(publicKeys.Keys) == 0 {
		return nil, errors.New("none of the Rekor public keys have been found")
	}

	return &publicKeys, nil
}

// rekorPubsFromClient returns a RekorPubKey keyed by the log ID from the Rekor client.
// NOTE: This **must not** be used in the verification path, but may be used in the
// sign path to validate return responses are consistent from Rekor.
func rekorPubsFromClient(rekorClient *client.Rekor) (*TrustedTransparencyLogPubKeys, error) {
	publicKeys := NewTrustedTransparencyLogPubKeys()
	pubOK, err := rekorClient.Pubkey.GetPublicKey(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch rekor public key from rekor: %w", err)
	}
	if err := publicKeys.AddTransparencyLogPubKey([]byte(pubOK.Payload), tuf.Active); err != nil {
		return nil, fmt.Errorf("constructRekorPubKey: %w", err)
	}
	return &publicKeys, nil
}

// TLogUpload will upload the signature, public key and payload to the transparency log.
func TLogUpload(ctx context.Context, rekorClient *client.Rekor, signature []byte, sha256CheckSum hash.Hash, pemBytes []byte) (*models.LogEntryAnon, error) {
	re := rekorEntry(sha256CheckSum, signature, pemBytes)
	returnVal := models.Hashedrekord{
		APIVersion: swag.String(re.APIVersion()),
		Spec:       re.HashedRekordObj,
	}
	return doUpload(ctx, rekorClient, &returnVal)
}

// TLogUploadInTotoAttestation will upload and in-toto entry for the signature and public key to the transparency log.
func TLogUploadInTotoAttestation(ctx context.Context, rekorClient *client.Rekor, signature, pemBytes []byte) (*models.LogEntryAnon, error) {
	e, err := intotoEntry(ctx, signature, pemBytes)
	if err != nil {
		return nil, err
	}

	return doUpload(ctx, rekorClient, e)
}

func doUpload(ctx context.Context, rekorClient *client.Rekor, pe models.ProposedEntry) (*models.LogEntryAnon, error) {
	params := entries.NewCreateLogEntryParamsWithContext(ctx)
	params.SetProposedEntry(pe)
	resp, err := rekorClient.Entries.CreateLogEntry(params)
	if err != nil {
		// If the entry already exists, we get a specific error.
		// Here, we display the proof and succeed.
		var existsErr *entries.CreateLogEntryConflict
		if errors.As(err, &existsErr) {
			fmt.Println("Signature already exists. Displaying proof")
			uriSplit := strings.Split(existsErr.Location.String(), "/")
			uuid := uriSplit[len(uriSplit)-1]
			e, err := GetTlogEntry(ctx, rekorClient, uuid)
			if err != nil {
				return nil, err
			}
			rekorPubsFromAPI, err := rekorPubsFromClient(rekorClient)
			if err != nil {
				return nil, err
			}
			return e, VerifyTLogEntryOffline(e, rekorPubsFromAPI)
		}
		return nil, err
	}
	// UUID is at the end of location
	for _, p := range resp.Payload {
		return &p, nil
	}
	return nil, errors.New("bad response from server")
}

func rekorEntry(sha256CheckSum hash.Hash, signature, pubKey []byte) hashedrekord_v001.V001Entry {
	// TODO: Signatures created on a digest using a hash algorithm other than SHA256 will fail
	// upload right now. Plumb information on the hash algorithm used when signing from the
	// SignerVerifier to use for the HashedRekordObj.Data.Hash.Algorithm.
	return hashedrekord_v001.V001Entry{
		HashedRekordObj: models.HashedrekordV001Schema{
			Data: &models.HashedrekordV001SchemaData{
				Hash: &models.HashedrekordV001SchemaDataHash{
					Algorithm: swag.String(models.HashedrekordV001SchemaDataHashAlgorithmSha256),
					Value:     swag.String(hex.EncodeToString(sha256CheckSum.Sum(nil))),
				},
			},
			Signature: &models.HashedrekordV001SchemaSignature{
				Content: strfmt.Base64(signature),
				PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
					Content: strfmt.Base64(pubKey),
				},
			},
		},
	}
}

func ComputeLeafHash(e *models.LogEntryAnon) ([]byte, error) {
	entryBytes, err := base64.StdEncoding.DecodeString(e.Body.(string))
	if err != nil {
		return nil, err
	}
	return rfc6962.DefaultHasher.HashLeaf(entryBytes), nil
}

func getUUID(entryUUID string) (string, error) {
	switch len(entryUUID) {
	case uuidHexStringLen:
		if _, err := hex.DecodeString(entryUUID); err != nil {
			return "", fmt.Errorf("uuid %v is not a valid hex string: %w", entryUUID, err)
		}
		return entryUUID, nil
	case entryIDHexStringLen:
		uid := entryUUID[len(entryUUID)-uuidHexStringLen:]
		return getUUID(uid)
	default:
		return "", fmt.Errorf("invalid ID len %v for %v", len(entryUUID), entryUUID)
	}
}

func getTreeUUID(entryUUID string) (string, error) {
	switch len(entryUUID) {
	case uuidHexStringLen:
		// No Tree ID provided
		return "", nil
	case entryIDHexStringLen:
		tid := entryUUID[:treeIDHexStringLen]
		return getTreeUUID(tid)
	case treeIDHexStringLen:
		// Check that it's a valid int64 in hex (base 16)
		i, err := strconv.ParseInt(entryUUID, 16, 64)
		if err != nil {
			return "", fmt.Errorf("could not convert treeID %v to int64: %w", entryUUID, err)
		}
		// Check for invalid TreeID values
		if i == 0 {
			return "", fmt.Errorf("0 is not a valid TreeID")
		}
		return entryUUID, nil
	default:
		return "", fmt.Errorf("invalid ID len %v for %v", len(entryUUID), entryUUID)
	}
}

// Validates UUID and also TreeID if present.
func isExpectedResponseUUID(requestEntryUUID string, responseEntryUUID string, treeid string) error {
	// Comparare UUIDs
	requestUUID, err := getUUID(requestEntryUUID)
	if err != nil {
		return err
	}
	responseUUID, err := getUUID(responseEntryUUID)
	if err != nil {
		return err
	}
	if requestUUID != responseUUID {
		return fmt.Errorf("expected EntryUUID %s got UUID %s", requestEntryUUID, responseEntryUUID)
	}
	// Compare tree ID if it is in the request.
	requestTreeID, err := getTreeUUID(requestEntryUUID)
	if err != nil {
		return err
	}
	if requestTreeID != "" {
		tid, err := getTreeUUID(treeid)
		if err != nil {
			return err
		}
		if requestTreeID != tid {
			return fmt.Errorf("expected EntryUUID %s got UUID %s from Tree %s", requestEntryUUID, responseEntryUUID, treeid)
		}
	}
	return nil
}

func verifyUUID(entryUUID string, e models.LogEntryAnon) error {
	// Verify and get the UUID.
	uid, err := getUUID(entryUUID)
	if err != nil {
		return err
	}
	uuid, _ := hex.DecodeString(uid)

	// Verify leaf hash matches hash of the entry body.
	computedLeafHash, err := ComputeLeafHash(&e)
	if err != nil {
		return err
	}
	if !bytes.Equal(computedLeafHash, uuid) {
		return fmt.Errorf("computed leaf hash did not match UUID")
	}
	return nil
}

func GetTlogEntry(ctx context.Context, rekorClient *client.Rekor, entryUUID string) (*models.LogEntryAnon, error) {
	params := entries.NewGetLogEntryByUUIDParamsWithContext(ctx)
	params.SetEntryUUID(entryUUID)
	resp, err := rekorClient.Entries.GetLogEntryByUUID(params)
	if err != nil {
		return nil, err
	}
	for k, e := range resp.Payload {
		// Validate that request EntryUUID matches the response UUID and response Tree ID
		if err := isExpectedResponseUUID(entryUUID, k, *e.LogID); err != nil {
			return nil, fmt.Errorf("unexpected entry returned from rekor server: %w", err)
		}
		// Check that body hash matches UUID
		if err := verifyUUID(k, e); err != nil {
			return nil, err
		}
		return &e, nil
	}
	return nil, errors.New("empty response")
}

func proposedEntry(b64Sig string, payload, pubKey []byte) ([]models.ProposedEntry, error) {
	var proposedEntry []models.ProposedEntry
	signature, err := base64.StdEncoding.DecodeString(b64Sig)
	if err != nil {
		return nil, fmt.Errorf("decoding base64 signature: %w", err)
	}

	// The fact that there's no signature (or empty rather), implies
	// that this is an Attestation that we're verifying.
	if len(signature) == 0 {
		e, err := intotoEntry(context.Background(), payload, pubKey)
		if err != nil {
			return nil, err
		}
		proposedEntry = []models.ProposedEntry{e}
	} else {
		sha256CheckSum := sha256.New()
		if _, err := sha256CheckSum.Write(payload); err != nil {
			return nil, err
		}
		re := rekorEntry(sha256CheckSum, signature, pubKey)
		entry := &models.Hashedrekord{
			APIVersion: swag.String(re.APIVersion()),
			Spec:       re.HashedRekordObj,
		}
		proposedEntry = []models.ProposedEntry{entry}
	}
	return proposedEntry, nil
}

func FindTlogEntry(ctx context.Context, rekorClient *client.Rekor,
	b64Sig string, payload, pubKey []byte) ([]models.LogEntryAnon, error) {
	searchParams := entries.NewSearchLogQueryParamsWithContext(ctx)
	searchLogQuery := models.SearchLogQuery{}
	proposedEntry, err := proposedEntry(b64Sig, payload, pubKey)
	if err != nil {
		return nil, err
	}

	searchLogQuery.SetEntries(proposedEntry)

	searchParams.SetEntry(&searchLogQuery)
	resp, err := rekorClient.Entries.SearchLogQuery(searchParams)
	if err != nil {
		return nil, fmt.Errorf("searching log query: %w", err)
	}
	if len(resp.Payload) == 0 {
		return nil, errors.New("signature not found in transparency log")
	}

	// This may accumulate multiple entries on multiple tree IDs.
	results := make([]models.LogEntryAnon, 0)
	for _, logEntry := range resp.GetPayload() {
		for k, e := range logEntry {
			// Check body hash matches uuid
			if err := verifyUUID(k, e); err != nil {
				continue
			}
			results = append(results, e)
		}
	}

	return results, nil
}

// VerifyTLogEntryOffline verifies a TLog entry against a map of trusted rekorPubKeys indexed
// by log id.
func VerifyTLogEntryOffline(e *models.LogEntryAnon, rekorPubKeys *TrustedTransparencyLogPubKeys) error {
	if e.Verification == nil || e.Verification.InclusionProof == nil {
		return errors.New("inclusion proof not provided")
	}

	if rekorPubKeys == nil || rekorPubKeys.Keys == nil {
		return errors.New("no trusted rekor public keys provided")
	}
	// Make sure all the rekorPubKeys are ecsda.PublicKeys
	for k, v := range rekorPubKeys.Keys {
		if _, ok := v.PubKey.(*ecdsa.PublicKey); !ok {
			return fmt.Errorf("rekor Public key for LogID %s is not type ecdsa.PublicKey", k)
		}
	}

	hashes := [][]byte{}
	for _, h := range e.Verification.InclusionProof.Hashes {
		hb, _ := hex.DecodeString(h)
		hashes = append(hashes, hb)
	}

	rootHash, _ := hex.DecodeString(*e.Verification.InclusionProof.RootHash)
	entryBytes, err := base64.StdEncoding.DecodeString(e.Body.(string))
	if err != nil {
		return err
	}
	leafHash := rfc6962.DefaultHasher.HashLeaf(entryBytes)

	// Verify the inclusion proof.
	if err := proof.VerifyInclusion(rfc6962.DefaultHasher, uint64(*e.Verification.InclusionProof.LogIndex), uint64(*e.Verification.InclusionProof.TreeSize),
		leafHash, hashes, rootHash); err != nil {
		return fmt.Errorf("verifying inclusion proof: %w", err)
	}

	// Verify rekor's signature over the SET.
	payload := bundle.RekorPayload{
		Body:           e.Body,
		IntegratedTime: *e.IntegratedTime,
		LogIndex:       *e.LogIndex,
		LogID:          *e.LogID,
	}

	pubKey, ok := rekorPubKeys.Keys[payload.LogID]
	if !ok {
		return errors.New("rekor log public key not found for payload")
	}
	err = VerifySET(payload, []byte(e.Verification.SignedEntryTimestamp), pubKey.PubKey.(*ecdsa.PublicKey))
	if err != nil {
		return fmt.Errorf("verifying signedEntryTimestamp: %w", err)
	}
	if pubKey.Status != tuf.Active {
		fmt.Fprintf(os.Stderr, "**Info** Successfully verified Rekor entry using an expired verification key\n")
	}
	return nil
}

func NewTrustedTransparencyLogPubKeys() TrustedTransparencyLogPubKeys {
	return TrustedTransparencyLogPubKeys{Keys: make(map[string]TransparencyLogPubKey, 0)}
}

// constructRekorPubkey returns a log ID and RekorPubKey from a given
// byte-array representing the PEM-encoded Rekor key and a status.
func (t *TrustedTransparencyLogPubKeys) AddTransparencyLogPubKey(pemBytes []byte, status tuf.StatusKind) error {
	pubKey, err := cryptoutils.UnmarshalPEMToPublicKey(pemBytes)
	if err != nil {
		return err
	}
	keyID, err := GetTransparencyLogID(pubKey)
	if err != nil {
		return err
	}
	t.Keys[keyID] = TransparencyLogPubKey{PubKey: pubKey, Status: status}
	return nil
}
