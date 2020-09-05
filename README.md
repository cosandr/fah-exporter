[![Go Report Card](https://goreportcard.com/badge/github.com/cosandr/fah-exporter)](https://goreportcard.com/report/github.com/cosandr/fah-exporter) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/cosandr/fah-exporter/blob/master/LICENSE)

# Prometheus exporter for the Folding@home client

The Folding@home client/service must be running, you can check this by running `telnet localhost 36330`.
Running the binary without arguments will expose the metrics at `http://localhost:9659/metrics`.

The [setup script](./setup.sh) can be used to build and install the binary and/or systemd services.

Optionally fetch data from FAH API (`-fah.api` option) for donor stats, the username is read from the FAH client.

## Grafana dashboard

A [sample dashboard](dashboards/fah.json) is provided.

![grafana-dash](https://user-images.githubusercontent.com/56652361/92312051-76d8d280-efbd-11ea-804e-3f18989b4dfd.png)

## Metrics example

```
# HELP fah_description Folding slot description
# TYPE fah_description gauge
fah_description{description="gpu:0:GP102 [GeForce GTX 1080 Ti] 11380",slot="01"} 1
# HELP fah_donor_credit Donor total credit
# TYPE fah_donor_credit gauge
fah_donor_credit{user="<user>"} 2.67209435e+08
# HELP fah_donor_id Donor user ID
# TYPE fah_donor_id gauge
fah_donor_id{user="<user>"} <value>
# HELP fah_donor_rank Donor rank
# TYPE fah_donor_rank gauge
fah_donor_rank{user="<user>"} 3844
# HELP fah_donor_team_credit Donor credit per team
# TYPE fah_donor_team_credit gauge
fah_donor_team_credit{name="Default (No team specified)",team="0",user="<user>"} 99805
fah_donor_team_credit{name="<team_name>",team="<team_id>",user="<user>"} 6.448344e+06
fah_donor_team_credit{name="<team_name>",team="<team_id>",user="<user>"} 2.60661286e+08
# HELP fah_frames_done Task frames done
# TYPE fah_frames_done gauge
fah_frames_done{queue="01",slot="01"} 16
# HELP fah_idle Whether slot is idle
# TYPE fah_idle gauge
fah_idle{slot="01"} 0
# HELP fah_options Client options
# TYPE fah_options gauge
fah_options{power="full",team="<team_id>",user="<user>"} 1
# HELP fah_paused Whether slot is paused
# TYPE fah_paused gauge
fah_paused{reason="",slot="01"} 0
# HELP fah_percent_done Task percent done
# TYPE fah_percent_done gauge
fah_percent_done{queue="01",slot="01"} 16.53
# HELP fah_ppd Task points per day
# TYPE fah_ppd gauge
fah_ppd{queue="01",slot="01"} 1.683157e+06
# HELP fah_queue_info Task state, ETA and eventual error
# TYPE fah_queue_info gauge
fah_queue_info{error="NO_ERROR",eta="2 hours 03 mins",queue="01",slot="01",state="RUNNING"} 1
# HELP fah_slot_count Count of folding slots
# TYPE fah_slot_count gauge
fah_slot_count 1
# HELP fah_total_frames Task total frames
# TYPE fah_total_frames gauge
fah_total_frames{queue="01",slot="01"} 100
# HELP fah_up FAH Metric Collection Operational
# TYPE fah_up gauge
fah_up 1
```
