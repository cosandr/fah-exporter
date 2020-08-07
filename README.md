[![Go Report Card](https://goreportcard.com/badge/github.com/cosandr/fah-exporter)](https://goreportcard.com/report/github.com/cosandr/fah-exporter) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/cosandr/fah-exporter/blob/master/LICENSE)

# Prometheus exporter for the Folding@home client

The Folding@home client/service must be running, you can check this by running `telnet localhost 36330`.
Running the binary without arguments will expose the metrics at `http://localhost:9659/metrics`.

The [setup script](./setup.sh) can be used to build and install the binary and/or systemd services.

## Grafana dashboard

A [sample dashboard](dashboards/fah.json) is provided.

![grafana-dash](https://user-images.githubusercontent.com/56652361/89687564-bf3e9b00-d900-11ea-8e89-eb15832ce575.png)

## Metrics example

```
# HELP fah_description Folding slot description
# TYPE fah_description gauge
fah_description{description="gpu:0:GP102 [GeForce GTX 1080 Ti] 11380",slot="01"} 1
# HELP fah_eta Task ETA
# TYPE fah_eta gauge
fah_eta{eta="59 mins 45 secs",queue="00",slot="01"} 1
# HELP fah_frames_done Task frames done
# TYPE fah_frames_done gauge
fah_frames_done{queue="00",slot="01"} 63
# HELP fah_idle Whether slot is idle
# TYPE fah_idle gauge
fah_idle{slot="01"} 0
# HELP fah_paused Whether slot is paused
# TYPE fah_paused gauge
fah_paused{slot="01"} 0
# HELP fah_percent_done Task percent done
# TYPE fah_percent_done gauge
fah_percent_done{queue="00",slot="01"} 63.04
# HELP fah_ppd Task points per day
# TYPE fah_ppd gauge
fah_ppd{queue="00",slot="01"} 1.966926e+06
# HELP fah_queue_error Task error
# TYPE fah_queue_error gauge
fah_queue_error{error="NO_ERROR",queue="00",slot="01"} 1
# HELP fah_reason Why the slot is idle
# TYPE fah_reason gauge
fah_reason{reason="",slot="01"} 1
# HELP fah_slot_count Count of folding slots
# TYPE fah_slot_count gauge
fah_slot_count 1
# HELP fah_state Task state
# TYPE fah_state gauge
fah_state{queue="00",slot="01",state="RUNNING"} 1
# HELP fah_total_frames Task total frames
# TYPE fah_total_frames gauge
fah_total_frames{queue="00",slot="01"} 100
# HELP fah_up FAH Metric Collection Operational
# TYPE fah_up gauge
fah_up 1
```
