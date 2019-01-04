FROM       quay.io/prometheus/busybox:latest
MAINTAINER Ryota Suginaga <transnano.jp@gmail.com>

COPY shell-exit-status-exporter /bin/shell-exit-status-exporter
COPY shell-exit-status-exporter.yml /etc/shell-exit-status-exporter/config.yml

EXPOSE     9121
ENTRYPOINT ["/bin/shell-exit-status-exporter"]
CMD        ["-config.file=/etc/shell-exit-status-exporter/config.yml"]
