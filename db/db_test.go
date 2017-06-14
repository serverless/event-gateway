package db

import (
	"bytes"
	"crypto/rand"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

const (
	etcdDir     = "testing.etcd"
	etcdCliAddr = "127.0.0.1:2389"
)

func testingEtcd(log *zap.Logger) (chan struct{}, <-chan struct{}) {
	shutdownInitiateChan := make(chan struct{})
	cleanupChan := make(chan struct{})

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	startedChan, stoppedChan := EmbedEtcd(wd+"/"+etcdDir, "http://localhost:2390",
		"http://"+etcdCliAddr, shutdownInitiateChan, log, false)

	select {
	case <-startedChan:
	case <-stoppedChan:
		panic("Failed to start testing etcd")
	}

	go func() {
		<-stoppedChan
		err := os.RemoveAll(wd + "/" + etcdDir)
		if err != nil {
			panic(err)
		}
		close(cleanupChan)
	}()

	return shutdownInitiateChan, cleanupChan
}

type TestReactor struct {
	expect   []byte
	created  chan struct{}
	modified chan struct{}
	deleted  chan struct{}
}

func (t *TestReactor) Created(key string, value []byte) {
	if bytes.Equal(value, t.expect) {
		close(t.created)
	}
}

func (t *TestReactor) Modified(key string, newValue []byte) {
	if bytes.Equal(newValue, t.expect) {
		close(t.modified)
	}
}

func (t *TestReactor) Deleted(key string, lastKnownValue []byte) {
	if bytes.Equal(lastKnownValue, t.expect) {
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

func watchTests(listener *ReactiveCfgStore, buf []byte, trx TestReactor) {
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

	waitForIt(listener.Put("k1", buf, nil), trx.created)
	waitForIt(listener.Put("k1", buf, nil), trx.modified)
	waitForIt(listener.Delete("k1"), trx.deleted)
}

func getSetTests(log *zap.Logger) {
	buf := randomHumanReadableBytes(10)

	writer := NewReactiveCfgStore("/test2", []string{etcdCliAddr}, log)

	// clear state before continuing
	writer.Delete("k1")

	_, err := writer.CachedGet("k1")
	if err == nil {
		panic("should not have gotten key")
	}
	err = writer.Put("k1", buf, nil)
	if err != nil {
		panic("could not set key")
	}
	val, err := writer.CachedGet("k1")
	if err != nil {
		panic("writer could not get key that was set")
	}
	if !bytes.Equal(val, buf) {
		panic("read a value other than the one that was written")
	}
	err = writer.Delete("k1")
	if err != nil {
		panic("could not delete key")
	}
	_, err = writer.CachedGet("k1")
	if err == nil {
		panic("got a key that should have been deleted")
	}
}

func TestReactiveCfgStore(t *testing.T) {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	log, _ := cfg.Build()

	buf := randomHumanReadableBytes(10)

	// watch for events with the reactor
	trx := TestReactor{
		expect:   buf,
		created:  make(chan struct{}),
		modified: make(chan struct{}),
		deleted:  make(chan struct{}),
	}
	closeReact := make(chan struct{})

	listener := NewReactiveCfgStore("/test1", []string{etcdCliAddr}, log)
	listener.React(&trx, closeReact)

	shutdownChan, stoppedChan := testingEtcd(log)
	if shutdownChan == nil {
		panic("could not start testing etcd")
	}

	watchTests(listener, buf, trx)
	getSetTests(log)

	close(closeReact)
	close(shutdownChan)
	<-stoppedChan
}
