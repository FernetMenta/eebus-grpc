// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/enbility/eebus-grpc/utils/logging"

	"github.com/enbility/eebus-grpc/eebus_service"
	"github.com/enbility/eebus-grpc/rpc_server/control_service_server"
	"github.com/enbility/eebus-grpc/utils/certificates"
)

type app struct {
	eebusService         *eebus_service.Service
	controlServiceServer *control_service_server.Server

	log.Logger
}

func usage() {
	// filepath.Base sorgt dafür, dass nur der Dateiname (z.B. "myapp")
	// ausgegeben wird und nicht der gesamte Pfad (z.B. "./bin/myapp").
	exeName := filepath.Base(os.Args[0])

	fmt.Printf("Usage:\n")
	fmt.Printf("  %s -port=<port> -ipv4Addr=<bind address> -certificate-path=<certificate path> -private-key-path=<private key path>\n", exeName)
}

const (
	OrganizationalUnit = "EVerest"
	Organization       = "EVerest"
	Country            = "DE"
	CommonName         = "EVerest"
)

// interfacesForAddrs returns the names of the local network interfaces that
// carry any of the given comma-separated IP addresses.
func interfacesForAddrs(list string) ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, item := range strings.Split(list, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ip := net.ParseIP(item)
		if ip == nil {
			return nil, fmt.Errorf("%q is not an IP address", item)
		}
		found := false
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				if n, ok := a.(*net.IPNet); ok && n.IP.Equal(ip) {
					names = append(names, iface.Name)
					found = true
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("no local interface has address %s", item)
		}
	}
	return names, nil
}

func main() {
	var rpcPort = flag.Int("port", 50051, "The gRPC server port")
	var rpcIpv4Addr = flag.String("ipv4Addr", "", "ipv4 address to bind the gRPC server to")
	var certificateFilePath = flag.String("certificate-path", "", "The path to the certificate")
	var privateKeyFilePath = flag.String("private-key-path", "", "The path to the private key")
	var logLevel = flag.String("log-level", "info", "Log level: trace, debug, info, error")
	var eebusAddr = flag.String("eebus-addr", "", "Local IP address(es), comma-separated, whose network interface is used for mDNS/SHIP (default: all interfaces)")
	flag.Parse()

	log.SetLevel(log.ParseLevel(*logLevel))
	log.Debugf("Log level set to: %s", *logLevel)

	var app app
	if *certificateFilePath == "" {
		app.Error("No certificate path provided")
		usage()
		os.Exit(1)
	}
	if *privateKeyFilePath == "" {
		app.Error("No private key path provided")
		usage()
		os.Exit(1)
	}
	app.Infof("Certificate path: %s", *certificateFilePath)
	app.Infof("Private key path: %s", *privateKeyFilePath)
	app.Infof("Port: %d", *rpcPort)
	app.Infof("IPv4 Address: %s", *rpcIpv4Addr)

	cert := certificates.SetupCertificate(
		*certificateFilePath,
		*privateKeyFilePath,
		OrganizationalUnit,
		Organization,
		Country,
		CommonName,
	)

	app.eebusService = eebus_service.NewService()
	app.controlServiceServer = control_service_server.NewServer(app.eebusService, &cert)

	if *eebusAddr != "" {
		ifaces, err := interfacesForAddrs(*eebusAddr)
		if err != nil {
			app.Errorf("Invalid eebus-addr: %v", err)
			os.Exit(1)
		}
		app.Infof("EEBUS address %s -> network interfaces: %v", *eebusAddr, ifaces)
		app.controlServiceServer.SetInterfaces(ifaces)
	}
	app.controlServiceServer.Start(*rpcIpv4Addr, rpcPort)

	// Clean exit to make sure mdns shutdown is invoked
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	// User exit

	app.controlServiceServer.Stop()
}
