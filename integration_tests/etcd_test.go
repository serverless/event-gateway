package tests

import (
	"os"
	"strconv"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"

	"github.com/serverless/event-gateway/db"
)

func init() {
	etcd.Register()
}

func kvAddr(port int) string {
	return "127.0.0.1:" + strconv.Itoa(port)
}

func TestingEtcd() (store.Store, chan struct{}, <-chan struct{}) {
	shutdownInitiateChan := make(chan struct{})
	cleanupChan := make(chan struct{})

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	peerPort := newPort()
	peerAddr := "http://localhost:" + strconv.Itoa(peerPort)

	etcdDir := "testing.etcd"
	dataDir := wd + "/" + etcdDir + "." + strconv.Itoa(peerPort)

	cliPort := newPort()
	cliKvAddr := kvAddr(cliPort)
	cliAddr := "http://" + cliKvAddr

	startedChan, stoppedChan := db.EmbedEtcd(dataDir, peerAddr, cliAddr, shutdownInitiateChan, false)

	cli, err := libkv.NewStore(
		store.ETCD,
		[]string{cliKvAddr},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		panic(err)
	}

	select {
	case <-startedChan:
	case <-stoppedChan:
		panic("Failed to start testing etcd")
	}

	go func() {
		<-stoppedChan
		err := os.RemoveAll(dataDir)
		if err != nil {
			panic(err)
		}
		close(cleanupChan)
	}()

	return cli, shutdownInitiateChan, cleanupChan
}
