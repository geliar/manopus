package boltdb

import (
	"context"

	"github.com/boltdb/bolt"

	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/store"
)

func init() {
	store.RegisterBuilder(log.Logger.WithContext(context.Background()), serviceName, builder)
}

func builder(ctx context.Context, name string, config map[string]interface{}) {
	l := logger(ctx)
	l = l.With().
		Str("store_name", name).
		Str("store_type", serviceName).
		Logger()
	ctx = l.WithContext(ctx)
	l.Debug().Msgf("Initializing new instance of %s", name)

	i := new(BoltDB)
	i.name = name
	i.bucket = config["bucket"].(string)
	file, _ := config["file"].(string)
	var err error
	if file == "" {
		l.Fatal().Msg("Cannot initialize BoltDB store with empty file field")
	}
	if i.bucket == "" {
		l.Fatal().Msg("Cannot initialize BoltDB store with empty bucket field")
	}

	i.db, err = bolt.Open(file, 0600, nil)
	if err != nil {
		l.Fatal().Str("file", file).Msg("Cannot open BoltDB file")
	}
	err = i.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(i.bucket))
		return err
	})
	if err != nil {
		l.Fatal().
			Str("file", file).
			Str("bucket", i.bucket).
			Msg("Cannot open BoltDB bucket")
	}
	store.RegisterStore(ctx, i)
}
