package app

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

var (
	_DBs map[string]*DBx
)

func DB(file string) (*DBx, error) {
	if len(_DBs) == 0 {
		_DBs = make(map[string]*DBx)
	}
	if _, ok := _DBs[file]; !ok {
		instance := &DBx{File: file}
		if err := instance.Init(); err != nil {
			return nil, err
		}
		_DBs[file] = instance
	}
	return _DBs[file], nil
}

type DBx struct {
	File string
	db   *bolt.DB
}

func (t *DBx) Init() error {
	if t.File == "" {
		return errors.New("DB | error DB.Init() | missing file name")
	}
	if t.db != nil {
		return errors.New("DB | error DB.Init() | initialization already set")
	}
	db, err := bolt.Open(t.File, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	t.db = db
	return nil
}

func (t *DBx) Close() error {
	return t.db.Close()
}

func (t *DBx) Get(bucket, key string) (val string, ok bool) {
	err := t.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("DB | error DB.Get() | bucket '" + bucket + "' not found")
		}
		if v := b.Get([]byte(key)); v != nil {
			val = string(v)
			ok = true
		}
		return nil
	})
	if err != nil {
		log.Printf("DB | error DB.Get() | %s", err)
	}
	return
}

func (t *DBx) Set(bucket, key, val string) {
	err := t.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return errors.New("DB | error DB.Set() | bucket '" + bucket + "' cannot be created")
		}
		err = b.Put([]byte(key), []byte(val))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("DB | error DB.Set() | %s", err)
	}
}

func (t *DBx) Del(bucket, key string) {
	err := t.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("DB | error DB.Del() | bucket '" + bucket + "' not found")
		}
		if err := b.Delete([]byte(key)); err != nil {
			log.Printf("WTF: %s\n", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("DB | error DB.Del() | %s", err)
	}
	return
}

func (t *DBx) BucketGet(bucket string) map[string]string {
	list := make(map[string]string)
	err := t.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New("DB | error DB.BucketGet() | bucket '" + bucket + "' not found")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			list[string(k)] = string(v)
		}
		return nil
	})
	if err != nil {
		log.Printf("DB | error DB.BucketGet() | %s", err)
	}
	return list
}

func (t *DBx) BucketDel(bucket string) {
	err := t.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
	if err != nil {
		log.Printf("DB | error DB.BucketDel() | %s", err)
	}
	return
}

func (t *DBx) PrepareValues(a ...interface{}) []string {
	args := []string{}
	for _, v := range a {
		switch v.(type) {
		case string:
			args = append(args, v.(string))
		case int:
			args = append(args, strconv.Itoa(v.(int)))
		case byte: // uint8
			args = append(args, strconv.FormatInt(int64(v.(byte)), 10))
		case int64:
			args = append(args, strconv.FormatInt(v.(int64), 10))
		case float64:
			args = append(args, strconv.FormatFloat(v.(float64), 'f', 6, 64))
		case bool:
			args = append(args, strconv.FormatBool(v.(bool)))
		default:
			log.Printf("DB | error DB.PrepareValue() | missing support for type '%T' to 'string' conversion", v)
		}
	}
	return args
}

func (t *DBx) PrepareValue(a ...interface{}) string {
	return strings.Join(t.PrepareValues(a...), "|")
}
