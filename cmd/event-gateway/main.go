package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/serverless/event-gateway/router"
	"github.com/serverless/libkv"
	"github.com/serverless/libkv/store"
	etcd "github.com/serverless/libkv/store/etcd/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/serverless/event-gateway/httpapi"
	"github.com/serverless/event-gateway/internal/cache"
	"github.com/serverless/event-gateway/internal/embedded"
	intstore "github.com/serverless/event-gateway/internal/store"
	"github.com/serverless/event-gateway/internal/sync"
	eventgateway "github.com/serverless/event-gateway/libkv"
	"github.com/serverless/event-gateway/plugin"
)

var version = "dev"

func init() {
	etcd.Register()
}

// nolint: gocyclo
func main() {
	showVersion := flag.Bool("version", false, "Show version.")
	logLevel := zap.LevelFlag("log-level", zap.InfoLevel, `The level of logging to show after the event gateway has started. The available log levels are "debug", "info", "warn", and "err".`)
	logFormat := flag.String("log-format", "", `The format of logs. The available formats are "text", "json".`)
	dbHosts := flag.String("db-hosts", "127.0.0.1:2379", "Comma-separated list of database hosts to connect to.")
	developmentMode := flag.Bool("dev", false, `Run in development mode with embedded etcd and "text" log format.`)
	embedPeerAddr := flag.String("embed-peer-addr", "http://127.0.0.1:2380", "Address for testing embedded etcd to receive peer connections.")
	embedCliAddr := flag.String("embed-cli-addr", "http://127.0.0.1:2379", "Address for testing embedded etcd to receive client connections.")
	embedDataDir := flag.String("embed-data-dir", "default.etcd", "Path for testing embedded etcd to store its state.")
	configPort := flag.Uint("config-port", 4001, "Port to serve configuration API on.")
	configTLSCrt := flag.String("config-tls-cert", "", "Path to configuration API TLS certificate file.")
	configTLSKey := flag.String("config-tls-key", "", "Path to configuration API TLS key file.")
	eventsPort := flag.Uint("events-port", 4000, "Port to serve events API on.")
	eventsTLSCrt := flag.String("events-tls-cert", "", "Path to events API TLS certificate file.")
	eventsTLSKey := flag.String("events-tls-key", "", "Path to events API TLS key file.")
	workersNumber := flag.Uint("workers", 100, "Number of workers processing incoming events.")
	workersBacklog := flag.Uint("workers-backlog", 200, "Length of workers backlog. Maximum number of events that wait for processing.")
	plugins := paths{}
	flag.Var(&plugins, "plugin", "Path to a plugin to load.")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Event Gateway version: %s\n", version)
		os.Exit(0)
	}

	log, err := logger(*developmentMode, *logLevel, *logFormat).Build()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	shutdownGuard := sync.NewShutdownGuard()

	if *developmentMode {
		embedded.EmbedEtcd(*embedDataDir, *embedPeerAddr, *embedCliAddr, shutdownGuard)
	}

	// KV store
	kvstore, err := libkv.NewStore(
		store.ETCDV3,
		strings.Split(*dbHosts, ","),
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("Cannot create KV client.", zap.Error(err))
	}

	// Implementation of function and subscription services
	service := &eventgateway.Service{
		FunctionStore:     intstore.NewPrefixed("/serverless-event-gateway/functions", kvstore),
		SubscriptionStore: intstore.NewPrefixed("/serverless-event-gateway/subscriptions", kvstore),
		EndpointStore:     intstore.NewPrefixed("/serverless-event-gateway/endpoints", kvstore),
		Log:               log,
	}

	// Plugin manager
	pluginManager := plugin.NewManager(plugins, log)
	err = pluginManager.Connect()
	if err != nil {
		log.Fatal("Loading plugins failed.", zap.Error(err))
	}

	// Router
	targetCache := cache.NewTarget("/serverless-event-gateway", kvstore, log)
	router := router.New(*workersNumber, *workersBacklog, targetCache, pluginManager, log)
	router.StartWorkers()

	httpapi.StartEventsAPI(router, httpapi.ServerConfig{
		TLSCrt:        eventsTLSCrt,
		TLSKey:        eventsTLSKey,
		Port:          *eventsPort,
		ShutdownGuard: shutdownGuard,
	})

	httpapi.StartConfigAPI(service, service, httpapi.ServerConfig{
		TLSCrt:        configTLSCrt,
		TLSKey:        configTLSKey,
		Port:          *configPort,
		ShutdownGuard: shutdownGuard,
	})

	if *developmentMode {
		eventProto := "http"
		if *eventsTLSCrt != "" && *eventsTLSKey != "" {
			eventProto = "https"
		}
		configProto := "http"
		if *configTLSCrt != "" && *configTLSKey != "" {
			configProto = "https"
		}

		log.Info(fmt.Sprintf("Running in development mode with embedded etcd. Events API listening on %s://localhost:%d. Config API listening on %s://localhost:%d.", eventProto, *eventsPort, configProto, *configPort))
	}

	shutdownGuard.Wait()
	router.Drain()

	if pluginManager != nil {
		pluginManager.Kill()
	}
}

const (
	consoleEncoding = "console"
	jsonEncoding    = "json"
)

func logger(dev bool, level zapcore.Level, format string) zap.Config {
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Sampling:         &zap.SamplingConfig{Initial: 100, Thereafter: 100},
		Encoding:         "json",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if dev {
		cfg.Sampling = nil
		cfg.Encoding = consoleEncoding
	}

	if format != "" {
		if format == "text" {
			cfg.Encoding = consoleEncoding
		} else if format == jsonEncoding {
			cfg.Encoding = jsonEncoding
		} else {
			cfg.Encoding = ""
		}
	}

	if cfg.Encoding == jsonEncoding {
		cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	}

	if cfg.Encoding == consoleEncoding {
		cfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	cfg.DisableCaller = true
	cfg.DisableStacktrace = true

	return cfg
}

type paths []string

func (p *paths) String() string {
	return strings.Join(*p, ",")
}

func (p *paths) Set(value string) error {
	*p = append(*p, value)
	return nil
}
