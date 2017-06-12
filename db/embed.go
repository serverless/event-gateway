package db

import (
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/coreos/etcd/embed"
	"github.com/coreos/pkg/capnslog"
)

// EmbedEtcd starts an embedded etcd instance. It can be shut down by closing the shutdown chan.
// It returns a chan that is closed upon startup, and a chan that is closed once shutdown is complete.
func EmbedEtcd(dataDir, peerAddr, cliAddr string, shutdown chan struct{}, log *zap.Logger, verboseLogging bool) (<-chan struct{}, <-chan struct{}) {
	startedChan := make(chan struct{})
	stoppedChan := make(chan struct{})

	cfg := embed.NewConfig()

	// set advertise urls
	clientURL, err := url.Parse(cliAddr)
	if err != nil {
		panic(err)
	}

	peerURL, err := url.Parse(peerAddr)
	if err != nil {
		panic(err)
	}

	// client/peer advertisement addresses
	cfg.ACUrls = []url.URL{*clientURL}
	cfg.APUrls = []url.URL{*peerURL}

	// client/peer listen addresses
	cfg.LCUrls = []url.URL{*clientURL}
	cfg.LPUrls = []url.URL{*peerURL}

	cfg.InitialCluster = "default=" + peerAddr

	// reduce log spam unless in verbose mode
	if !verboseLogging {
		etcdLogger, err := capnslog.GetRepoLogger("github.com/coreos/etcd")
		if err != nil {
			panic(err)
		}
		etcdLogger.SetLogLevel(map[string]capnslog.LogLevel{
			"etcdserver/api":        capnslog.CRITICAL,
			"etcdserver/membership": capnslog.CRITICAL,
			"etcdserver":            capnslog.CRITICAL,
			"raft":                  capnslog.CRITICAL,
		})
	}

	cfg.Dir = dataDir

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		panic(err)
	}

	// startup or timeout
	go func() {
		select {
		case <-e.Server.ReadyNotify():
			if log != nil {
				log.Info("Embedded etcd is ready.")
			}
			close(startedChan)
		case <-time.After(60 * time.Second):
			log.Error("Embedded etcd took too long to start!")
			e.Server.Stop()
			close(stoppedChan)
			return
		}

		// run until error or shutdown
		select {
		case <-shutdown:
			if log != nil {
				log.Info("Shutting down embedded etcd.")
			}
			e.Server.Stop()
			close(stoppedChan)
		case err := <-e.Err():
			e.Server.Stop()
			close(stoppedChan)
			log.Error("Etcd failed to start.", zap.Error(err))
		}
	}()

	return startedChan, stoppedChan
}
