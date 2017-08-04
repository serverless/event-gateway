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
	"go.uber.org/zap/zapcore"

	"github.com/serverless/event-gateway/api"
	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/util"
	"github.com/serverless/event-gateway/util/httpapi"
)

var version = "dev"

func init() {
	etcd.Register()
}

func main() {
	showVersion := flag.Bool("version", false, "Show version.")
	logLevel := zap.LevelFlag("log-level", zap.InfoLevel, `The level of logging to show after the event gateway has started. The available log levels are "debug", "info", "warn", and "err".`)
	dbHosts := flag.String("db-hosts", "127.0.0.1:2379", "Comma-separated list of database hosts to connect to.")
	developmentMode := flag.Bool("dev", false, "Run embedded etcd for testing.")
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

	prometheus.MustRegister(metrics.RequestDuration)
	prometheus.MustRegister(metrics.DroppedPubSubEvents)

	logCfg := zap.NewProductionConfig()
	logCfg.Level = zap.NewAtomicLevelAt(*logLevel)
	if *developmentMode {
		logCfg = zap.NewDevelopmentConfig()
		logCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
		logCfg.DisableCaller = true
		logCfg.DisableStacktrace = true
		logCfg.Level = zap.NewAtomicLevelAt(*logLevel)
	}
	log, err := logCfg.Build()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	shutdownGuard := util.NewShutdownGuard()

	if *developmentMode {
		db.EmbedEtcd(*embedDataDir, *embedPeerAddr, *embedCliAddr, shutdownGuard)
		log.Info("Running in development mode with embedded etcd.")
	}

	dbHostStrings := strings.Split(*dbHosts, ",")
	kv, err := libkv.NewStore(
		store.ETCD,
		dbHostStrings,
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("Cannot create KV client.", zap.Error(err))
	}

	// start API handler
	api.StartConfigAPI(httpapi.Config{
		KV:            kv,
		Log:           log,
		TLSCrt:        configTLSCrt,
		TLSKey:        configTLSKey,
		Port:          *configPort,
		ShutdownGuard: shutdownGuard,
	})

	// start Event Gateway handler
	api.StartEventsAPI(httpapi.Config{
		KV:            kv,
		Log:           log,
		TLSCrt:        eventsTLSCrt,
		TLSKey:        eventsTLSKey,
		Port:          *eventsPort,
		ShutdownGuard: shutdownGuard,
	})

	shutdownGuard.Wait()
}
