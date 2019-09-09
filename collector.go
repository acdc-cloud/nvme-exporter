package main

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type NVMeDeviceInfoCollector struct {
	socketPath string
}

func (c NVMeDeviceInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c NVMeDeviceInfoCollector) Collect(ch chan<- prometheus.Metric) {
	log.Info("Collecting NVMe info")
	conn := connectToUnixSocket(c.socketPath)
	bytes := readFromSocket(conn)

	deviceInfoList := unmarshal(bytes)
	for _, deviceInfo := range deviceInfoList {
		processDeviceInfo(deviceInfo, ch)
	}
	conn.Close()

}

func processDeviceInfo(deviceInfo NVMeDeviceInfo, ch chan<- prometheus.Metric) {
	labels := prometheus.Labels{
		"device": deviceInfo.DevicePath,
		"model":  deviceInfo.ModelNumber,
		"serial": deviceInfo.SerialNumber,
	}

	for key, value := range deviceInfo.SmartLog {
		desc := prometheus.NewDesc(key, "", nil, labels)
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(value),
		)

	}
}
