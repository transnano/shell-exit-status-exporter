# Shell exit status Exporter ![Releases](https://github.com/transnano/shell-exit-status-exporter/workflows/Releases/badge.svg) ![Publish Docker image](https://github.com/transnano/shell-exit-status-exporter/workflows/Publish%20Docker%20image/badge.svg) ![Vulnerability Scan](https://github.com/transnano/shell-exit-status-exporter/workflows/Vulnerability%20Scan/badge.svg)

![License](https://img.shields.io/github/license/transnano/shell-exit-status-exporter?style=flat)

![Container image version](https://img.shields.io/docker/v/transnano/shell-exit-status-exporter/latest?style=flat)
![Container image size](https://img.shields.io/docker/image-size/transnano/shell-exit-status-exporter/latest?style=flat)
![Container image pulls](https://img.shields.io/docker/pulls/transnano/shell-exit-status-exporter?style=flat)

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/transnano/shell-exit-status-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/transnano/shell-exit-status-exporter)](https://goreportcard.com/report/github.com/transnano/shell-exit-status-exporter)

GitHub: https://github.com/transnano/shell-exit-status-exporter

Prometheus exporter written to execute and collect metrics on script exit status
and duration. Designed to allow the execution of probes where support for the
probe type wasn't easily configured with the Prometheus blackbox exporter.

Minimum supported Go Version: 1.15.0

## Sample Configuration

``` yaml
scripts:
  - name: success
    script: sleep 5

  - name: failure
    script: sleep 2 && exit 255

  - name: timeout
    script: sleep 5
    timeout: 1
```

## Running

You can run via docker with:

``` shell
docker run -d -p 9062:9062 --name shell-exit-status-exporter \
  -v `pwd`/config.yml:/etc/shell-exit-status-exporter/config.yml:ro \
  -config.file=/etc/shell-exit-status-exporter/config.yml
  -web.listen-address=":9062" \
  -web.telemetry-path="/metrics" \
  -config.shell="/bin/sh" \
  transnano/shell-exit-status-exporter:v0.0.4
```

You'll need to customize the docker image or use the binary on the host system
to install tools such as curl for certain scenarios.

## Probing

To return the shell exit status exporter internal metrics exposed by the default Prometheus
handler:

`$ curl http://localhost:9062/metrics`

To execute a script, use the `name` parameter to the `/probe` endpoint:

`$ curl http://localhost:9062/probe?name=failure`

```
shell_exit_status_duration_seconds{script="failure"} 2.008337
shell_exit_status_status{script="failure"} 255
shell_exit_status_finished{script="failure"} 1
```

A regular expression may be specified with the `pattern` paremeter:

`$ curl http://localhost:9062/probe?pattern=.*`

```
shell_exit_status_duration_seconds{script="timeout"} 1.005727
shell_exit_status_status{script="timeout"} 1
shell_exit_status_finished{script="timeout"} 0
shell_exit_status_duration_seconds{script="failure"} 2.015021
shell_exit_status_status{script="failure"} 255
shell_exit_status_finished{script="failure"} 1
shell_exit_status_duration_seconds{script="success"} 5.013670
shell_exit_status_status{script="success"} 0
shell_exit_status_finished{script="success"} 1
```

## Design

YMMV if you're attempting to execute a large number of scripts, and you'd be
better off creating an exporter that can handle your protocol without launching
shell processes for each scrape.
