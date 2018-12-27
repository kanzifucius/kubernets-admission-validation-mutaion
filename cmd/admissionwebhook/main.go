package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"vod-ms-kubernetes-admission-webhook/pkg/config"
	"vod-ms-kubernetes-admission-webhook/pkg/webhook"
)

func check(err error, message string) {
	if err != nil {
		glog.Errorf("error: %v %v", err, message)
	}
}

// Webhook Server parameters
type WhSvrParameters struct {
	port         int    // webhook server port
	certFile     string // path to the x509 certificate for https
	keyFile      string // path to the x509 private key matching `CertFile`
	configFolder string
	configName   string
}

func main() {
	var parameters WhSvrParameters

	// get command line parameters
	flag.IntVar(&parameters.port, "port", 443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.configFolder, "configFolder", "/etc/config/", "folder of config location")
	flag.StringVar(&parameters.configName, "configName", "config", "nameof config file no extention")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
	check(err, "Failed to Load cert")

	Laodedconfiguration := config.LoadConfig(parameters.configFolder, parameters.configName)
	whsvr := &webhook.WHServer{
		Configuration: Laodedconfiguration,
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.Serve)
	mux.HandleFunc("/validate", whsvr.Serve)
	whsvr.Server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := whsvr.Server.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	glog.Info("Server started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsvr.Server.Shutdown(context.Background())
}
