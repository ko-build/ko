package tuf

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/encrypted"
	"github.com/theupdateframework/go-tuf/sign"
	"github.com/theupdateframework/go-tuf/util"
)

func MemoryStore(meta map[string]json.RawMessage, files map[string][]byte) LocalStore {
	if meta == nil {
		meta = make(map[string]json.RawMessage)
	}
	return &memoryStore{
		meta:       meta,
		stagedMeta: make(map[string]json.RawMessage),
		files:      files,
		signers:    make(map[string][]sign.Signer),
	}
}

type memoryStore struct {
	meta       map[string]json.RawMessage
	stagedMeta map[string]json.RawMessage
	files      map[string][]byte
	signers    map[string][]sign.Signer
}

func (m *memoryStore) GetMeta() (map[string]json.RawMessage, error) {
	meta := make(map[string]json.RawMessage, len(m.meta)+len(m.stagedMeta))
	for key, value := range m.meta {
		meta[key] = value
	}
	for key, value := range m.stagedMeta {
		meta[key] = value
	}
	return meta, nil
}

func (m *memoryStore) SetMeta(name string, meta json.RawMessage) error {
	m.stagedMeta[name] = meta
	return nil
}

func (m *memoryStore) WalkStagedTargets(paths []string, targetsFn TargetsWalkFunc) error {
	if len(paths) == 0 {
		for path, data := range m.files {
			if err := targetsFn(path, bytes.NewReader(data)); err != nil {
				return err
			}
		}
		return nil
	}

	for _, path := range paths {
		data, ok := m.files[path]
		if !ok {
			return ErrFileNotFound{path}
		}
		if err := targetsFn(path, bytes.NewReader(data)); err != nil {
			return err
		}
	}
	return nil
}

func (m *memoryStore) Commit(consistentSnapshot bool, versions map[string]int, hashes map[string]data.Hashes) error {
	for name, meta := range m.stagedMeta {
		paths := computeMetadataPaths(consistentSnapshot, name, versions)
		for _, path := range paths {
			m.meta[path] = meta
		}
	}
	return nil
}

func (m *memoryStore) GetSigningKeys(role string) ([]sign.Signer, error) {
	return m.signers[role], nil
}

func (m *memoryStore) SavePrivateKey(role string, key *sign.PrivateKey) error {
	m.signers[role] = append(m.signers[role], key.Signer())
	return nil
}

func (m *memoryStore) Clean() error {
	return nil
}

type persistedKeys struct {
	Encrypted bool            `json:"encrypted"`
	Data      json.RawMessage `json:"data"`
}

func FileSystemStore(dir string, p util.PassphraseFunc) LocalStore {
	return &fileSystemStore{
		dir:            dir,
		passphraseFunc: p,
		signers:        make(map[string][]sign.Signer),
	}
}

type fileSystemStore struct {
	dir            string
	passphraseFunc util.PassphraseFunc

	// signers is a cache of persisted keys to avoid decrypting multiple times
	signers map[string][]sign.Signer
}

func (f *fileSystemStore) repoDir() string {
	return filepath.Join(f.dir, "repository")
}

func (f *fileSystemStore) stagedDir() string {
	return filepath.Join(f.dir, "staged")
}

func (f *fileSystemStore) GetMeta() (map[string]json.RawMessage, error) {
	meta := make(map[string]json.RawMessage)
	var err error
	notExists := func(path string) bool {
		_, err := os.Stat(path)
		return os.IsNotExist(err)
	}
	for _, name := range topLevelManifests {
		path := filepath.Join(f.stagedDir(), name)
		if notExists(path) {
			path = filepath.Join(f.repoDir(), name)
			if notExists(path) {
				continue
			}
		}
		meta[name], err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	return meta, nil
}

func (f *fileSystemStore) SetMeta(name string, meta json.RawMessage) error {
	if err := f.createDirs(); err != nil {
		return err
	}
	if err := util.AtomicallyWriteFile(filepath.Join(f.stagedDir(), name), meta, 0644); err != nil {
		return err
	}
	return nil
}

func (f *fileSystemStore) createDirs() error {
	for _, dir := range []string{"keys", "repository", "staged/targets"} {
		if err := os.MkdirAll(filepath.Join(f.dir, dir), 0755); err != nil {
			return err
		}
	}
	return nil
}

func (f *fileSystemStore) WalkStagedTargets(paths []string, targetsFn TargetsWalkFunc) error {
	if len(paths) == 0 {
		walkFunc := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !info.Mode().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(filepath.Join(f.stagedDir(), "targets"), path)
			if err != nil {
				return err
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			return targetsFn(rel, file)
		}
		return filepath.Walk(filepath.Join(f.stagedDir(), "targets"), walkFunc)
	}

	// check all the files exist before processing any files
	for _, path := range paths {
		realPath := filepath.Join(f.stagedDir(), "targets", path)
		if _, err := os.Stat(realPath); err != nil {
			if os.IsNotExist(err) {
				return ErrFileNotFound{realPath}
			}
			return err
		}
	}

	for _, path := range paths {
		realPath := filepath.Join(f.stagedDir(), "targets", path)
		file, err := os.Open(realPath)
		if err != nil {
			if os.IsNotExist(err) {
				return ErrFileNotFound{realPath}
			}
			return err
		}
		err = targetsFn(path, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *fileSystemStore) createRepoFile(path string) (*os.File, error) {
	dst := filepath.Join(f.repoDir(), path)
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return nil, err
	}
	return os.Create(dst)
}

func (f *fileSystemStore) Commit(consistentSnapshot bool, versions map[string]int, hashes map[string]data.Hashes) error {
	isTarget := func(path string) bool {
		return strings.HasPrefix(path, "targets/")
	}
	copyToRepo := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(f.stagedDir(), path)
		if err != nil {
			return err
		}

		var paths []string
		if isTarget(rel) {
			paths = computeTargetPaths(consistentSnapshot, rel, hashes)
		} else {
			paths = computeMetadataPaths(consistentSnapshot, rel, versions)
		}
		var files []io.Writer
		for _, path := range paths {
			file, err := f.createRepoFile(path)
			if err != nil {
				return err
			}
			defer file.Close()
			files = append(files, file)
		}
		staged, err := os.Open(path)
		if err != nil {
			return err
		}
		defer staged.Close()
		if _, err = io.Copy(io.MultiWriter(files...), staged); err != nil {
			return err
		}
		return nil
	}
	needsRemoval := func(path string) bool {
		if consistentSnapshot {
			// strip out the hash
			name := strings.SplitN(filepath.Base(path), ".", 2)
			if len(name) != 2 || name[1] == "" {
				return false
			}
			path = filepath.Join(filepath.Dir(path), name[1])
		}
		_, ok := hashes[path]
		return !ok
	}
	removeFile := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(f.repoDir(), path)
		if err != nil {
			return err
		}
		if !info.IsDir() && isTarget(rel) && needsRemoval(rel) {
			if err := os.Remove(path); err != nil {
				// TODO: log / handle error
			}
			// TODO: remove empty directory
		}
		return nil
	}
	if err := filepath.Walk(f.stagedDir(), copyToRepo); err != nil {
		return err
	}
	if err := filepath.Walk(f.repoDir(), removeFile); err != nil {
		return err
	}
	return f.Clean()
}

func (f *fileSystemStore) GetSigningKeys(role string) ([]sign.Signer, error) {
	if keys, ok := f.signers[role]; ok {
		return keys, nil
	}
	keys, _, err := f.loadKeys(role)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	f.signers[role] = f.privateKeySigners(keys)
	return f.signers[role], nil
}

func (f *fileSystemStore) SavePrivateKey(role string, key *sign.PrivateKey) error {
	if err := f.createDirs(); err != nil {
		return err
	}

	// add the key to the existing keys (if any)
	keys, pass, err := f.loadKeys(role)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	keys = append(keys, key)

	// if loadKeys didn't return a passphrase (because no keys yet exist)
	// and passphraseFunc is set, get the passphrase so the keys file can
	// be encrypted later (passphraseFunc being nil indicates the keys file
	// should not be encrypted)
	if pass == nil && f.passphraseFunc != nil {
		pass, err = f.passphraseFunc(role, true)
		if err != nil {
			return err
		}
	}

	pk := &persistedKeys{}
	if pass != nil {
		pk.Data, err = encrypted.Marshal(keys, pass)
		if err != nil {
			return err
		}
		pk.Encrypted = true
	} else {
		pk.Data, err = json.MarshalIndent(keys, "", "\t")
		if err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(pk, "", "\t")
	if err != nil {
		return err
	}
	if err := util.AtomicallyWriteFile(f.keysPath(role), append(data, '\n'), 0600); err != nil {
		return err
	}
	f.signers[role] = f.privateKeySigners(keys)
	return nil
}

func (f *fileSystemStore) privateKeySigners(keys []*sign.PrivateKey) []sign.Signer {
	res := make([]sign.Signer, len(keys))
	for i, k := range keys {
		res[i] = k.Signer()
	}
	return res
}

// loadKeys loads keys for the given role and returns them along with the
// passphrase (if read) so that callers don't need to re-read it.
func (f *fileSystemStore) loadKeys(role string) ([]*sign.PrivateKey, []byte, error) {
	file, err := os.Open(f.keysPath(role))
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	pk := &persistedKeys{}
	if err := json.NewDecoder(file).Decode(pk); err != nil {
		return nil, nil, err
	}

	var keys []*sign.PrivateKey
	if !pk.Encrypted {
		if err := json.Unmarshal(pk.Data, &keys); err != nil {
			return nil, nil, err
		}
		return keys, nil, nil
	}

	// the keys are encrypted so cannot be loaded if passphraseFunc is not set
	if f.passphraseFunc == nil {
		return nil, nil, ErrPassphraseRequired{role}
	}

	// try the empty string as the password first
	pass := []byte("")
	if err := encrypted.Unmarshal(pk.Data, &keys, pass); err != nil {
		pass, err = f.passphraseFunc(role, false)
		if err != nil {
			return nil, nil, err
		}
		if err = encrypted.Unmarshal(pk.Data, &keys, pass); err != nil {
			return nil, nil, err
		}
	}
	return keys, pass, nil
}

func (f *fileSystemStore) keysPath(role string) string {
	return filepath.Join(f.dir, "keys", role+".json")
}

func (f *fileSystemStore) Clean() error {
	_, err := os.Stat(filepath.Join(f.repoDir(), "root.json"))
	if os.IsNotExist(err) {
		return ErrNewRepository
	} else if err != nil {
		return err
	}
	if err := os.RemoveAll(f.stagedDir()); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(f.stagedDir(), "targets"), 0755)
}

func computeTargetPaths(consistentSnapshot bool, name string, hashes map[string]data.Hashes) []string {
	if consistentSnapshot {
		return util.HashedPaths(name, hashes[name])
	} else {
		return []string{name}
	}
}

func computeMetadataPaths(consistentSnapshot bool, name string, versions map[string]int) []string {
	copyVersion := false

	switch name {
	case "root.json":
		copyVersion = true
	case "timestamp.json":
		copyVersion = false
	default:
		if consistentSnapshot {
			copyVersion = true
		} else {
			copyVersion = false
		}
	}

	paths := []string{name}
	if copyVersion {
		paths = append(paths, util.VersionedPath(name, versions[name]))
	}

	return paths
}
