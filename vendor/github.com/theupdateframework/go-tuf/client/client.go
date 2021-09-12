package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/util"
	"github.com/theupdateframework/go-tuf/verify"
)

const (
	// This is the upper limit in bytes we will use to limit the download
	// size of the root/timestamp roles, since we might not don't know how
	// big it is.
	defaultRootDownloadLimit      = 512000
	defaultTimestampDownloadLimit = 16384
	defaultMaxDelegations         = 32
)

// LocalStore is local storage for downloaded top-level metadata.
type LocalStore interface {
	// GetMeta returns top-level metadata from local storage. The keys are
	// in the form `ROLE.json`, with ROLE being a valid top-level role.
	GetMeta() (map[string]json.RawMessage, error)

	// SetMeta persists the given top-level metadata in local storage, the
	// name taking the same format as the keys returned by GetMeta.
	SetMeta(name string, meta json.RawMessage) error
}

// RemoteStore downloads top-level metadata and target files from a remote
// repository.
type RemoteStore interface {
	// GetMeta downloads the given metadata from remote storage.
	//
	// `name` is the filename of the metadata (e.g. "root.json")
	//
	// `err` is ErrNotFound if the given file does not exist.
	//
	// `size` is the size of the stream, -1 indicating an unknown length.
	GetMeta(name string) (stream io.ReadCloser, size int64, err error)

	// GetTarget downloads the given target file from remote storage.
	//
	// `path` is the path of the file relative to the root of the remote
	//        targets directory (e.g. "/path/to/file.txt").
	//
	// `err` is ErrNotFound if the given file does not exist.
	//
	// `size` is the size of the stream, -1 indicating an unknown length.
	GetTarget(path string) (stream io.ReadCloser, size int64, err error)
}

// Client provides methods for fetching updates from a remote repository and
// downloading remote target files.
type Client struct {
	local  LocalStore
	remote RemoteStore

	// The following four fields represent the versions of metatdata either
	// from local storage or from recently downloaded metadata
	rootVer      int
	targetsVer   int
	snapshotVer  int
	timestampVer int

	// targets is the list of available targets, either from local storage
	// or from recently downloaded targets metadata
	targets data.TargetFiles

	// localMeta is the raw metadata from local storage and is used to
	// check whether remote metadata is present locally
	localMeta map[string]json.RawMessage

	// db is a key DB used for verifying metadata
	db *verify.DB

	// consistentSnapshot indicates whether the remote storage is using
	// consistent snapshots (as specified in root.json)
	consistentSnapshot bool

	// MaxDelegations limits by default the number of delegations visited for any
	// target
	MaxDelegations int
}

func NewClient(local LocalStore, remote RemoteStore) *Client {
	return &Client{
		local:          local,
		remote:         remote,
		MaxDelegations: defaultMaxDelegations,
	}
}

// Init initializes a local repository.
//
// The latest root.json is fetched from remote storage, verified using rootKeys
// and threshold, and then saved in local storage. It is expected that rootKeys
// were securely distributed with the software being updated.
func (c *Client) Init(rootKeys []*data.Key, threshold int) error {
	if len(rootKeys) < threshold {
		return ErrInsufficientKeys
	}
	rootJSON, err := c.downloadMetaUnsafe("root.json", defaultRootDownloadLimit)
	if err != nil {
		return err
	}

	// create a new key database, and add all the public `rootKeys` to it.
	c.db = verify.NewDB()
	rootKeyIDs := make([]string, 0, len(rootKeys))
	for _, key := range rootKeys {
		for _, id := range key.IDs() {
			rootKeyIDs = append(rootKeyIDs, id)
			if err := c.db.AddKey(id, key); err != nil {
				return err
			}
		}
	}

	// add a mock "root" role that trusts the passed in key ids. These keys
	// will be used to verify the `root.json` we just fetched.
	role := &data.Role{Threshold: threshold, KeyIDs: rootKeyIDs}
	if err := c.db.AddRole("root", role); err != nil {
		return err
	}

	// verify that the new root is valid.
	if err := c.decodeRoot(rootJSON); err != nil {
		return err
	}

	return c.local.SetMeta("root.json", rootJSON)
}

// Update downloads and verifies remote metadata and returns updated targets.
//
// It performs the update part of "The client application" workflow from
// section 5.1 of the TUF spec:
//
// https://github.com/theupdateframework/tuf/blob/v0.9.9/docs/tuf-spec.txt#L714
func (c *Client) Update() (data.TargetFiles, error) {
	return c.update(false)
}

func (c *Client) update(latestRoot bool) (data.TargetFiles, error) {
	// Always start the update using local metadata
	if err := c.getLocalMeta(); err != nil {
		if _, ok := err.(verify.ErrExpired); ok {
			if !latestRoot {
				return c.updateWithLatestRoot(nil)
			}
			// this should not be reached as if the latest root has
			// been downloaded and it is expired, updateWithLatestRoot
			// should not have continued the update
			return nil, err
		}
		if latestRoot && isErrRoleThreshold(err) {
			// Root was updated with new keys, so our local metadata is no
			// longer validating. Read only the versions from the local metadata
			// and re-download everything.
			if err := c.getRootAndLocalVersionsUnsafe(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Get timestamp.json, extract snapshot.json file meta and save the
	// timestamp.json locally
	timestampJSON, err := c.downloadMetaUnsafe("timestamp.json", defaultTimestampDownloadLimit)
	if err != nil {
		return nil, err
	}
	snapshotMeta, err := c.decodeTimestamp(timestampJSON)
	if err != nil {
		// ErrRoleThreshold could indicate timestamp keys have been
		// revoked, so retry with the latest root.json
		if isDecodeFailedWithErrRoleThreshold(err) && !latestRoot {
			return c.updateWithLatestRoot(nil)
		}
		return nil, err
	}
	if err := c.local.SetMeta("timestamp.json", timestampJSON); err != nil {
		return nil, err
	}

	// Return ErrLatestSnapshot if we already have the latest snapshot.json
	if c.hasMetaFromTimestamp("snapshot.json", snapshotMeta) {
		return nil, ErrLatestSnapshot{c.snapshotVer}
	}

	// Get snapshot.json, then extract file metas.
	//
	// The snapshot.json is only saved locally after checking
	// targets.json so that it will be re-downloaded on subsequent updates
	// if this update fails.
	// root.json meta should not be stored in the snapshot, if it is,
	// the root will be checked, re-downloaded
	snapshotJSON, err := c.downloadMetaFromTimestamp("snapshot.json", snapshotMeta)
	if err != nil {
		return nil, err
	}
	snapshotMetas, err := c.decodeSnapshot(snapshotJSON)
	if err != nil {
		// ErrRoleThreshold could indicate snapshot keys have been
		// revoked, so retry with the latest root.json
		if isDecodeFailedWithErrRoleThreshold(err) && !latestRoot {
			return c.updateWithLatestRoot(nil)
		}
		return nil, err
	}

	// If we don't have the root.json, download it, save it in local
	// storage and restart the update
	// Root should no longer be pinned in snapshot meta https://github.com/theupdateframework/tuf/pull/988
	if rootMeta, ok := snapshotMetas["root.json"]; ok {
		if !c.hasMetaFromSnapshot("root.json", rootMeta) {
			return c.updateWithLatestRoot(&rootMeta)
		}
	}

	// If we don't have the targets.json, download it, determine updated
	// targets and save targets.json in local storage
	var updatedTargets data.TargetFiles
	targetsMeta := snapshotMetas["targets.json"]
	if !c.hasMetaFromSnapshot("targets.json", targetsMeta) {
		targetsJSON, err := c.downloadMetaFromSnapshot("targets.json", targetsMeta)
		if err != nil {
			return nil, err
		}
		updatedTargets, err = c.decodeTargets(targetsJSON)
		if err != nil {
			return nil, err
		}
		if err := c.local.SetMeta("targets.json", targetsJSON); err != nil {
			return nil, err
		}
	}

	// Save the snapshot.json now it has been processed successfully
	if err := c.local.SetMeta("snapshot.json", snapshotJSON); err != nil {
		return nil, err
	}

	return updatedTargets, nil
}

func (c *Client) updateWithLatestRoot(m *data.SnapshotFileMeta) (data.TargetFiles, error) {
	var rootJSON json.RawMessage
	var err error
	if m == nil {
		rootJSON, err = c.downloadMetaUnsafe("root.json", defaultRootDownloadLimit)
	} else {
		rootJSON, err = c.downloadMetaFromSnapshot("root.json", *m)
	}
	if err != nil {
		return nil, err
	}
	if err := c.decodeRoot(rootJSON); err != nil {
		return nil, err
	}
	if err := c.local.SetMeta("root.json", rootJSON); err != nil {
		return nil, err
	}
	return c.update(true)
}

// getLocalMeta decodes and verifies metadata from local storage.
//
// The verification of local files is purely for consistency, if an attacker
// has compromised the local storage, there is no guarantee it can be trusted.
func (c *Client) getLocalMeta() error {
	meta, err := c.local.GetMeta()
	if err != nil {
		return err
	}

	if rootJSON, ok := meta["root.json"]; ok {
		// unmarshal root.json without verifying as we need the root
		// keys first
		s := &data.Signed{}
		if err := json.Unmarshal(rootJSON, s); err != nil {
			return err
		}
		root := &data.Root{}
		if err := json.Unmarshal(s.Signed, root); err != nil {
			return err
		}
		c.db = verify.NewDB()
		for id, k := range root.Keys {
			if err := c.db.AddKey(id, k); err != nil {
				// TUF is considering in TAP-12 removing the
				// requirement that the keyid hash algorithm be derived
				// from the public key. So to be forwards compatible,
				// we ignore `ErrWrongID` errors.
				//
				// TAP-12: https://github.com/theupdateframework/taps/blob/master/tap12.md
				if _, ok := err.(verify.ErrWrongID); !ok {
					return err
				}
			}
		}
		for name, role := range root.Roles {
			if err := c.db.AddRole(name, role); err != nil {
				return err
			}
		}
		if err := c.db.Verify(s, "root", 0); err != nil {
			return err
		}
		c.consistentSnapshot = root.ConsistentSnapshot
	} else {
		return ErrNoRootKeys
	}

	if snapshotJSON, ok := meta["snapshot.json"]; ok {
		snapshot := &data.Snapshot{}
		if err := c.db.UnmarshalTrusted(snapshotJSON, snapshot, "snapshot"); err != nil {
			return err
		}
		c.snapshotVer = snapshot.Version
	}

	if targetsJSON, ok := meta["targets.json"]; ok {
		targets := &data.Targets{}
		if err := c.db.UnmarshalTrusted(targetsJSON, targets, "targets"); err != nil {
			return err
		}
		c.targetsVer = targets.Version
		// FIXME(TUF-0.9) temporarily support files with leading path separators.
		// c.targets = targets.Targets
		c.loadTargets(targets.Targets)
	}

	if timestampJSON, ok := meta["timestamp.json"]; ok {
		timestamp := &data.Timestamp{}
		if err := c.db.UnmarshalTrusted(timestampJSON, timestamp, "timestamp"); err != nil {
			return err
		}
		c.timestampVer = timestamp.Version
	}

	c.localMeta = meta
	return nil
}

// FIXME(TUF-0.9) TUF is considering removing support for target files starting
// with a leading path separator. In order to be backwards compatible, we'll
// just remove leading separators for now.
func (c *Client) loadTargets(targets data.TargetFiles) {
	c.targets = make(data.TargetFiles)
	for name, meta := range targets {
		c.targets[name] = meta
		c.targets[util.NormalizeTarget(name)] = meta
	}
}

// downloadMetaUnsafe downloads top-level metadata from remote storage without
// verifying it's length and hashes (used for example to download timestamp.json
// which has unknown size). It will download at most maxMetaSize bytes.
func (c *Client) downloadMetaUnsafe(name string, maxMetaSize int64) ([]byte, error) {
	r, size, err := c.remote.GetMeta(name)
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrMissingRemoteMetadata{name}
		}
		return nil, ErrDownloadFailed{name, err}
	}
	defer r.Close()

	// return ErrMetaTooLarge if the reported size is greater than maxMetaSize
	if size > maxMetaSize {
		return nil, ErrMetaTooLarge{name, size, maxMetaSize}
	}

	// although the size has been checked above, use a LimitReader in case
	// the reported size is inaccurate, or size is -1 which indicates an
	// unknown length
	return ioutil.ReadAll(io.LimitReader(r, maxMetaSize))
}

// getRootAndLocalVersionsUnsafe decodes the versions stored in the local
// metadata without verifying signatures to protect against downgrade attacks
// when the root is replaced and contains new keys. It also sets the local meta
// cache to only contain the local root metadata.
func (c *Client) getRootAndLocalVersionsUnsafe() error {
	type versionData struct {
		Signed struct {
			Version int
		}
	}

	meta, err := c.local.GetMeta()
	if err != nil {
		return err
	}

	getVersion := func(name string) (int, error) {
		m, ok := meta[name]
		if !ok {
			return 0, nil
		}
		var data versionData
		if err := json.Unmarshal(m, &data); err != nil {
			return 0, err
		}
		return data.Signed.Version, nil
	}

	c.timestampVer, err = getVersion("timestamp.json")
	if err != nil {
		return err
	}
	c.snapshotVer, err = getVersion("snapshot.json")
	if err != nil {
		return err
	}
	c.targetsVer, err = getVersion("targets.json")
	if err != nil {
		return err
	}

	root, ok := meta["root.json"]
	if !ok {
		return errors.New("tuf: missing local root after downloading, this should not be possible")
	}
	c.localMeta = map[string]json.RawMessage{"root.json": root}

	return nil
}

// remoteGetFunc is the type of function the download method uses to download
// remote files
type remoteGetFunc func(string) (io.ReadCloser, int64, error)

// downloadHashed tries to download the hashed prefixed version of the file.
func (c *Client) downloadHashed(file string, get remoteGetFunc, hashes data.Hashes) (io.ReadCloser, int64, error) {
	// try each hashed path in turn, and either return the contents,
	// try the next one if a 404 is returned, or return an error
	for _, path := range util.HashedPaths(file, hashes) {
		r, size, err := get(path)
		if err != nil {
			if IsNotFound(err) {
				continue
			}
			return nil, 0, err
		}
		return r, size, nil
	}
	return nil, 0, ErrNotFound{file}
}

// download downloads the given target file from remote storage using the get
// function, adding hashes to the path if consistent snapshots are in use
func (c *Client) downloadTarget(file string, get remoteGetFunc, hashes data.Hashes) (io.ReadCloser, int64, error) {
	if c.consistentSnapshot {
		return c.downloadHashed(file, get, hashes)
	} else {
		return get(file)
	}
}

// downloadVersionedMeta downloads top-level metadata from remote storage and
// verifies it using the given file metadata.
func (c *Client) downloadMeta(name string, version int, m data.FileMeta) ([]byte, error) {
	r, size, err := func() (io.ReadCloser, int64, error) {
		if c.consistentSnapshot {
			path := util.VersionedPath(name, version)
			r, size, err := c.remote.GetMeta(path)
			if err == nil {
				return r, size, nil
			}

			return nil, 0, err
		} else {
			return c.remote.GetMeta(name)
		}
	}()
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrMissingRemoteMetadata{name}
		}
		return nil, err
	}
	defer r.Close()

	// return ErrWrongSize if the reported size is known and incorrect
	var stream io.Reader
	if m.Length != 0 {
		if size >= 0 && size != m.Length {
			return nil, ErrWrongSize{name, size, m.Length}
		}

		// wrap the data in a LimitReader so we download at most m.Length bytes
		stream = io.LimitReader(r, m.Length)
	} else {
		stream = r
	}

	return ioutil.ReadAll(stream)
}

func (c *Client) downloadMetaFromSnapshot(name string, m data.SnapshotFileMeta) ([]byte, error) {
	b, err := c.downloadMeta(name, m.Version, m.FileMeta)
	if err != nil {
		return nil, err
	}

	meta, err := util.GenerateSnapshotFileMeta(bytes.NewReader(b), m.HashAlgorithms()...)
	if err != nil {
		return nil, err
	}
	if err := util.SnapshotFileMetaEqual(meta, m); err != nil {
		return nil, ErrDownloadFailed{name, err}
	}
	return b, nil
}

func (c *Client) downloadMetaFromTimestamp(name string, m data.TimestampFileMeta) ([]byte, error) {
	b, err := c.downloadMeta(name, m.Version, m.FileMeta)
	if err != nil {
		return nil, err
	}

	meta, err := util.GenerateTimestampFileMeta(bytes.NewReader(b), m.HashAlgorithms()...)
	if err != nil {
		return nil, err
	}
	if err := util.TimestampFileMetaEqual(meta, m); err != nil {
		return nil, ErrDownloadFailed{name, err}
	}
	return b, nil
}

// decodeRoot decodes and verifies root metadata.
func (c *Client) decodeRoot(b json.RawMessage) error {
	root := &data.Root{}
	if err := c.db.Unmarshal(b, root, "root", c.rootVer); err != nil {
		return ErrDecodeFailed{"root.json", err}
	}
	c.rootVer = root.Version
	c.consistentSnapshot = root.ConsistentSnapshot
	return nil
}

// decodeSnapshot decodes and verifies snapshot metadata, and returns the new
// root and targets file meta.
func (c *Client) decodeSnapshot(b json.RawMessage) (data.SnapshotFiles, error) {
	snapshot := &data.Snapshot{}
	if err := c.db.Unmarshal(b, snapshot, "snapshot", c.snapshotVer); err != nil {
		return data.SnapshotFiles{}, ErrDecodeFailed{"snapshot.json", err}
	}
	c.snapshotVer = snapshot.Version
	return snapshot.Meta, nil
}

// decodeTargets decodes and verifies targets metadata, sets c.targets and
// returns updated targets.
func (c *Client) decodeTargets(b json.RawMessage) (data.TargetFiles, error) {
	targets := &data.Targets{}
	if err := c.db.Unmarshal(b, targets, "targets", c.targetsVer); err != nil {
		return nil, ErrDecodeFailed{"targets.json", err}
	}
	updatedTargets := make(data.TargetFiles)
	for path, meta := range targets.Targets {
		if local, ok := c.targets[path]; ok {
			if err := util.TargetFileMetaEqual(local, meta); err == nil {
				continue
			}
		}
		updatedTargets[path] = meta
	}
	c.targetsVer = targets.Version
	// FIXME(TUF-0.9) temporarily support files with leading path separators.
	// c.targets = targets.Targets
	c.loadTargets(targets.Targets)
	return updatedTargets, nil
}

// decodeTimestamp decodes and verifies timestamp metadata, and returns the
// new snapshot file meta.
func (c *Client) decodeTimestamp(b json.RawMessage) (data.TimestampFileMeta, error) {
	timestamp := &data.Timestamp{}
	if err := c.db.Unmarshal(b, timestamp, "timestamp", c.timestampVer); err != nil {
		return data.TimestampFileMeta{}, ErrDecodeFailed{"timestamp.json", err}
	}
	c.timestampVer = timestamp.Version
	return timestamp.Meta["snapshot.json"], nil
}

// hasMetaFromSnapshot checks whether local metadata has the given meta
func (c *Client) hasMetaFromSnapshot(name string, m data.SnapshotFileMeta) bool {
	_, ok := c.localMetaFromSnapshot(name, m)
	return ok
}

// localMetaFromSnapshot returns localmetadata if it matches the snapshot
func (c *Client) localMetaFromSnapshot(name string, m data.SnapshotFileMeta) (json.RawMessage, bool) {
	b, ok := c.localMeta[name]
	if !ok {
		return nil, false
	}
	meta, err := util.GenerateSnapshotFileMeta(bytes.NewReader(b), m.HashAlgorithms()...)
	if err != nil {
		return nil, false
	}
	err = util.SnapshotFileMetaEqual(meta, m)
	return b, err == nil
}

// hasTargetsMeta checks whether local metadata has the given snapshot meta
func (c *Client) hasTargetsMeta(m data.SnapshotFileMeta) bool {
	b, ok := c.localMeta["targets.json"]
	if !ok {
		return false
	}
	meta, err := util.GenerateSnapshotFileMeta(bytes.NewReader(b), m.HashAlgorithms()...)
	if err != nil {
		return false
	}
	err = util.SnapshotFileMetaEqual(meta, m)
	return err == nil
}

// hasSnapshotMeta checks whether local metadata has the given meta
func (c *Client) hasMetaFromTimestamp(name string, m data.TimestampFileMeta) bool {
	b, ok := c.localMeta[name]
	if !ok {
		return false
	}
	meta, err := util.GenerateTimestampFileMeta(bytes.NewReader(b), m.HashAlgorithms()...)
	if err != nil {
		return false
	}
	err = util.TimestampFileMetaEqual(meta, m)
	return err == nil
}

type Destination interface {
	io.Writer
	Delete() error
}

// Download downloads the given target file from remote storage into dest.
//
// dest will be deleted and an error returned in the following situations:
//
//   * The target does not exist in the local targets.json
//   * Failed to fetch the chain of delegations accessible from local snapshot.json
//   * The target does not exist in any targets
//   * Metadata cannot be generated for the downloaded data
//   * Generated metadata does not match local metadata for the given file
func (c *Client) Download(name string, dest Destination) (err error) {
	// delete dest if there is an error
	defer func() {
		if err != nil {
			dest.Delete()
		}
	}()

	// populate c.targets from local storage if not set
	if c.targets == nil {
		if err := c.getLocalMeta(); err != nil {
			return err
		}
	}

	normalizedName := util.NormalizeTarget(name)
	localMeta, ok := c.targets[normalizedName]
	if !ok {
		// search in delegations
		localMeta, err = c.getTargetFileMeta(normalizedName)
		if err != nil {
			return err
		}
	}

	// get the data from remote storage
	r, size, err := c.downloadTarget(normalizedName, c.remote.GetTarget, localMeta.Hashes)
	if err != nil {
		return err
	}
	defer r.Close()

	// return ErrWrongSize if the reported size is known and incorrect
	if size >= 0 && size != localMeta.Length {
		return ErrWrongSize{name, size, localMeta.Length}
	}

	// wrap the data in a LimitReader so we download at most localMeta.Length bytes
	stream := io.LimitReader(r, localMeta.Length)

	// read the data, simultaneously writing it to dest and generating metadata
	actual, err := util.GenerateTargetFileMeta(io.TeeReader(stream, dest), localMeta.HashAlgorithms()...)
	if err != nil {
		return ErrDownloadFailed{name, err}
	}

	// check the data has the correct length and hashes
	if err := util.TargetFileMetaEqual(actual, localMeta); err != nil {
		if e, ok := err.(util.ErrWrongLength); ok {
			return ErrWrongSize{name, e.Actual, e.Expected}
		}
		return ErrDownloadFailed{name, err}
	}

	return nil
}

// Target returns the target metadata for a specific target if it
// exists. If it does not, ErrNotFound will be returned.
func (c *Client) Target(name string) (data.TargetFileMeta, error) {
	m, err := c.Targets()
	if err != nil {
		return data.TargetFileMeta{}, err
	}

	target, ok := m[util.NormalizeTarget(name)]
	if ok {
		return target, nil
	}

	return data.TargetFileMeta{}, ErrNotFound{name}
}

// Targets returns the complete list of available targets.
func (c *Client) Targets() (data.TargetFiles, error) {
	// populate c.targets from local storage if not set
	if c.targets == nil {
		if err := c.getLocalMeta(); err != nil {
			return nil, err
		}
	}
	return c.targets, nil
}
