package verify

import (
	"encoding/json"
	"strings"
	"time"

	cjson "github.com/tent/canonical-json-go"
	"github.com/theupdateframework/go-tuf/data"
)

type signedMeta struct {
	Type    string    `json:"_type"`
	Expires time.Time `json:"expires"`
	Version int       `json:"version"`
}

func (db *DB) Verify(s *data.Signed, role string, minVersion int) error {
	if err := db.VerifySignatures(s, role); err != nil {
		return err
	}

	sm := &signedMeta{}
	if err := json.Unmarshal(s.Signed, sm); err != nil {
		return err
	}

	if isTopLevelRole(role) {
		// Top-level roles can only sign metadata of the same type (e.g. snapshot
		// metadata must be signed by the snapshot role).
		if strings.ToLower(sm.Type) != strings.ToLower(role) {
			return ErrWrongMetaType
		}
	} else {
		// Delegated (non-top-level) roles may only sign targets metadata.
		if strings.ToLower(sm.Type) != "targets" {
			return ErrWrongMetaType
		}
	}

	if IsExpired(sm.Expires) {
		return ErrExpired{sm.Expires}
	}

	if sm.Version < minVersion {
		return ErrLowVersion{sm.Version, minVersion}
	}

	return nil
}

var IsExpired = func(t time.Time) bool {
	return t.Sub(time.Now()) <= 0
}

func (db *DB) VerifySignatures(s *data.Signed, role string) error {
	if len(s.Signatures) == 0 {
		return ErrNoSignatures
	}

	roleData := db.GetRole(role)
	if roleData == nil {
		return ErrUnknownRole{role}
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(s.Signed, &decoded); err != nil {
		return err
	}
	msg, err := cjson.Marshal(decoded)
	if err != nil {
		return err
	}

	// Verify that a threshold of keys signed the data. Since keys can have
	// multiple key ids, we need to protect against multiple attached
	// signatures that just differ on the key id.
	seen := make(map[string]struct{})
	valid := 0
	for _, sig := range s.Signatures {
		if !roleData.ValidKey(sig.KeyID) {
			continue
		}
		key := db.GetKey(sig.KeyID)
		if key == nil {
			continue
		}

		if err := Verifiers[key.Type].Verify(key.Value.Public, msg, sig.Signature); err != nil {
			return err
		}

		// Only consider this key valid if we haven't seen any of it's
		// key ids before.
		if _, ok := seen[sig.KeyID]; !ok {
			for _, id := range key.IDs() {
				seen[id] = struct{}{}
			}

			valid++
		}
	}
	if valid < roleData.Threshold {
		return ErrRoleThreshold{roleData.Threshold, valid}
	}

	return nil
}

func (db *DB) Unmarshal(b []byte, v interface{}, role string, minVersion int) error {
	s := &data.Signed{}
	if err := json.Unmarshal(b, s); err != nil {
		return err
	}
	if err := db.Verify(s, role, minVersion); err != nil {
		return err
	}
	return json.Unmarshal(s.Signed, v)
}

func (db *DB) UnmarshalTrusted(b []byte, v interface{}, role string) error {
	s := &data.Signed{}
	if err := json.Unmarshal(b, s); err != nil {
		return err
	}
	if err := db.VerifySignatures(s, role); err != nil {
		return err
	}
	return json.Unmarshal(s.Signed, v)
}
