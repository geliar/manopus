package boltdb

import (
	"context"

	"github.com/boltdb/bolt"
)

type BoltDB struct {
	name   string
	bucket string
	db     *bolt.DB
}

func (s BoltDB) Name() string {
	return s.name
}

func (s BoltDB) Type() string {
	return serviceName
}

func (s *BoltDB) Save(ctx context.Context, key string, value []byte) (err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		if b == nil {
			l := logger(ctx)
			l.Error().Str("bucket", s.bucket).Msg("Cannot get BoltDB bucket")
			return nil
		}
		return b.Put([]byte(key), value)
	})
	return
}

func (s *BoltDB) Load(ctx context.Context, key string) (value []byte, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		if b == nil {
			l := logger(ctx)
			l.Error().Str("bucket", s.bucket).Msg("Cannot get BoltDB bucket")
			return nil
		}
		value = b.Get([]byte(key))
		return nil
	})
	return value, err
}

func (s *BoltDB) Stop(ctx context.Context) {

}
