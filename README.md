# Shell exit status Exporter

GitLab: https://gitlab.com/transnano/shell_exit_status_exporter

Prometheus exporter written to execute and collect metrics on script exit status
and duration. Designed to allow the execution of probes where support for the
probe type wasn't easily configured with the Prometheus blackbox exporter.

Minimum supported Go Version: 1.11.0

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
docker run -d -p 9121:9121 --name shell-exit-status-exporter \
  -v `pwd`/config.yml:/etc/shell-exit-status-exporter/config.yml:ro \
  -config.file=/etc/shell-exit-status-exporter/config.yml
  -web.listen-address=":9121" \
  -web.telemetry-path="/metrics" \
  -config.shell="/bin/sh" \
  adhocteam/shell-exit-status-exporter:master
```

You'll need to customize the docker image or use the binary on the host system
to install tools such as curl for certain scenarios.

## Probing

To return the shell exit status exporter internal metrics exposed by the default Prometheus
handler:

`$ curl http://localhost:9121/metrics`

To execute a script, use the `name` parameter to the `/probe` endpoint:

`$ curl http://localhost:9121/probe?name=failure`

```
shell_exit_status_duration_seconds{script="failure"} 2.008337
shell_exit_status_status{script="failure"} 255
shell_exit_status_finished{script="failure"} 1
```

A regular expression may be specified with the `pattern` paremeter:

`$ curl http://localhost:9121/probe?pattern=.*`

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
