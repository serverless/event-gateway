package tests

import (
	"os"
	"strconv"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/util"
)

func init() {
	etcd.Register()
}

func kvAddr(port int) string {
	return "127.0.0.1:" + strconv.Itoa(port)
}

func TestingEtcd() (store.Store, *util.ShutdownGuard) {
	shutdownGuard := util.NewShutdownGuard()

	wd, err := os.Getwd()
	if err != nil {
		shutdownGuard.ShutdownAndWait()
		panic(err)
	}

	peerPort := newPort()
	peerAddr := "http://localhost:" + strconv.Itoa(peerPort)

	etcdDir := "testing.etcd"
	dataDir := wd + "/" + etcdDir + "." + strconv.Itoa(peerPort)

	cliPort := newPort()
	cliKvAddr := kvAddr(cliPort)
	cliAddr := "http://" + cliKvAddr

	db.EmbedEtcd(dataDir, peerAddr, cliAddr, shutdownGuard, false)

	cli, err := libkv.NewStore(
		store.ETCD,
		[]string{cliKvAddr},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		shutdownGuard.ShutdownAndWait()
		panic(err)
	}

	go func() {
		shutdownGuard.Add(1)
		<-shutdownGuard.ShuttingDown
		err := os.RemoveAll(dataDir)
		shutdownGuard.Done()
		if err != nil {
			panic(err)
		}
	}()

	return cli, shutdownGuard
}
