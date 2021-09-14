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

package tuf

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"sync"

	"github.com/pkg/errors"

	"github.com/theupdateframework/go-tuf"
	"github.com/theupdateframework/go-tuf/client"
	tuf_leveldbstore "github.com/theupdateframework/go-tuf/client/leveldbstore"
	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/util"
)

const (
	TufRootEnv        = "TUF_ROOT"
	defaultLocalStore = ".sigstore/root/"
)

// Global TUF client. Stores local targets in $HOME/.sigstore/root.
// Could be in memory local store, but that would mean re-download each time cosign is run.
var rootClient *client.Client
var rootClientMu = &sync.Mutex{}

func CosignRoot() string {
	rootDir := os.Getenv(TufRootEnv)
	if rootDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = ""
		}
		return path.Join(home, defaultLocalStore)
	}
	return rootDir
}

func CosignTargets() string {
	return path.Join(CosignRoot(), "targets")
}

// Target destinations compatible with go-tuf.
type targetDestination struct {
	*os.File
}

func (t *targetDestination) Delete() error {
	t.Close()
	return nil
}

type ByteDestination struct {
	*bytes.Buffer
}

func (b *ByteDestination) Delete() error {
	b.Reset()
	return nil
}

func getRootKeys(rootFileBytes []byte) ([]*data.Key, error) {
	store := tuf.MemoryStore(map[string]json.RawMessage{"root.json": rootFileBytes}, nil)
	repo, err := tuf.NewRepo(store)
	if err != nil {
		return nil, err
	}
	return repo.RootKeys()
}

// Gets the global TUF client if the directory exists.
// This will not make a remote call.
func RootClient(ctx context.Context, local string, remote client.RemoteStore) (*client.Client, error) {
	rootClientMu.Lock()
	defer rootClientMu.Unlock()
	if rootClient == nil {
		// Instantiate the global TUF client from the local store. This does not do a download.
		local, err := tuf_leveldbstore.FileLocalStore(local)
		if err != nil {
			return nil, errors.Wrap(err, "initializing local store")
		}
		rootClient = client.NewClient(local, remote)
	}
	return rootClient, nil
}

func updateMetadataAndDownloadTargets(c *client.Client) error {
	// Download initial targets and store in $HOME/.sigstore/root/targets/.
	targetFiles, err := c.Update()
	if err != nil && !client.IsLatestSnapshot(err) {
		return errors.Wrap(err, "updating tuf metadata")
	}

	// Download targets, if they don't already exist and match the updated metadata.
	if err := os.MkdirAll(CosignTargets(), 0700); err != nil {
		return errors.Wrap(err, "creating targets dir")
	}
	for name := range targetFiles {
		if err := downloadTarget(name, c, nil); err != nil {
			return err
		}
	}
	return nil
}

func downloadTarget(name string, c *client.Client, out client.Destination) error {
	f, err := os.Create(path.Join(CosignTargets(), name))
	if err != nil {
		return errors.Wrap(err, "creating target file")
	}
	defer f.Close()
	dest := targetDestination{f}

	if err := c.Download(name, &dest); err != nil {
		return errors.Wrap(err, "downloading target")
	}
	if out != nil {
		_, err = io.Copy(out, dest)
	}
	return err
}

// Instantiates the global TUF client. Downloads all initial targets and stores in $HOME/.sigstore/root/targets/.
func Init(ctx context.Context, rootBytes []byte, remote client.RemoteStore, threshold int) error {
	rootClient, err := RootClient(ctx, CosignRoot(), remote)
	if err != nil {
		return errors.Wrap(err, "initializing root client")
	}
	rootKeys, err := getRootKeys(rootBytes)
	if err != nil {
		return errors.Wrap(err, "retrieving root keys")
	}
	if err := rootClient.Init(rootKeys, threshold); err != nil {
		return errors.Wrap(err, "initializing tuf client")
	}
	// Download initial targets and store in $HOME/.sigstore/root/targets/.
	if err := os.MkdirAll(CosignRoot(), 0755); err != nil {
		return errors.Wrap(err, "creating targets dir")
	}
	if err := updateMetadataAndDownloadTargets(rootClient); err != nil {
		return errors.Wrap(err, "updating local metadata and targets")
	}

	return nil
}

func getTargetHelper(name string, out client.Destination, c *client.Client) error {
	// Get valid target metadata. Does a local verification.
	validMeta, err := c.Target(name)
	if err != nil {
		return errors.Wrap(err, "missing target metadata")
	}

	// Get local target.
	localTarget, err := os.Open(path.Join(CosignTargets(), name))
	if err != nil {
		// If the file does not exist, download the target and copy to out.
		return downloadTarget(name, c, out)
	}

	// Otherwise, the file exists in the local store.
	localMeta, err := util.GenerateTargetFileMeta(localTarget)
	if err != nil {
		return errors.Wrap(err, "generating local target metadata")
	}

	// If local target meta does not match valid meta, update and re-download.
	if err := util.TargetFileMetaEqual(validMeta, localMeta); err != nil {
		if err := updateMetadataAndDownloadTargets(c); err != nil {
			return errors.Wrap(err, "updating target metadata")
		}
		// Try again with updated metadata.
		return getTargetHelper(name, out, c)
	}

	// Target metadata equal, copy the file into out.
	if _, err := localTarget.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := io.Copy(out, localTarget); err != nil {
		return errors.Wrap(err, "copying target")
	}
	return localTarget.Close()
}

func GetTarget(ctx context.Context, name string, out client.Destination) error {
	// Reads the root in CosignRoot() directory. Does not use a remote.
	c, err := RootClient(ctx, CosignRoot(), nil)
	if err != nil {
		return errors.Wrap(err, "retrieving root")
	}

	return getTargetHelper(name, out, c)
}
