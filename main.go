package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type appConfig struct {
	socketPath    string
	listenAddress string
}

type NVMeDeviceInfo struct {
	DevicePath   string         `json:"device_path"`
	ModelNumber  string         `json:"model_number"`
	SerialNumber string         `json:"serial_number"`
	SmartLog     map[string]int `json:"smart_log"`
}

type NVMeDeviceInfoList []NVMeDeviceInfo

func main() {

	config := loadConfig()

	log.Info("NVMe exporter is starting. Listening on ", config.listenAddress)

	registry := configureRegistry(config)

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.ListenAndServe(config.listenAddress, nil)
}

func loadConfig() *appConfig {
	socketPath := flag.String("socket-path", "/var/run/nvme_exporter.sock", "Path to UNIX socket from which to read NVMe device info")
	listenAddress := flag.String("listen-address", ":9110", "The address to listen on")
	flag.Parse()

	return &appConfig{
		socketPath:    *socketPath,
		listenAddress: *listenAddress,
	}
}

func configureRegistry(config *appConfig) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	collector := NVMeDeviceInfoCollector{socketPath: config.socketPath}
	registry.MustRegister(collector)

	registry.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)
	return registry
}
