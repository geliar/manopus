package boltdb

import (
	"context"

	bolt "go.etcd.io/bbolt"
)

// BoltDB store implementation
type BoltDB struct {
	name   string
	bucket string
	db     *bolt.DB
}

// Name returns name of the store
func (s BoltDB) Name() string {
	return s.name
}

// Type returns type of store
func (s BoltDB) Type() string {
	return serviceName
}

// Save puts data to specified store
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

// Load returns data from the store
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

// Stop stops store
func (s *BoltDB) Stop(ctx context.Context) {
	_ = s.db.Close()
}
