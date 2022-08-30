package boltdb

import (
	"log"

	"github.com/boltdb/bolt"
)

type Database struct {
	db *bolt.DB
}

func NewDatabase(db *bolt.DB) *Database {
	return &Database{db}
}

func (db *Database) Save(key string, value string, bucket string) error {
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
	var value string
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		data := b.Get([]byte(key))
		value = string(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, err
}

func (db *Database) GetAll(bucket string) (map[string]string, error) {
	value := make(map[string]string)

	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		log.Println(b)
		if b != nil {
			return b.ForEach(func(k, v []byte) error {
				value[string(k)] = string(v)
				log.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
		}
		return nil
	})
	return value, err
}
