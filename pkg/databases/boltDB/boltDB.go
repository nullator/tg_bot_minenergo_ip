package boltdb

import (
	"sync"

	"github.com/boltdb/bolt"
)

type Database struct {
	db    *bolt.DB
	mutex sync.Mutex
}

func NewDatabase(db *bolt.DB) *Database {
	return &Database{db: db}
}

func (db *Database) Save(key string, value string, bucket string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(value))
	})
	return err
}

func (db *Database) Get(key string, bucket string) (string, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var value string
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			data := b.Get([]byte(key))
			value = string(data)
		} else {
			value = ""
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, err
}

func (db *Database) GetAll(bucket string) (map[string]string, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	value := make(map[string]string)

	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			return b.ForEach(func(k, v []byte) error {
				value[string(k)] = string(v)
				return nil
			})
		}
		return nil
	})
	return value, err
}
