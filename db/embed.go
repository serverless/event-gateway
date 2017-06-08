package db

import (
	"log"
	"net/url"
	"time"

	"github.com/coreos/etcd/embed"
)

// EmbedEtcd starts an embedded etcd instance. It can be shut down by closing the shutdown chan.
// It returns a chan that is closed upon startup, and a chan that is closed once shutdown is complete.
func EmbedEtcd(dataDir, peerAddr, cliAddr string, shutdown chan struct{}) (<-chan struct{}, <-chan struct{}) {
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

	cfg.Dir = dataDir

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		// startup or timeout
		select {
		case <-e.Server.ReadyNotify():
			log.Printf("Embedded etcd is ready!")
			close(startedChan)
		case <-time.After(60 * time.Second):
			log.Printf("Embedded etcd took too long to start!")
			e.Server.Stop()
			close(stoppedChan)
			return
		}

		// run until error or shutdown
		select {
		case <-shutdown:
			log.Printf("Shutting down embedded etcd.")
			e.Server.Stop()
			close(stoppedChan)
		case err := <-e.Err():
			e.Server.Stop()
			close(stoppedChan)
			log.Fatal(err)
		}
	}()

	return startedChan, stoppedChan
}
