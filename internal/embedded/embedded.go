package embedded

import (
	"net/url"
	"time"

	"github.com/coreos/etcd/embed"
	"github.com/coreos/pkg/capnslog"
	"github.com/serverless/event-gateway/internal/sync"
)

// EmbedEtcd starts an embedded etcd instance. It can be shut down by closing the shutdown chan.
// It returns a chan that is closed upon startup, and a chan that is closed once shutdown is complete.
func EmbedEtcd(dataDir, peerAddr, cliAddr string, shutdownGuard *sync.ShutdownGuard) {
	cfg := embed.NewConfig()

	// set advertise urls
	clientURL := parse(cliAddr)
	peerURL := parse(peerAddr)

	// client/peer advertisement addresses
	cfg.ACUrls = []url.URL{*clientURL}
	cfg.APUrls = []url.URL{*peerURL}

	// client/peer listen addresses
	cfg.LCUrls = []url.URL{*clientURL}
	cfg.LPUrls = []url.URL{*peerURL}

	cfg.InitialCluster = "default=" + peerAddr

	// reduce log spam unless in verbose mode
	etcdLogger, err := capnslog.GetRepoLogger("github.com/coreos/etcd")
	if err != nil {
		shutdownGuard.ShutdownAndWait()
		panic(err)
	}
	etcdLogger.SetLogLevel(map[string]capnslog.LogLevel{
		"etcdserver/api":        capnslog.CRITICAL,
		"etcdserver/membership": capnslog.CRITICAL,
		"etcdserver":            capnslog.CRITICAL,
		"raft":                  capnslog.CRITICAL,
		"auth":                  capnslog.CRITICAL,
		"embed":                 capnslog.CRITICAL,
		"wal":                   capnslog.CRITICAL,
	})

	cfg.Dir = dataDir

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		shutdownGuard.ShutdownAndWait()
		panic(err)
	}

	select {
	case <-e.Server.ReadyNotify():
		shutdownGuard.Add(1)
	case <-time.After(60 * time.Second):
		e.Server.Stop()
		shutdownGuard.ShutdownAndWait()
		panic("Embedded etcd took too long to start!")
	}

	// startup or timeout
	go func() {
		// run until error or shutdown
		select {
		case <-shutdownGuard.ShuttingDown:
			e.Server.Stop()
			shutdownGuard.Done()
		case err := <-e.Err():
			e.Server.Stop()
			shutdownGuard.ShutdownAndDone()
			shutdownGuard.Wait()
			panic("etcd failed to start: " + err.Error())
		}
	}()
}

func parse(input string) *url.URL {
	output, err := url.Parse(input)
	if err != nil {
		panic(err)
	}
	return output
}
