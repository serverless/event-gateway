package db

import (
	"path"

	"github.com/boltdb/bolt"
)

// DB is as internal DB backed  by BoltDB
type DB struct {
	bolt *bolt.DB
}

// New returns new DB instance
func New(dir string) (*DB, error) {
	bolt, err := bolt.Open(path.Join(dir, "gateway.db"), 0600, nil)
	return &DB{bolt}, err
}

// Set the value for a key in the bucket
func (d *DB) Set(bucket, key string, value []byte) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), value)
	})
}

// Get value for a key from the bucket
func (d *DB) Get(bucket, key string) ([]byte, error) {
	value := []byte{}

	err := d.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket != nil {
			value = append(value, bucket.Get([]byte(key))...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return value, nil

}

// Close releases all database resources. All transactions must be closed before closing the database.
func (d *DB) Close() error {
	return d.bolt.Close()
}
