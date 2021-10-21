package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/phuslu/log"
	"gorobei/utils"
	"strings"
	"time"
)

type (
	Db struct {
		b *badger.DB
	}

	DailyReport struct {
		Run    int
		Posted int
		Errors int
		Total int
		SentAt time.Time
	}

	UsersStore interface {
		StoreUserId(user string, id int64) error
		ReadUserId(user string) (int64, error)
	}
)

var (
	ErrNotFound = errors.New("key not found")
	_ UsersStore = (*Db)(nil)
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

func (d *Db) StoreUrlProcessed(key string) (byte, error) {
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

func (d *Db) ReadUrlProcessed(key string, value byte) error {

	err := d.b.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte{value})
		return err
	})

	return err
}

func (d *Db) constructUserKey(user string) []byte {
	return []byte("username_" + strings.ToLower(user))
}

func (d *Db) StoreUserId(user string, id int64) error {
	if user == "" {
		return errors.New("empty user name")
	}
	err := d.b.Update(func(txn *badger.Txn) error {
		err := txn.Set(d.constructUserKey(user), utils.Int64ToByteArr(id))
		return err
	})

	return err
}

func (d *Db) ReadUserId(user string) (int64, error) {
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
	id, err := utils.ByteArrToInt64(v)
	if err != nil {
		return 0, err
	}
	return id, nil
}

var dbKeyDailyReport = []byte("daily_report")

func (d *Db) StoreDailyReport(r *DailyReport) error {
	var err error
	err = d.b.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(r)
		if err != nil {
			return err
		}
		err = txn.Set(dbKeyDailyReport, data)
		return err
	})
	return err
}

func (d *Db) ReadDailyReport() (*DailyReport, error) {
	var er2 error
	var r DailyReport

	er2 = d.b.View(func(txn *badger.Txn) error {
		item, err := txn.Get(dbKeyDailyReport)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotFound
			}
			return err
		}

		err = item.Value(func(val []byte) error {
			err := json.Unmarshal(val, &r)
			return err
		})
		return err

	})
	if er2 != nil {
		return nil, er2
	}
	return &r, nil
}