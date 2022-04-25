package storage

import (
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
	assert.Equal(t, exp, act)

	assert.Equal(t, 1, mustCount(s, hooksName))
	assert.Equal(t, 1, mustCount(s, hooksTTLName))

	err = s.db.View(func(tx *bolt.Tx) error {
		count := btoi(tx.Bucket(countersName).Get(hooksName))
		assert.Equal(t, 1, count)

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

	assert.Equal(t, 3, mustCount(s, reqsName))
	assert.Equal(t, 2, mustCount(s, reqsTTLName))

	err = s.db.View(func(tx *bolt.Tx) error {
		count := btoi(tx.Bucket(countersName).Get(reqsName))
		assert.Equal(t, 2, count)

		return nil
	})
	assert.NoError(t, err)
}

func TestSweep(t *testing.T) {
	s := newTestBoltDB()
	defer s.Close()

	now := time.Now().UTC()

	for i := 0; i < 10; i++ {
		hook := models.NewHook(false)
		hook.Created = now.Add(time.Duration(-i) * time.Hour)

		err := s.PutHook(hook)
		assert.NoError(t, err)
	}

	assert.Equal(t, 10, mustCount(s, hooksName))
	assert.Equal(t, 10, mustCount(s, hooksTTLName))

	err := s.Sweep(hooksName, hooksTTLName, 5*time.Hour)

	assert.NoError(t, err)
	assert.Equal(t, 5, mustCount(s, hooksName))
	assert.Equal(t, 5, mustCount(s, hooksTTLName))
}

func TestExpired(t *testing.T) {
	s := newTestBoltDB()
	defer s.Close()

	now := time.Now().UTC()

	for i := 0; i < 10; i++ {
		hook := models.NewHook(false)
		hook.Created = now.Add(time.Duration(-i) * time.Hour)

		err := s.PutHook(hook)
		assert.NoError(t, err)
	}

	assert.Equal(t, 10, mustCount(s, hooksName))
	assert.Equal(t, 10, mustCount(s, hooksTTLName))

	keys, err := s.Expired(hooksTTLName, 5*time.Hour)

	assert.NoError(t, err)
	assert.Len(t, keys, 5)
	assert.Equal(t, 10, mustCount(s, hooksName))
	assert.Equal(t, 5, mustCount(s, hooksTTLName)) // deleted
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
	count, err := db.Count(hooksName)
	if err != nil {
		panic(err)
	}

	return count
}
