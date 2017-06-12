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
	etcdCliAddr = "localhost:2389"
)

func testingEtcd() (chan struct{}, <-chan struct{}) {
	shutdownInitiateChan := make(chan struct{})
	cleanupChan := make(chan struct{})

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	startedChan, stoppedChan := EmbedEtcd(wd+"/"+etcdDir, "http://localhost:2390",
		"http://"+etcdCliAddr, shutdownInitiateChan, nil, false)

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

func withEtcd(f func()) {
	shutdownChan, stoppedChan := testingEtcd()
	if shutdownChan == nil {
		panic("could not start testing etcd")
	}

	f()

	close(shutdownChan)
	<-stoppedChan
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

func TestWatch(t *testing.T) {
	withEtcd(func() {
		buf := randomHumanReadableBytes(10)
		log, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}

		trx := TestReactor{
			expect:   buf,
			created:  make(chan struct{}),
			modified: make(chan struct{}),
			deleted:  make(chan struct{}),
		}

		listener := NewReactiveCfgStore("/test", []string{etcdCliAddr}, log)
		configurer := NewReactiveCfgStore("/test", []string{etcdCliAddr}, log)

		// clear state before continuing
		configurer.Delete("k1")

		// watch for events with the reactor
		closeReact := make(chan struct{})
		listener.React(&trx, closeReact)

		rxShutdown := time.After(10 * time.Second)

		doIt := func(err error, listen chan struct{}, shutdown <-chan time.Time) {
			if err != nil {
				panic(err)
			}
			select {
			case <-listen:
			case <-shutdown:
				panic("did not receive creation update within timeout")
			}
		}

		doIt(configurer.Put("k1", buf, nil), trx.created, rxShutdown)
		doIt(configurer.Put("k1", buf, nil), trx.modified, rxShutdown)
		doIt(configurer.Delete("k1"), trx.deleted, rxShutdown)

		close(closeReact)
	})
}
