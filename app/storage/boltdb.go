package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/dotzero/hooks/app/models"
)

// BoltDB is a wrapper over Bolt DB
type BoltDB struct {
	db *bolt.DB
}

var (
	// BucketHooks name of the hooks bucket
	BucketHooks = []byte("hooks")
	// BucketHooksTTL name of the hooks ttl bucket
	BucketHooksTTL = []byte("hooks_ttl")
	// BucketReqs name of the requests bucket
	BucketReqs = []byte("requests")
	// BucketReqsTTL name of the requests bucket
	BucketReqsTTL = []byte("requests_ttl")
	// BucketCounters name of the counters bucket
	BucketCounters = []byte("counters")
)

// New returns a wrapper over Bolt DB
func New(path string) (*BoltDB, error) {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, err
	}

	// ensure buckets exists
	buckets := [][]byte{BucketHooks, BucketHooksTTL, BucketReqs, BucketReqsTTL, BucketCounters}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range buckets {
			if _, e := tx.CreateBucketIfNotExists(name); e != nil {
				return fmt.Errorf("failed to create `%s` bucket: %w", name, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &BoltDB{
		db: db,
	}, nil
}

// Close releases all database resources
func (b *BoltDB) Close() error {
	return b.db.Close()
}

// Hook returns hook model by name
func (b *BoltDB) Hook(name string) (*models.Hook, error) {
	var hook *models.Hook

	err := b.db.View(func(tx *bolt.Tx) error {
		bHooks := tx.Bucket(BucketHooks)

		return b.load(bHooks, name, &hook)
	})

	return hook, err
}

// PutHook save hook model into storage
func (b *BoltDB) PutHook(hook *models.Hook) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bHooks := tx.Bucket(BucketHooks)
		if err := b.save(bHooks, hook.Name, hook); err != nil {
			return err
		}

		bTTL := tx.Bucket(BucketHooksTTL)
		key := []byte(hook.Created.UTC().Format(time.RFC3339Nano))
		if err := bTTL.Put(key, []byte(hook.Name)); err != nil {
			return err
		}

		bCounters := tx.Bucket(BucketCounters)
		count := btoi(bCounters.Get(BucketHooks)) + 1

		return bCounters.Put(BucketHooks, itob(count))
	})
}

// RecentHooks returns recent public hooks
func (b *BoltDB) RecentHooks(max int) ([]*models.Hook, error) {
	hooks := make([]*models.Hook, 0, max)

	err := b.db.View(func(tx *bolt.Tx) error {
		bHooks := tx.Bucket(BucketHooks)

		return bHooks.ForEach(func(k, v []byte) error {
			var hook models.Hook

			err := json.Unmarshal(v, &hook)
			if err != nil {
				return fmt.Errorf("failed to unmarshal: %w", err)
			}

			if !hook.Private {
				hooks = append(hooks, &hook)
			}

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(hooks, func(i, j int) bool {
		return hooks[i].Created.After(hooks[j].Created)
	})

	if len(hooks) > max {
		return hooks[0:max], nil
	}

	return hooks, nil
}

// Requests returns hook requests by hook name
func (b *BoltDB) Requests(hook string) ([]*models.Request, error) {
	requests := make([]*models.Request, 0)

	err := b.db.View(func(tx *bolt.Tx) error {
		bRequests := tx.Bucket(BucketReqs).Bucket([]byte(hook))
		if bRequests == nil {
			return nil
		}

		return bRequests.ForEach(func(k, v []byte) error {
			request := &models.Request{}
			if err := json.Unmarshal(v, &request); err != nil {
				return err
			}

			requests = append(requests, request)

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return requests, nil
}

// PutRequest save request model into storage
func (b *BoltDB) PutRequest(hook string, req *models.Request) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bReqs, err := b.reqsBucket(tx, hook)
		if err != nil {
			return fmt.Errorf("can't get requests bucket for %s: %w", hook, err)
		}

		if err := b.save(bReqs, req.Name, req); err != nil {
			return err
		}

		bTTL := tx.Bucket(BucketReqsTTL)
		key := []byte(req.Created.UTC().Format(time.RFC3339Nano))
		if err := bTTL.Put(key, []byte(req.Name)); err != nil {
			return err
		}

		bCounters := tx.Bucket(BucketCounters)
		count := btoi(bCounters.Get(BucketReqs)) + 1

		return bCounters.Put(BucketReqs, itob(count))
	})
}

func (b *BoltDB) reqsBucket(tx *bolt.Tx, name string) (*bolt.Bucket, error) {
	bkt, err := tx.Bucket(BucketReqs).CreateBucketIfNotExists([]byte(name))
	if err != nil {
		return nil, err
	}

	return bkt, nil
}

// Sweep performs a batch delete of all bucket items using the keys picked up from expired func
func (b *BoltDB) Sweep(bktName []byte, ttlName []byte, maxAge time.Duration) (err error) {
	keys, err := b.Expired(ttlName, maxAge)
	if err != nil || len(keys) == 0 {
		return
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bktName)

		for _, key := range keys {
			if err = bkt.Delete(key); err != nil {
				return err
			}
		}

		bCounters := tx.Bucket(BucketCounters)
		count := btoi(bCounters.Get(bktName)) - len(keys)

		return bCounters.Put(bktName, itob(count))
	})
}

// Expired returns list of keys that have ttl older than maxAge
func (b *BoltDB) Expired(ttlName []byte, maxAge time.Duration) (keys [][]byte, err error) {
	keys = [][]byte{}
	ttlKeys := [][]byte{}

	err = b.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(ttlName).Cursor()

		max := []byte(time.Now().UTC().Add(-maxAge).Format(time.RFC3339Nano))
		for k, v := c.First(); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			keys = append(keys, v)
			ttlKeys = append(ttlKeys, k)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ttlName)
		for _, key := range ttlKeys {
			if err = b.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})

	return
}

// Count returns number of keys in bucket
func (b *BoltDB) Count(name []byte) (int, error) {
	var stats bolt.BucketStats

	err := b.db.View(func(tx *bolt.Tx) error {
		stats = tx.Bucket(name).Stats()
		return nil
	})
	if err != nil {
		return 0, err
	}

	return stats.KeyN, nil
}

// save marshaled value to key for bucket. Should run in update tx
func (b *BoltDB) save(bkt *bolt.Bucket, key string, value interface{}) error {
	if value == nil {
		return fmt.Errorf("can't save nil value for %s", key)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("can't marshal: %w", err)
	}

	err = bkt.Put([]byte(key), data)
	if err != nil {
		return fmt.Errorf("failed to save key %s: %w", key, err)
	}

	return nil
}

// load and unmarshal json value by key from bucket. Should run in view tx
func (b *BoltDB) load(bkt *bolt.Bucket, key string, res interface{}) error {
	value := bkt.Get([]byte(key))
	if value == nil {
		return fmt.Errorf("no value for %s", key)
	}

	err := json.Unmarshal(value, &res)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	return nil
}

func itob(i int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))

	return b
}

func btoi(b []byte) int {
	if len(b) == 0 {
		return 0
	}

	return int(binary.BigEndian.Uint32(b))
}
