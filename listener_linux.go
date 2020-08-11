package main

import (
	"net"

	"github.com/coreos/go-systemd/v22/activation"
	log "github.com/sirupsen/logrus"
)

func getListener(socketActivate bool, listenAddress string) (listener net.Listener) {
	if socketActivate {
		listeners, err := activation.Listeners()
		if err != nil {
			log.Panic(err)
		}

		if len(listeners) != 1 {
			log.Panic("Unexpected number of socket activation fds")
		}
		listener = listeners[0]
	} else {
		var err error
		listener, err = net.Listen("tcp", listenAddress)
		if err != nil {
			log.Panicf("Cannot listen: %s", err)
		}
	}
	return listener
}
