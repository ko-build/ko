package sign

import (
	"crypto/rand"
	"sync"

	"github.com/theupdateframework/go-tuf/data"
	"golang.org/x/crypto/ed25519"
)

type PrivateKey struct {
	Type       string          `json:"keytype"`
	Scheme     string          `json:"scheme,omitempty"`
	Algorithms []string        `json:"keyid_hash_algorithms,omitempty"`
	Value      PrivateKeyValue `json:"keyval"`
}

type PrivateKeyValue struct {
	Public  data.HexBytes `json:"public"`
	Private data.HexBytes `json:"private"`
}

func (k *PrivateKey) PublicData() *data.Key {
	return &data.Key{
		Type:       k.Type,
		Scheme:     k.Scheme,
		Algorithms: k.Algorithms,
		Value:      data.KeyValue{Public: k.Value.Public},
	}
}

func (k *PrivateKey) Signer() Signer {
	return &ed25519Signer{
		PrivateKey:    ed25519.PrivateKey(k.Value.Private),
		keyType:       k.Type,
		keyScheme:     k.Scheme,
		keyAlgorithms: k.Algorithms,
	}
}

func GenerateEd25519Key() (*PrivateKey, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &PrivateKey{
		Type:       data.KeyTypeEd25519,
		Scheme:     data.KeySchemeEd25519,
		Algorithms: data.KeyAlgorithms,
		Value: PrivateKeyValue{
			Public:  data.HexBytes(public),
			Private: data.HexBytes(private),
		},
	}, nil
}

type ed25519Signer struct {
	ed25519.PrivateKey

	keyType       string
	keyScheme     string
	keyAlgorithms []string
	ids           []string
	idOnce        sync.Once
}

var _ Signer = &ed25519Signer{}

func (s *ed25519Signer) IDs() []string {
	s.idOnce.Do(func() { s.ids = s.publicData().IDs() })
	return s.ids
}

func (s *ed25519Signer) ContainsID(id string) bool {
	for _, keyid := range s.IDs() {
		if id == keyid {
			return true
		}
	}
	return false
}

func (s *ed25519Signer) publicData() *data.Key {
	return &data.Key{
		Type:       s.keyType,
		Scheme:     s.keyScheme,
		Algorithms: s.keyAlgorithms,
		Value:      data.KeyValue{Public: []byte(s.PrivateKey.Public().(ed25519.PublicKey))},
	}
}

func (s *ed25519Signer) Type() string {
	return s.keyType
}

func (s *ed25519Signer) Scheme() string {
	return s.keyScheme
}
