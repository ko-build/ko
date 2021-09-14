package sign

import (
	"crypto"
	"crypto/rand"

	cjson "github.com/tent/canonical-json-go"
	"github.com/theupdateframework/go-tuf/data"
)

type Signer interface {
	// IDs returns the TUF key ids
	IDs() []string

	// ContainsID returns if the signer contains the key id
	ContainsID(id string) bool

	// Type returns the TUF key type
	Type() string

	// Scheme returns the TUF key scheme
	Scheme() string

	// Signer is used to sign messages and provides access to the public key.
	// The signer is expected to do its own hashing, so the full message will be
	// provided as the message to Sign with a zero opts.HashFunc().
	crypto.Signer
}

func Sign(s *data.Signed, k Signer) error {
	ids := k.IDs()
	signatures := make([]data.Signature, 0, len(s.Signatures)+1)
	for _, sig := range s.Signatures {
		found := false
		for _, id := range ids {
			if sig.KeyID == id {
				found = true
				break
			}
		}
		if !found {
			signatures = append(signatures, sig)
		}
	}

	sig, err := k.Sign(rand.Reader, s.Signed, crypto.Hash(0))
	if err != nil {
		return err
	}

	s.Signatures = signatures
	for _, id := range ids {
		s.Signatures = append(s.Signatures, data.Signature{
			KeyID:     id,
			Signature: sig,
		})
	}

	return nil
}

func Marshal(v interface{}, keys ...Signer) (*data.Signed, error) {
	b, err := cjson.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := &data.Signed{Signed: b}
	for _, k := range keys {
		if err := Sign(s, k); err != nil {
			return nil, err
		}

	}
	return s, nil
}
