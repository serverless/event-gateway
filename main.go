package main

import (
	"flag"
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

func init() {
	etcd.Register()
}

func main() {
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	dbHosts := flag.String("db-hosts", "localhost:2379", "Comma-separated list of database hosts to connect to.")
	embedMaster := flag.Bool("dev", false, "Run embedded etcd for testing.")
	embedPeerAddr := flag.String("embed-peer-addr", "http://localhost:2380", "Address for testing embedded etcd to receive peer connections.")
	embedCliAddr := flag.String("embed-cli-addr", "http://localhost:2379", "Address for testing embedded etcd to receive client connections.")
	embedDataDir := flag.String("embed-data-dir", "default.etcd", "Path for testing embedded etcd to store its state.")
	apiPort := flag.Uint("api-port", 8081, "Port to serve configuration API on.")
	apiTLSCrt := flag.String("api-tls-cert", "", "Path to API TLS certificate file.")
	apiTLSKey := flag.String("api-tls-key", "", "Path to API TLS key file.")
	gatewayTLSCrt := flag.String("gateway-tls-cert", "", "Path to gateway TLS certificate file.")
	gatewayTLSKey := flag.String("gateway-tls-key", "", "Path to gateway TLS key file.")
	gatewayPort := flag.Uint("gateway-port", 8080, "Port to serve configured endpoints on.")
	flag.Parse()

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
	httplisteners.StartAPI(httplisteners.Config{
		KV:            kv,
		Log:           log,
		TLSCrt:        apiTLSCrt,
		TLSKey:        apiTLSKey,
		Port:          *apiPort,
		ShutdownGuard: shutdownGuard,
	})

	// start Event Gateway handler
	httplisteners.StartGateway(httplisteners.Config{
		KV:            kv,
		Log:           log,
		TLSCrt:        gatewayTLSCrt,
		TLSKey:        gatewayTLSKey,
		Port:          *gatewayPort,
		ShutdownGuard: shutdownGuard,
	})

	shutdownGuard.Wait()
}
