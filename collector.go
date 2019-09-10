package main

import (
	"encoding/json"
	"net"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type NVMeDeviceInfoCollector struct {
	socketPath string
}

func (collector NVMeDeviceInfoCollector) Describe(channel chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(collector, channel)
}

func (collector NVMeDeviceInfoCollector) Collect(channel chan<- prometheus.Metric) {
	log.Info("Collecting NVMe info")
	conn := connectToUnixSocket(collector.socketPath)
	bytes := readFromSocket(conn)

	deviceInfoList := unmarshal(bytes)
	for _, deviceInfo := range deviceInfoList {
		processDeviceInfo(deviceInfo, channel)
	}
	conn.Close()

}

func processDeviceInfo(deviceInfo NVMeDeviceInfo, channel chan<- prometheus.Metric) {
	labels := prometheus.Labels{
		"device": deviceInfo.DevicePath,
		"model":  deviceInfo.ModelNumber,
		"serial": deviceInfo.SerialNumber,
	}

	for key, value := range deviceInfo.SmartLog {
		desc := prometheus.NewDesc(key, "", nil, labels)
		channel <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(value),
		)

	}
}

func connectToUnixSocket(socketPath string) net.Conn {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatal("error dialing to unix socket:", err)
	}
	return conn
}

func readFromSocket(conn net.Conn) []byte {
	buf := make([]byte, 256*1024)
	n, err := conn.Read(buf[:])
	if err != nil {

	}
	return buf[:n]
}

func unmarshal(bytes []byte) NVMeDeviceInfoList {
	deviceInfoList := NVMeDeviceInfoList{}
	err := json.Unmarshal(bytes, &deviceInfoList)
	if err != nil {
		log.Error("error unmarshalling deviceInfoList json", err)
		return NVMeDeviceInfoList{}
	}
	return deviceInfoList
}
