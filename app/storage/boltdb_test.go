package storage

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"

	"github.com/dotzero/hooks/app/models"
)

func TestHook(t *testing.T) {
	s := newTestBoltDB()
	defer s.Close()

	exp := models.NewHook(true)

	err := s.PutHook(exp)
	assert.NoError(t, err)

	act, err := s.Hook(exp.Name)
	assert.NoError(t, err)

	now := time.Now()
	exp.Created = now
	act.Created = now

	assert.Equal(t, exp, act)

	assert.Equal(t, 1, mustCount(s, BucketHooks))
	assert.Equal(t, 1, mustCount(s, BucketTTL))

	err = s.db.View(func(tx *bolt.Tx) error {
		count := btoi(tx.Bucket(BucketCounters).Get(BucketHooks))
		assert.Equal(t, 1, count)

		return nil
	})
	assert.NoError(t, err)
}

func TestRecentHooks(t *testing.T) {
	s := newTestBoltDB()
	defer s.Close()

	for i := 0; i < 10; i++ {
		err := s.PutHook(models.NewHook(false))
		assert.NoError(t, err)

		err = s.PutHook(models.NewHook(true))
		assert.NoError(t, err)
	}

	assert.Equal(t, 20, mustCount(s, BucketHooks))

	hooks, err := s.RecentHooks(5)

	assert.NoError(t, err)
	assert.Len(t, hooks, 5)
}

func TestSweepHooks(t *testing.T) {
	s := newTestBoltDB()
	defer s.Close()

	now := time.Now()

	for i := 0; i < 10; i++ {
		hook := models.NewHook(false)
		hook.Name = fmt.Sprintf("%d", i)
		hook.Created = now.Add(time.Duration(-i) * time.Hour)

		err := s.PutHook(hook)
		assert.NoError(t, err)

		for j := 0; j < 5; j++ {
			req := &models.Request{
				Name:    fmt.Sprintf("%s|%d", hook.Name, j),
				Created: now.Add(time.Duration(-j) * time.Hour),
			}

			err := s.PutRequest(hook.Name, req)
			assert.NoError(t, err)
		}
	}

	assert.Equal(t, 10, mustCount(s, BucketHooks))
	assert.Equal(t, 60, mustCount(s, BucketReqs))
	assert.Equal(t, 10, mustCount(s, BucketTTL))

	err := s.SweepHooks(time.Duration(7) * time.Hour)
	assert.NoError(t, err)

	assert.Equal(t, 7, mustCount(s, BucketHooks))
	assert.Equal(t, 42, mustCount(s, BucketReqs))
	assert.Equal(t, 7, mustCount(s, BucketTTL))

	err = s.SweepHooks(time.Duration(5) * time.Hour)
	assert.NoError(t, err)

	assert.Equal(t, 5, mustCount(s, BucketHooks))
	assert.Equal(t, 30, mustCount(s, BucketReqs))
	assert.Equal(t, 5, mustCount(s, BucketTTL))

	err = s.db.View(func(tx *bolt.Tx) error {
		count := btoi(tx.Bucket(BucketCounters).Get(BucketHooks))
		assert.Equal(t, 5, count)

		return nil
	})
	assert.NoError(t, err)
}

func TestRequest(t *testing.T) {
	s := newTestBoltDB()
	defer s.Close()

	hook := models.NewHook(false)

	err := s.PutHook(hook)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/foobar?foo=bar", nil)
	getReq := models.NewRequest(req)

	err = s.PutRequest(hook.Name, getReq)
	assert.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/foobar", strings.NewReader(`{"foo": "bar"}`))
	postReq := models.NewRequest(req)

	err = s.PutRequest(hook.Name, postReq)
	assert.NoError(t, err)

	reqs, err := s.Requests(hook.Name)
	assert.NoError(t, err)
	assert.Len(t, reqs, 2)

	assert.Equal(t, 3, mustCount(s, BucketReqs)) // 1 hook keys + 2 nested
}

func TestExpired(t *testing.T) {
	s := newTestBoltDB()
	defer s.Close()

	now := time.Now()

	for i := 0; i < 10; i++ {
		hook := models.NewHook(false)
		hook.Created = now.Add(time.Duration(-i) * time.Hour)

		err := s.PutHook(hook)
		assert.NoError(t, err)
	}

	assert.Equal(t, 10, mustCount(s, BucketHooks))
	assert.Equal(t, 10, mustCount(s, BucketTTL))

	keys, err := s.Expired(BucketTTL, 5*time.Hour)

	assert.NoError(t, err)
	assert.Len(t, keys, 5)
	assert.Equal(t, 10, mustCount(s, BucketHooks))
	assert.Equal(t, 5, mustCount(s, BucketTTL)) // deleted
}

func newTestBoltDB() *BoltDB {
	backend, err := New(tempfile())
	if err != nil {
		panic(err)
	}

	return backend
}

func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}

	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}

	return f.Name()
}

func mustCount(db *BoltDB, bkt []byte) int {
	count, err := db.Count(bkt)
	if err != nil {
		panic(err)
	}

	return count
}

// nolint
func tree(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			fmt.Printf("[%s]\n", name)
			return nested(b, "  ")
		})
	})
}

// nolint
func nested(b *bolt.Bucket, prefix string) error {
	return b.ForEach(func(k, v []byte) error {
		fmt.Printf("%s%s:%s\n", prefix, k, v)

		if b := b.Bucket(k); b != nil {
			return nested(b, prefix+"  ")
		}

		return nil
	})
}
