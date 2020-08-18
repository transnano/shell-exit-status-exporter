FROM       quay.io/prometheus/busybox:latest
LABEL      maintainer="Transnano <transnano.jp@gmail.com>"

COPY       shell-exit-status-exporter /bin/shell-exit-status-exporter
COPY       shell-exit-status-exporter.yml /etc/shell-exit-status-exporter/config.yml

EXPOSE     9062
ENTRYPOINT ["/bin/shell-exit-status-exporter"]
CMD        ["-config.file=/etc/shell-exit-status-exporter/config.yml"]
