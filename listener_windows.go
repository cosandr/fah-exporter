package main

import (
	"net"

	log "github.com/sirupsen/logrus"
)

func getListener(socketActivate bool, listenAddress string) (listener net.Listener) {
	var err error
	listener, err = net.Listen("tcp", listenAddress)
	if err != nil {
		log.Panicf("Cannot listen: %s", err)
	}
	return listener
}
