package main

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const namespace = "fah"

var (
	prevMetrics Metrics
	lastUpdate  time.Time
)

// Exporter is the struct for all metrics
type Exporter struct {
	// Generic info
	up        prometheus.Gauge
	slotCount prometheus.Gauge
	options   *prometheus.GaugeVec
	// Slot info
	description *prometheus.GaugeVec
	idle        *prometheus.GaugeVec
	paused      *prometheus.GaugeVec
	// Queue info
	framesDone  *prometheus.GaugeVec
	totalFrames *prometheus.GaugeVec
	percentDone *prometheus.GaugeVec
	ppd         *prometheus.GaugeVec
	queueInfo   *prometheus.GaugeVec
	// Donor API
	donorCredit     *prometheus.GaugeVec
	donorID         *prometheus.GaugeVec
	donorRank       *prometheus.GaugeVec
	donorTeamCredit *prometheus.GaugeVec
}

// Metrics collected metrics
type Metrics struct {
	Slots   []SlotInfo
	Queues  []QueueInfo
	Options Options
	Donor   DonorAPI
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
		options: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "options",
				Help:      "Client options",
			},
			[]string{"user", "team", "power"},
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
			[]string{"slot", "reason"},
		),
		// Queue info
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
		queueInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "queue_info",
				Help:      "Task state, ETA and eventual error",
			},
			[]string{"slot", "queue", "state", "eta", "error"},
		),
		// Donor API
		donorCredit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "donor_credit",
				Help:      "Donor total credit",
			},
			[]string{"user"},
		),
		donorID: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "donor_id",
				Help:      "Donor user ID",
			},
			[]string{"user"},
		),
		donorRank: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "donor_rank",
				Help:      "Donor rank",
			},
			[]string{"user"},
		),
		donorTeamCredit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "donor_team_credit",
				Help:      "Donor credit per team",
			},
			[]string{"user", "name", "team"},
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
	err = ReadFAH(conn, "queue-info", &data.Queues)
	if err != nil {
		log.Errorf("Cannot read queue info: %v", err)
		return
	}
	err = ReadFAH(conn, "slot-info", &data.Slots)
	if err != nil {
		log.Errorf("Cannot read slot info: %v", err)
		return
	}
	err = ReadFAH(conn, "options", &data.Options)
	if err != nil {
		log.Errorf("Cannot read options: %v", err)
		return
	}
	if getAPI {
		if time.Since(lastUpdate).Seconds() > apiThrottle.Seconds() {
			log.Debugf("Getting donor API data")
			err = ReadAPI("/donor/"+data.Options.User, &data.Donor)
			if err != nil {
				log.Errorf("Cannot get donor info from API: %v", err)
			}
			lastUpdate = time.Now()
		} else {
			data.Donor = prevMetrics.Donor
		}
	}
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
	e.options.DeleteLabelValues(prevMetrics.Options.User, prevMetrics.Options.Team, data.Options.Power)
	for _, s := range prevMetrics.Slots {
		e.description.DeleteLabelValues(s.ID, s.Description)
		e.idle.DeleteLabelValues(s.ID)
		e.paused.DeleteLabelValues(s.ID, s.Reason)
	}
	for _, q := range prevMetrics.Queues {
		e.framesDone.DeleteLabelValues(q.Slot, q.ID)
		e.totalFrames.DeleteLabelValues(q.Slot, q.ID)
		e.percentDone.DeleteLabelValues(q.Slot, q.ID)
		e.ppd.DeleteLabelValues(q.Slot, q.ID)
		e.queueInfo.DeleteLabelValues(q.Slot, q.ID, q.State, q.Eta, q.Error)
	}

	e.options.WithLabelValues(data.Options.User, data.Options.Team, data.Options.Power).Set(1)

	// Add collected slot data
	for _, s := range data.Slots {
		e.description.WithLabelValues(s.ID, s.Description).Set(1)
		if s.Idle {
			e.idle.WithLabelValues(s.ID).Set(1)
		} else {
			e.idle.WithLabelValues(s.ID).Set(0)
		}
		if s.Options.Paused {
			e.paused.WithLabelValues(s.ID, s.Reason).Set(1)
		} else {
			e.paused.WithLabelValues(s.ID, s.Reason).Set(0)
		}
	}

	// Add collected queue data
	for _, q := range data.Queues {
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
		e.queueInfo.WithLabelValues(q.Slot, q.ID, q.State, q.Eta, q.Error).Set(1)
	}

	e.up.Collect(metrics)
	e.slotCount.Collect(metrics)
	e.options.Collect(metrics)
	e.description.Collect(metrics)
	e.idle.Collect(metrics)
	e.paused.Collect(metrics)
	e.framesDone.Collect(metrics)
	e.totalFrames.Collect(metrics)
	e.percentDone.Collect(metrics)
	e.ppd.Collect(metrics)
	e.queueInfo.Collect(metrics)

	if getAPI {
		e.donorCredit.DeleteLabelValues(data.Donor.Name)
		e.donorID.DeleteLabelValues(data.Donor.Name)
		e.donorRank.DeleteLabelValues(data.Donor.Name)
		for _, t := range prevMetrics.Donor.Teams {
			e.donorTeamCredit.DeleteLabelValues(prevMetrics.Donor.Name, t.Name, strconv.Itoa(t.Team))
		}
		e.donorCredit.WithLabelValues(data.Donor.Name).Set(float64(data.Donor.Credit))
		e.donorID.WithLabelValues(data.Donor.Name).Set(float64(data.Donor.ID))
		e.donorRank.WithLabelValues(data.Donor.Name).Set(float64(data.Donor.Rank))
		for _, t := range data.Donor.Teams {
			e.donorTeamCredit.WithLabelValues(data.Donor.Name, t.Name, strconv.Itoa(t.Team)).Set(float64(t.Credit))
		}
		e.donorCredit.Collect(metrics)
		e.donorID.Collect(metrics)
		e.donorRank.Collect(metrics)
		e.donorTeamCredit.Collect(metrics)
	}

	prevMetrics = data
}

// Describe sends the super-set of all possible descriptors
func (e *Exporter) Describe(descs chan<- *prometheus.Desc) {
	e.up.Describe(descs)
	e.slotCount.Describe(descs)
	e.options.Describe(descs)
	e.description.Describe(descs)
	e.idle.Describe(descs)
	e.paused.Describe(descs)
	e.framesDone.Describe(descs)
	e.totalFrames.Describe(descs)
	e.percentDone.Describe(descs)
	e.ppd.Describe(descs)
	e.queueInfo.Describe(descs)

	if getAPI {
		e.donorCredit.Describe(descs)
		e.donorID.Describe(descs)
		e.donorRank.Describe(descs)
		e.donorTeamCredit.Describe(descs)
	}
}
