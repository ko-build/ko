package client

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"

	tuf_client "github.com/theupdateframework/go-tuf/client"
)

func FileLocalStore(path string) (tuf_client.LocalStore, error) {
	fd, err := storage.OpenFile(path, false)
	if err != nil {
		return nil, err
	}

	db, err := leveldb.Open(fd, nil)
	return &fileLocalStore{fd: fd, db: db}, err
}

type fileLocalStore struct {
	fd storage.Storage
	db *leveldb.DB
}

func (f *fileLocalStore) GetMeta() (map[string]json.RawMessage, error) {
	meta := make(map[string]json.RawMessage)
	db_itr := f.db.NewIterator(nil, nil)
	for db_itr.Next() {
		vcopy := make([]byte, len(db_itr.Value()))
		copy(vcopy, db_itr.Value())
		meta[string(db_itr.Key())] = vcopy
	}
	db_itr.Release()
	return meta, db_itr.Error()
}

func (f *fileLocalStore) SetMeta(name string, meta json.RawMessage) error {
	return f.db.Put([]byte(name), []byte(meta), nil)
}

func (f *fileLocalStore) Close() error {
	if err := f.db.Close(); err != nil {
		return err
	}
	if err := f.fd.Close(); err != nil {
		return err
	}

	return nil
}
