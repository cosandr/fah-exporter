package main

import (
	"flag"
	"net"
	"net/http"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	defaultFahAddress = "127.0.0.1:36330"
)

var (
	fahAddress = defaultFahAddress
)

func main() {
	var (
		level          string
		listenAddress  string
		metricsPath    string
		socketActivate bool
		noTimestamps   bool
	)

	flag.StringVar(&level, "log.level", "info", "Set the output log level")
	flag.BoolVar(&noTimestamps, "log.no-timestamps", false, "Disable logging timestamps, true when using systemd activation")
	flag.StringVar(&listenAddress, "web.listen-address", "0.0.0.0:9659", "Address to listen on for web interface and telemetry.")
	flag.StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	flag.BoolVar(&socketActivate, "systemd", false, "Run using systemd socket activation")
	flag.StringVar(&fahAddress, "fah.address", defaultFahAddress, "Listen address of FAH client")
	flag.Parse()
	setLogLevel(level)

	if noTimestamps || socketActivate {
		log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	} else {
		log.SetFormatter(&log.TextFormatter{DisableTimestamp: false, FullTimestamp: true})
	}

	prometheus.MustRegister(NewExporter())

	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>FAH Exporter</title></head>
             <body>
             <h1>FAH Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
	     <h2>More information:</h2>
	     <p><a href="https://github.com/cosandr/fah-exporter">github.com/cosandr/fah-exporter</a></p>
             </body>
             </html>`))
	})

	var listener net.Listener
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

	log.Infof("FAH client address: %s", fahAddress)
	log.Infof("Starting HTTP server on %s", listener.Addr().String())
	log.Fatal(http.Serve(listener, nil))
}

func setLogLevel(level string) {
	switch level {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.Warnln("Unrecognized minimum log level; using 'info' as default")
	}
}
