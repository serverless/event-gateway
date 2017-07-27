package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/httplisteners"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/util"
)

var version = "master"

func init() {
	etcd.Register()
}

func main() {
	showVersion := flag.Bool("version", false, "Show version.")
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	dbHosts := flag.String("db-hosts", "127.0.0.1:2379", "Comma-separated list of database hosts to connect to.")
	embedMaster := flag.Bool("dev", false, "Run embedded etcd for testing.")
	embedPeerAddr := flag.String("embed-peer-addr", "http://127.0.0.1:2380", "Address for testing embedded etcd to receive peer connections.")
	embedCliAddr := flag.String("embed-cli-addr", "http://127.0.0.1:2379", "Address for testing embedded etcd to receive client connections.")
	embedDataDir := flag.String("embed-data-dir", "default.etcd", "Path for testing embedded etcd to store its state.")
	configPort := flag.Uint("config-port", 4001, "Port to serve configuration API on.")
	configTLSCrt := flag.String("config-tls-cert", "", "Path to configuration API TLS certificate file.")
	configTLSKey := flag.String("config-tls-key", "", "Path to configuration API TLS key file.")
	eventsPort := flag.Uint("events-port", 4000, "Port to serve events API on.")
	eventsTLSCrt := flag.String("events-tls-cert", "", "Path to events API TLS certificate file.")
	eventsTLSKey := flag.String("events-tls-key", "", "Path to events API TLS key file.")
	flag.Parse()

	if *showVersion {
		println(version)
		os.Exit(0)
	}

	dbHostStrings := strings.Split(*dbHosts, ",")

	prometheus.MustRegister(metrics.RequestDuration)
	prometheus.MustRegister(metrics.DroppedPubSubEvents)

	logCfg := zap.NewDevelopmentConfig()
	if !*verbose {
		logCfg = zap.NewProductionConfig()
		logCfg.DisableStacktrace = true
	}

	log, err := logCfg.Build()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	shutdownGuard := util.NewShutdownGuard()

	if *embedMaster {
		db.EmbedEtcd(*embedDataDir, *embedPeerAddr, *embedCliAddr, shutdownGuard, *verbose)
	}

	kv, err := libkv.NewStore(
		store.ETCD,
		dbHostStrings,
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("Cannot create kv client.", zap.Error(err))
	}

	// start API handler
	httplisteners.StartConfigAPI(httplisteners.Config{
		KV:            kv,
		Log:           log,
		TLSCrt:        configTLSCrt,
		TLSKey:        configTLSKey,
		Port:          *configPort,
		ShutdownGuard: shutdownGuard,
	})

	// start Event Gateway handler
	httplisteners.StartEventsAPI(httplisteners.Config{
		KV:            kv,
		Log:           log,
		TLSCrt:        eventsTLSCrt,
		TLSKey:        eventsTLSKey,
		Port:          *eventsPort,
		ShutdownGuard: shutdownGuard,
	})

	shutdownGuard.Wait()
}
