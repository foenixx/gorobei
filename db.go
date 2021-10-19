package main

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/phuslu/log"
	"strings"
)

type Db struct {
	b *badger.DB
}

var (
	ErrNotFound = errors.New("key not found")
)

var _ badger.Logger = (*LogAdapter)(nil)

func OpenDb(path string) (*Db, error) {
	log.Info().Str("path", path).Msg("opening database")
	opts := badger.DefaultOptions(path).WithLogger(&LogAdapter{&log.DefaultLogger})

	d, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}

	err = d.RunValueLogGC(0.5)
	if err != nil && !errors.Is(err, badger.ErrNoRewrite) {
		return nil, err
	}

	return &Db{b: d}, nil
}

func (d *Db) Close() error {
	return d.b.Close()
}

func (d *Db) Get(key string) (byte, error) {
	var v []byte

	err := d.b.View(func(txn *badger.Txn) error {
		val, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		v, err = val.ValueCopy(nil)
		return err
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if len(v) != 1 {
		return 0, fmt.Errorf("unexpected value: %v", v)
	}
	return v[0], nil
}

func (d *Db) Set(key string, value byte) error {

	err := d.b.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte{value})
		return err
	})

	return err
}

func (d *Db) constructUserKey(user string) []byte {
	return []byte("username_" + strings.ToLower(user))
}

func (d *Db) SetUserID(user string, id int64) error {
	if user == "" {
		return errors.New("empty user name")
	}
	err := d.b.Update(func(txn *badger.Txn) error {
		err := txn.Set(d.constructUserKey(user), Int64ToByteArr(id))
		return err
	})

	return err
}

func (d *Db) GetUserID(user string) (int64, error) {
	var v []byte

	err := d.b.View(func(txn *badger.Txn) error {
		val, err := txn.Get(d.constructUserKey(user))
		if err != nil {
			return err
		}
		v, err = val.ValueCopy(nil)
		return err
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	id, err := ByteArrToInt64(v)
	if err != nil {
		return 0, err
	}
	return id, nil
}
