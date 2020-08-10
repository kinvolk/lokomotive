// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"

	controller "github.com/kinvolk/lokomotive/internal/admission-webhook-server"
)

const port = "8080"

func usage() {
	flag.PrintDefaults()
	os.Exit(0)
}

func returnError(msg string, err error) {
	glog.Fatalf("%s: %v", msg, err)
}

func main() {
	var tlscert, tlskey string

	flag.Usage = usage

	flag.StringVar(&tlscert, "tlsCertFile", "/etc/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&tlskey, "tlsKeyFile", "/etc/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	certs, err := tls.LoadX509KeyPair(tlscert, tlskey)
	if err != nil {
		returnError("loading key pair failed", err)
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{certs}, MinVersion: tls.VersionTLS13},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", controller.ServeMutateServiceAccount)
	server.Handler = mux

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			returnError("listening and serving webhook server failed", err)
		}
	}()

	glog.Infof("Server running in port: %s", port)

	// listening shutdown signal.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Info("Got shutdown signal, shutting down webhook server gracefully...")
	glog.Flush()

	err = server.Shutdown(context.Background())
	if err != nil {
		returnError("shutting down server failed", err)
	}
}
