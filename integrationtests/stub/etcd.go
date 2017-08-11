package stub

import (
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"

	"github.com/serverless/event-gateway/internal/kv"
	"github.com/serverless/event-gateway/internal/sync"
)

func init() {
	etcd.Register()
}

// TestEtcd returns etcd store for testing.
func TestEtcd() (store.Store, *sync.ShutdownGuard) {
	shutdownGuard := sync.NewShutdownGuard()

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

	kv.EmbedEtcd(dataDir, peerAddr, cliAddr, shutdownGuard)

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

func kvAddr(port int) string {
	return "127.0.0.1:" + strconv.Itoa(port)
}

var (
	testPortAllocator = uint32(3370)
)

func newPort() int {
	return int(atomic.AddUint32(&testPortAllocator, 1))
}
