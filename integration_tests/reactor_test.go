package tests

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/db"
)

type TestReactor struct {
	expect   []byte
	created  chan struct{}
	modified chan struct{}
	deleted  chan struct{}
	log      *zap.Logger
}

func (t *TestReactor) Created(key string, value []byte) {
	if bytes.Equal(value, t.expect) {
		t.log.Debug("received created callback")
		close(t.created)
	}
}

func (t *TestReactor) Modified(key string, newValue []byte) {
	if bytes.Equal(newValue, t.expect) {
		t.log.Debug("received modified callback")
		close(t.modified)
	}
}

func (t *TestReactor) Deleted(key string, lastKnownValue []byte) {
	if bytes.Equal(lastKnownValue, t.expect) {
		t.log.Debug("received deleted callback")
		close(t.deleted)
	}
}

func randomHumanReadableBytes(n int) []byte {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	for i, v := range buf {
		// make sure newV is in the printable range
		newV := 32 + (v % 94)
		buf[i] = newV
	}
	return buf
}

func watchTests(listener *db.PathWatcher, buf []byte, trx *TestReactor, kv store.Store, log *zap.Logger) {
	rxShutdown := time.After(30 * time.Second)

	waitForIt := func(err error, listen chan struct{}) {
		if err != nil {
			panic(err)
		}
		select {
		case <-listen:
		case <-rxShutdown:
			panic("did not receive expected update within timeout")
		}
	}

	writer := db.NewPrefixedStore("/test1", kv)

	waitForIt(writer.Put("k1", buf, nil), trx.created)
	waitForIt(writer.Put("k1", buf, nil), trx.modified)
	waitForIt(writer.Delete("k1"), trx.deleted)
}

func TestReactiveCfgStore(t *testing.T) {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	log, _ := cfg.Build()

	kv, shutdownChan, stoppedChan := TestingEtcd()
	if shutdownChan == nil {
		panic("could not start testing etcd")
	}

	buf := randomHumanReadableBytes(10)

	// watch for events with the reactor
	trx := &TestReactor{
		expect:   buf,
		created:  make(chan struct{}),
		modified: make(chan struct{}),
		deleted:  make(chan struct{}),
		log:      log,
	}
	closeReact := make(chan struct{})

	listener := db.NewPathWatcher("/test1", kv, log)
	listener.ReconciliationJitter = 0
	listener.ReconciliationBaseDelay = 3

	listener.React(trx, closeReact)

	watchTests(listener, buf, trx, kv, log)

	close(closeReact)
	close(shutdownChan)
	<-stoppedChan
}
