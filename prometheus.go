package main

import (
	"net"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const namespace = "fah"

var prevMetrics Metrics

// Exporter is the struct for all metrics
type Exporter struct {
	// Generic info
	up        prometheus.Gauge
	slotCount prometheus.Gauge
	// Slot info
	description *prometheus.GaugeVec
	idle        *prometheus.GaugeVec
	paused      *prometheus.GaugeVec
	reason      *prometheus.GaugeVec
	// Queue info
	eta         *prometheus.GaugeVec
	framesDone  *prometheus.GaugeVec
	percentDone *prometheus.GaugeVec
	ppd         *prometheus.GaugeVec
	qError      *prometheus.GaugeVec
	state       *prometheus.GaugeVec
	totalFrames *prometheus.GaugeVec
}

// Metrics collected metrics
type Metrics struct {
	Slots  []SlotInfo
	Queues []QueueInfo
}

// NewExporter initializes the Exporter struct
func NewExporter() *Exporter {
	return &Exporter{
		// Generic info
		up: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "up",
				Help:      "FAH Metric Collection Operational",
			},
		),
		slotCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "slot_count",
				Help:      "Count of folding slots",
			},
		),
		// Slot info
		description: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "description",
				Help:      "Folding slot description",
			},
			[]string{"slot", "description"},
		),
		reason: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "reason",
				Help:      "Why the slot is paused",
			},
			[]string{"slot", "reason"},
		),
		idle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "idle",
				Help:      "Whether slot is idle",
			},
			[]string{"slot"},
		),
		paused: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "paused",
				Help:      "Whether slot is paused",
			},
			[]string{"slot"},
		),
		// Queue info
		eta: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "eta",
				Help:      "Task ETA",
			},
			[]string{"slot", "queue", "eta"},
		),
		qError: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "queue_error",
				Help:      "Task error",
			},
			[]string{"slot", "queue", "error"},
		),
		state: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "state",
				Help:      "Task state",
			},
			[]string{"slot", "queue", "state"},
		),
		framesDone: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "frames_done",
				Help:      "Task frames done",
			},
			[]string{"slot", "queue"},
		),
		totalFrames: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "total_frames",
				Help:      "Task total frames",
			},
			[]string{"slot", "queue"},
		),
		percentDone: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "percent_done",
				Help:      "Task percent done",
			},
			[]string{"slot", "queue"},
		),
		ppd: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "ppd",
				Help:      "Task points per day",
			},
			[]string{"slot", "queue"},
		),
	}
}

func collectMetrics() (data Metrics, err error) {
	conn, err := net.Dial("tcp", fahAddress)
	if err != nil {
		log.Errorf("Cannot connect to FAH client: %v", err)
		return
	}
	defer conn.Close()
	qInfo, err := ReadQueueInfo(conn)
	if err != nil {
		log.Errorf("Cannot read queue info: %v", err)
		return
	}
	sInfo, err := ReadSlotInfo(conn)
	if err != nil {
		log.Errorf("Cannot read slot info: %v", err)
		return
	}
	data.Queues = qInfo
	data.Slots = sInfo
	return
}

// Collect is called by the Prometheus registry when collecting metrics
func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	data, err := collectMetrics()
	if err != nil {
		log.Errorf("Failed to collect metrics: %s", err)
		e.up.Set(0)
		e.up.Collect(metrics)
		return
	}

	e.up.Set(1)
	e.slotCount.Set(float64(len(data.Slots)))

	// Delete previous data
	// This is done so that we don't keep showing old data
	// when a slot is removed for example
	for _, s := range prevMetrics.Slots {
		e.description.DeleteLabelValues(s.ID, s.Description)
		e.reason.DeleteLabelValues(s.ID, s.Reason)
		e.idle.DeleteLabelValues(s.ID)
		e.paused.DeleteLabelValues(s.ID)
	}
	for _, q := range prevMetrics.Queues {
		e.eta.DeleteLabelValues(q.Slot, q.ID, q.Eta)
		e.qError.DeleteLabelValues(q.Slot, q.ID, q.Error)
		e.state.DeleteLabelValues(q.Slot, q.ID, q.State)
		e.framesDone.DeleteLabelValues(q.Slot, q.ID)
		e.totalFrames.DeleteLabelValues(q.Slot, q.ID)
		e.percentDone.DeleteLabelValues(q.Slot, q.ID)
		e.ppd.DeleteLabelValues(q.Slot, q.ID)
	}

	prevMetrics = data

	// Add collected slot data
	for _, s := range data.Slots {
		e.description.WithLabelValues(s.ID, s.Description).Set(1)
		if len(s.Reason) > 0 {
			e.reason.WithLabelValues(s.ID, s.Reason).Set(1)
		} else {
			e.reason.WithLabelValues(s.ID, s.Reason).Set(0)
		}
		if s.Idle {
			e.idle.WithLabelValues(s.ID).Set(1)
		} else {
			e.idle.WithLabelValues(s.ID).Set(0)
		}
		if s.Options.Paused {
			e.paused.WithLabelValues(s.ID).Set(1)
		} else {
			e.paused.WithLabelValues(s.ID).Set(0)
		}
	}

	// Add collected queue data
	for _, q := range data.Queues {
		e.eta.WithLabelValues(q.Slot, q.ID, q.Eta).Set(1)
		e.qError.WithLabelValues(q.Slot, q.ID, q.Error).Set(1)
		e.state.WithLabelValues(q.Slot, q.ID, q.State).Set(1)
		e.framesDone.WithLabelValues(q.Slot, q.ID).Set(float64(q.FramesDone))
		e.totalFrames.WithLabelValues(q.Slot, q.ID).Set(float64(q.TotalFrames))
		percDone, err := strconv.ParseFloat(strings.TrimSuffix(q.PercentDone, "%"), 64)
		if err != nil {
			log.Debugf("Cannot parse percetange done: %v", err)
			return
		}
		e.percentDone.WithLabelValues(q.Slot, q.ID).Set(percDone)
		ppd, _ := strconv.ParseFloat(q.Ppd, 64)
		if err != nil {
			log.Debugf("Cannot parse ppd: %v", err)
			return
		}
		e.ppd.WithLabelValues(q.Slot, q.ID).Set(ppd)
	}

	e.up.Collect(metrics)
	e.slotCount.Collect(metrics)
	e.description.Collect(metrics)
	e.idle.Collect(metrics)
	e.paused.Collect(metrics)
	e.reason.Collect(metrics)
	e.eta.Collect(metrics)
	e.framesDone.Collect(metrics)
	e.percentDone.Collect(metrics)
	e.ppd.Collect(metrics)
	e.qError.Collect(metrics)
	e.state.Collect(metrics)
	e.totalFrames.Collect(metrics)

}

// Describe sends the super-set of all possible descriptors
func (e *Exporter) Describe(descs chan<- *prometheus.Desc) {
	e.up.Describe(descs)
	e.slotCount.Describe(descs)
	e.description.Describe(descs)
	e.idle.Describe(descs)
	e.paused.Describe(descs)
	e.reason.Describe(descs)
	e.eta.Describe(descs)
	e.framesDone.Describe(descs)
	e.percentDone.Describe(descs)
	e.ppd.Describe(descs)
	e.qError.Describe(descs)
	e.state.Describe(descs)
	e.totalFrames.Describe(descs)
}
