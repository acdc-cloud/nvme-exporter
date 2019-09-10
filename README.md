# NVMe exporter
NVMe exporter written in golang. The exporter has two binaries

- nvme-exporter-helper - Runs as root in a systemd service, so it can access the devices. It can be supplied a socket from systemd through systemd socket activation, which is used to communicate with nvme-exporter
- nvme-exporter - Reads from the socket and exposes the metrics through a /metrics endpoint for prometheus to scrape