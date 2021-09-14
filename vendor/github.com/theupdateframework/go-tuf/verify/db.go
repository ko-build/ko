package verify

import (
	"github.com/theupdateframework/go-tuf/data"
)

type Role struct {
	KeyIDs    map[string]struct{}
	Threshold int
}

func (r *Role) ValidKey(id string) bool {
	_, ok := r.KeyIDs[id]
	return ok
}

type DB struct {
	roles map[string]*Role
	keys  map[string]*data.Key
}

func NewDB() *DB {
	return &DB{
		roles: make(map[string]*Role),
		keys:  make(map[string]*data.Key),
	}
}

func (db *DB) AddKey(id string, k *data.Key) error {
	v, ok := Verifiers[k.Type]
	if !ok {
		return nil
	}
	if !k.ContainsID(id) {
		return ErrWrongID{}
	}
	if !v.ValidKey(k.Value.Public) {
		return ErrInvalidKey
	}

	db.keys[id] = k

	return nil
}

var validRoles = map[string]struct{}{
	"root":      {},
	"targets":   {},
	"snapshot":  {},
	"timestamp": {},
}

func ValidRole(name string) bool {
	_, ok := validRoles[name]
	return ok
}

func (db *DB) AddRole(name string, r *data.Role) error {
	if !ValidRole(name) {
		return ErrInvalidRole
	}
	if r.Threshold < 1 {
		return ErrInvalidThreshold
	}

	role := &Role{
		KeyIDs:    make(map[string]struct{}),
		Threshold: r.Threshold,
	}
	for _, id := range r.KeyIDs {
		if len(id) != data.KeyIDLength {
			return ErrInvalidKeyID
		}
		role.KeyIDs[id] = struct{}{}
	}

	db.roles[name] = role
	return nil
}

func (db *DB) GetKey(id string) *data.Key {
	return db.keys[id]
}

func (db *DB) GetRole(name string) *Role {
	return db.roles[name]
}
