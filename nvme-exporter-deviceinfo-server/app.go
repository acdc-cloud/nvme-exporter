package main

import (
	"encoding/json"
	"net"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type App struct {
	config *appConfig
}

type appConfig struct {
	socketPath         string
	nvmeExecutablePath string
}

func (app *App) serveNVMeDeviceInfo(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		log.Info("Handling incoming connection on unix socket")
		app.handleConnection(conn)
	}
}

func (app *App) handleConnection(c net.Conn) {
	deviceInfoList := app.getDeviceInfoList()

	bytes, err := json.Marshal(deviceInfoList)
	if err != nil {
		log.Error("unable to marshal data to json")
	}

	_, err = c.Write(bytes)
	if err != nil {
		log.Error("unable to write data to socket")
	}

	c.Close()
}

type NVMeDevice struct {
	DevicePath   string `json:"DevicePath"`
	ModelNumber  string `json:"ModelNumber"`
	SerialNumber string `json:"SerialNumber"`
}

type NVMeDeviceList struct {
	Devices []NVMeDevice `json:"Devices"`
}

type NVMeDeviceInfo struct {
	DevicePath   string         `json:"device_path"`
	ModelNumber  string         `json:"model_number"`
	SerialNumber string         `json:"serial_number"`
	SmartLog     map[string]int `json:"smart_log"`
}

func (app *App) getDeviceInfoList() []NVMeDeviceInfo {

	devices, err := app.getNVMeDevices()
	if err != nil {
		return []NVMeDeviceInfo{}
	}

	deviceInfoList := make([]NVMeDeviceInfo, 0)

	for _, d := range devices {
		deviceInfo := NVMeDeviceInfo{
			DevicePath:   d.DevicePath,
			ModelNumber:  d.ModelNumber,
			SerialNumber: d.SerialNumber,
			SmartLog:     app.getSmartLog(d.DevicePath),
		}
		deviceInfoList = append(deviceInfoList, deviceInfo)
	}

	return deviceInfoList

}

func (app *App) getNVMeDevices() ([]NVMeDevice, error) {
	out, err := exec.Command(app.config.nvmeExecutablePath, "list", "-o", "json").Output()
	if err != nil {
		log.Error("unable to execute nvme list command:", err)
		return nil, err
	}

	deviceList := new(NVMeDeviceList)
	err = json.Unmarshal(out, deviceList)
	if err != nil {
		log.Error("unable to unmarshal json")
		return nil, err
	}

	return deviceList.Devices, nil

}

func (app *App) getSmartLog(device string) map[string]int {
	log.Info("Querying smart log from ", device)
	out, err := exec.Command(app.config.nvmeExecutablePath, "smart-log", device, "-o", "json").Output()
	if err != nil {
		log.Error("unable to execute nvme smart-log command", err)
		return nil
	}

	smartLog := make(map[string]int)
	err = json.Unmarshal(out, &smartLog)
	if err != nil {
		log.Error("unable to unmarshal json")
		return nil
	}

	return smartLog
}
