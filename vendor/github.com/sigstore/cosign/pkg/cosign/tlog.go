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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	_ "embed" // To enable the `go:embed` directive.

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/google/trillian/merkle/logverifier"
	"github.com/google/trillian/merkle/rfc6962/hasher"
	"github.com/pkg/errors"

	cremote "github.com/sigstore/cosign/pkg/cosign/remote"
	"github.com/sigstore/cosign/pkg/cosign/tuf"
	"github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/client/pubkey"
	"github.com/sigstore/rekor/pkg/generated/models"
	intoto_v001 "github.com/sigstore/rekor/pkg/types/intoto/v0.0.1"
	rekord_v001 "github.com/sigstore/rekor/pkg/types/rekord/v0.0.1"
)

// This is rekor's public key, via `curl -L rekor.sigstore.dev/api/ggcrv1/log/publicKey`
// rekor.pub should be updated whenever the Rekor public key is rotated & the bundle annotation should be up-versioned
//go:embed rekor.pub
var rekorPub string
var rekorTargetStr = `rekor.pub`

func GetRekorPub() string {
	ctx := context.Background() // TODO: pass in context?
	buf := tuf.ByteDestination{Buffer: &bytes.Buffer{}}
	err := tuf.GetTarget(ctx, rekorTargetStr, &buf)
	if err != nil {
		// The user may not have initialized the local root metadata. Log the error and use the embedded root.
		fmt.Fprintln(os.Stderr, "No TUF root installed, using embedded rekor key")
		return rekorPub
	}
	return buf.String()
}

// TLogUpload will upload the signature, public key and payload to the transparency log.
func TLogUpload(rekorClient *client.Rekor, signature, payload []byte, pemBytes []byte) (*models.LogEntryAnon, error) {
	re := rekorEntry(payload, signature, pemBytes)
	returnVal := models.Rekord{
		APIVersion: swag.String(re.APIVersion()),
		Spec:       re.RekordObj,
	}
	return doUpload(rekorClient, &returnVal)
}

// TLogUploadInTotoAttestation will upload and in-toto entry for the signature and public key to the transparency log.
func TLogUploadInTotoAttestation(rekorClient *client.Rekor, signature, pemBytes []byte) (*models.LogEntryAnon, error) {
	e := intotoEntry(signature, pemBytes)
	returnVal := models.Intoto{
		APIVersion: swag.String(e.APIVersion()),
		Spec:       e.IntotoObj,
	}
	return doUpload(rekorClient, &returnVal)
}

func doUpload(rekorClient *client.Rekor, pe models.ProposedEntry) (*models.LogEntryAnon, error) {
	params := entries.NewCreateLogEntryParams()
	params.SetProposedEntry(pe)
	resp, err := rekorClient.Entries.CreateLogEntry(params)
	if err != nil {
		// If the entry already exists, we get a specific error.
		// Here, we display the proof and succeed.
		if existsErr, ok := err.(*entries.CreateLogEntryConflict); ok {

			fmt.Println("Signature already exists. Displaying proof")
			uriSplit := strings.Split(existsErr.Location.String(), "/")
			uuid := uriSplit[len(uriSplit)-1]
			return verifyTLogEntry(rekorClient, uuid)
		}
		return nil, err
	}
	// UUID is at the end of location
	for _, p := range resp.Payload {
		return &p, nil
	}
	return nil, errors.New("bad response from server")
}

func intotoEntry(signature, pubKey []byte) intoto_v001.V001Entry {
	pub := strfmt.Base64(pubKey)
	return intoto_v001.V001Entry{
		IntotoObj: models.IntotoV001Schema{
			Content: &models.IntotoV001SchemaContent{
				Envelope: string(signature),
			},
			PublicKey: &pub,
		},
	}
}

func rekorEntry(payload, signature, pubKey []byte) rekord_v001.V001Entry {
	return rekord_v001.V001Entry{
		RekordObj: models.RekordV001Schema{
			Data: &models.RekordV001SchemaData{
				Content: strfmt.Base64(payload),
			},
			Signature: &models.RekordV001SchemaSignature{
				Content: strfmt.Base64(signature),
				Format:  models.RekordV001SchemaSignatureFormatX509,
				PublicKey: &models.RekordV001SchemaSignaturePublicKey{
					Content: strfmt.Base64(pubKey),
				},
			},
		},
	}
}

func getTlogEntry(rekorClient *client.Rekor, uuid string) (*models.LogEntryAnon, error) {
	params := entries.NewGetLogEntryByUUIDParams()
	params.SetEntryUUID(uuid)
	resp, err := rekorClient.Entries.GetLogEntryByUUID(params)
	if err != nil {
		return nil, err
	}
	for _, e := range resp.Payload {
		return &e, nil
	}
	return nil, errors.New("empty response")
}

func FindTlogEntry(rekorClient *client.Rekor, b64Sig string, payload, pubKey []byte) (uuid string, index int64, err error) {
	searchParams := entries.NewSearchLogQueryParams()
	searchLogQuery := models.SearchLogQuery{}
	signature, err := base64.StdEncoding.DecodeString(b64Sig)
	if err != nil {
		return "", 0, errors.Wrap(err, "decoding base64 signature")
	}
	re := rekorEntry(payload, signature, pubKey)
	entry := &models.Rekord{
		APIVersion: swag.String(re.APIVersion()),
		Spec:       re.RekordObj,
	}

	searchLogQuery.SetEntries([]models.ProposedEntry{entry})

	searchParams.SetEntry(&searchLogQuery)
	resp, err := rekorClient.Entries.SearchLogQuery(searchParams)
	if err != nil {
		return "", 0, errors.Wrap(err, "searching log query")
	}
	if len(resp.Payload) == 0 {
		return "", 0, errors.New("signature not found in transparency log")
	} else if len(resp.Payload) > 1 {
		return "", 0, errors.New("multiple entries returned; this should not happen")
	}
	logEntry := resp.Payload[0]
	if len(logEntry) != 1 {
		return "", 0, errors.New("UUID value can not be extracted")
	}

	for k := range logEntry {
		uuid = k
	}
	verifiedEntry, err := verifyTLogEntry(rekorClient, uuid)
	if err != nil {
		return "", 0, err
	}
	return uuid, *verifiedEntry.Verification.InclusionProof.LogIndex, nil
}

func verifyTLogEntry(rekorClient *client.Rekor, uuid string) (*models.LogEntryAnon, error) {
	params := entries.NewGetLogEntryByUUIDParams()
	params.EntryUUID = uuid

	lep, err := rekorClient.Entries.GetLogEntryByUUID(params)
	if err != nil {
		return nil, err
	}

	if len(lep.Payload) != 1 {
		return nil, errors.New("UUID value can not be extracted")
	}
	e := lep.Payload[params.EntryUUID]

	hashes := [][]byte{}
	for _, h := range e.Verification.InclusionProof.Hashes {
		hb, _ := hex.DecodeString(h)
		hashes = append(hashes, hb)
	}

	rootHash, _ := hex.DecodeString(*e.Verification.InclusionProof.RootHash)
	leafHash, _ := hex.DecodeString(params.EntryUUID)

	v := logverifier.New(hasher.DefaultHasher)
	if e.Verification == nil || e.Verification.InclusionProof == nil {
		return nil, fmt.Errorf("inclusion proof not provided")
	}
	if err := v.VerifyInclusionProof(*e.Verification.InclusionProof.LogIndex, *e.Verification.InclusionProof.TreeSize, hashes, rootHash, leafHash); err != nil {
		return nil, errors.Wrap(err, "verifying inclusion proof")
	}

	// Verify rekor's signature over the SET.
	resp, err := rekorClient.Pubkey.GetPublicKey(pubkey.NewGetPublicKeyParams())
	if err != nil {
		return nil, errors.Wrap(err, "rekor public key")
	}
	rekorPubKey, err := PemToECDSAKey([]byte(resp.Payload))
	if err != nil {
		return nil, errors.Wrap(err, "rekor public key pem to ecdsa")
	}

	payload := cremote.BundlePayload{
		Body:           e.Body,
		IntegratedTime: *e.IntegratedTime,
		LogIndex:       *e.LogIndex,
		LogID:          *e.LogID,
	}
	if err := VerifySET(payload, []byte(e.Verification.SignedEntryTimestamp), rekorPubKey); err != nil {
		return nil, errors.Wrap(err, "verifying signedEntryTimestamp")
	}

	return &e, nil
}
