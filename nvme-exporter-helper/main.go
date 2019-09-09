package main

import (
	"flag"
	"net"
	"os"

	"github.com/coreos/go-systemd/activation"
	log "github.com/sirupsen/logrus"
)

func loadConfig() *appConfig {
	socketPath := flag.String("socket-path", "/var/run/nvme_exporter.sock", "Path to UNIX socket when not using systemd socket activation")
	nvmeExecutablePath := flag.String("nvme-path", "/usr/sbin/nvme", "Path to nvme-cli executable")
	flag.Parse()

	return &appConfig{
		socketPath:         *socketPath,
		nvmeExecutablePath: *nvmeExecutablePath,
	}
}

func main() {

	log.Info("NVMe exporter helper is starting...")

	config := loadConfig()

	removeSocketIfAlreadyExists(config.socketPath)

	listener, usingSocketActivation := getSocketListener(config)
	if !usingSocketActivation {
		defer listener.Close()
	}

	app := App{config: config}
	app.serveNVMeDeviceInfo(listener)

}

func getSocketListener(config *appConfig) (net.Listener, bool) {

	if listener, ok := useSocketActivation(); ok {
		log.Info("Using systemd socket activation")
		return listener, true
	} else {
		log.Info("Using unix socket ", config.socketPath)
		return getUnixSocketListener(config), false
	}

}

func useSocketActivation() (net.Listener, bool) {
	listeners, err := activation.Listeners()
	if err != nil {
		log.Error("couldn't get socket activation file descriptors")
	}

	if len(listeners) == 0 {
		return nil, false
	}

	if len(listeners) == 1 {
		return listeners[0], true
	}

	if len(listeners) > 1 {
		log.Fatal("unexpected number of socket activation file descriptors")
	}

	panic("impossible code path")
}

func getUnixSocketListener(config *appConfig) net.Listener {
	listener, err := net.Listen("unix", config.socketPath)
	if err != nil {
		log.Fatal("unix socket listen error:", err)
	}
	return listener
}

func removeSocketIfAlreadyExists(socketPath string) {
	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal("couldn't remove socket", socketPath)
	}
}
