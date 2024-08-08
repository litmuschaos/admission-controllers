package main

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/litmuschaos/admission-controller/internal/http"
	"github.com/litmuschaos/admission-controller/internal/webhook"
	"github.com/litmuschaos/admission-controller/pkg/clients"
	"github.com/litmuschaos/admission-controller/pkg/log"
	"github.com/litmuschaos/admission-controller/pkg/utils"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	tlscert      string
	tlskey       string
	port         string
	logFormatter string
	logLevel     string
	kubeConfig   string
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("go version: %s", runtime.Version())
	logrus.Infof("go os/arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("go num cpu: %d", runtime.NumCPU())
	logrus.Infof("go num goroutine: %d", runtime.NumGoroutine())

	// init envs
	if err := utils.InitENV(); err != nil {
		logrus.Fatalf("failed to init envs: %v", err)
	}
}

func main() {
	flag.StringVar(&tlscert, "tlscert", "/etc/certs/tls.crt", "Path to the TLS certificate")
	flag.StringVar(&tlskey, "tlskey", "/etc/certs/tls.key", "Path to the TLS key")
	flag.StringVar(&port, "port", "8443", "The port on which to listen")
	flag.StringVar(&logFormatter, "log-formatter", "json", "Log formatter (text|json)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (trace|debug|info|warn|error|fatal|panic)")
	flag.StringVar(&kubeConfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	// Initialize logger
	log.InitLogger(logFormatter, logLevel)

	// create the kubernetes client
	clients, err := clients.GenerateClientSetFromKubeConfig(kubeConfig)
	if err != nil {
		log.Logger.Errorf("failed to create k8s client: %v", err)
		return
	}

	var tlsCert *tls.Certificate

	// manage the dependencies
	if utils.WebHookFilters.SelfManagedDependencies {
		tlsCert, err = webhook.ManageDependencies(clients)
		if err != nil {
			log.Logger.Errorf("failed to init dependencies: %v", err)
			return
		}
		tlscert, tlskey = "", ""
	} else {
		// check for existing of tls cert and key
		if !utils.IsFileExists(tlscert) || !utils.IsFileExists(tlskey) {
			log.Logger.Errorf("tls cert or key not found")
			return
		}
	}

	server := http.NewServer(port, clients, tlsCert)

	go func() {
		// listen shutdown signal
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-signalChan
		log.Logger.Errorf("Received %s signal; shutting down...", sig)
		if err := server.Shutdown(context.Background()); err != nil {
			log.Logger.Error(err.Error())
		}
	}()

	log.Logger.Infof("Starting server on port: %s", port)
	if err := server.ListenAndServeTLS(tlscert, tlskey); err != nil {
		log.Logger.Errorf("Failed to listen and serve: %v", err)
		os.Exit(1)
	}
}
