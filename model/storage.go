package model

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

type LStorage struct {
	db     *leveldb.DB
	closed bool
	mx     sync.RWMutex
}

func (d *LStorage) Open(path string) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return err
	}
	d.db = db
	d.closed = false
	return err
}

func (d *LStorage) IsClosed() bool {
	d.mx.Lock()
	defer d.mx.Unlock()
	return d.closed
}

func (d *LStorage) Put(key []byte, val []byte) error {
	return d.db.Put(key, val, nil)
}

func (d *LStorage) Get(key []byte) (val []byte, err error) {
	return d.db.Get(key, nil)
}
