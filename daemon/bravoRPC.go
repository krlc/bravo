package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/mdlayher/wireguardctrl"
)

// WireGuard interface info
type Interface struct {
	Name string
	IP   string
}

type CurrentInterface struct {
	Interface

	Received int64
	Sent     int64
}

const (
	disconnected = iota
	connected
)

// Init RPC struct
type BravoRPC struct{
	w *wireguardctrl.Client
}

// Returns a list of available configurations (e.g. servers)
func (b *BravoRPC) ListAll(configPath *string, reply *[]Interface) error {
	// Basically, this procedure is called only once per GUI run,
	// so there's no *real* need to optimize this.

	files, err := ioutil.ReadDir(*configPath)
	if err != nil {
		log.Println(err)
		return err
	}

	list := make([]Interface, len(files))

	for i, f := range files {
		if !f.IsDir() {
			list[i].Name = strings.TrimSuffix(f.Name(), ".conf")
			list[i].IP = getIPFromConfig(path.Join(*configPath, f.Name()))
		}
	}

	*reply = list
	return nil
}

// Gathers VPN server IP from the WireGuard configuration file
func getIPFromConfig(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("File reading error", err)
		return ""
	}

	re := regexp.MustCompile("Endpoint = (\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})(:(\\d{1,5}))")
	st := re.FindString(string(data))

	return strings.TrimPrefix(st, "Endpoint = ")
}

func (b *BravoRPC) Connect(target *Interface, status *int) error {
	err := exec.Command("wg-quick", "up", target.Name).Start()
	*status = connected

	return err
}

func (b *BravoRPC) Disconnect(target *Interface, status *int) error {
	err := exec.Command("wg-quick", "down", target.Name).Start()
	*status = disconnected

	return err
}

const defaultError = "error"

// Returns the stats of the current connection, if connected
func (b *BravoRPC) GetCurrent(current *Interface, reply *CurrentInterface) error {
	devices, err := b.w.Devices()
	if err != nil {
		log.Println("Error getting devices:", err)
		return errors.New(defaultError) // TODO: am I sure?
	}

	if devices == nil {
		return nil
	}

	peers := devices[0].Peers
	if peers == nil {
		return nil
	}

	peer := peers[0] // we only expect 1 simultaneous connection

	*reply = CurrentInterface{
		Interface: Interface{
			Name: current.Name,
			IP:   peer.Endpoint.IP.String(), // updating IP; quite a questionable procedure, though
		},

		Received: peer.ReceiveBytes,
		Sent:     peer.TransmitBytes,
	}

	return nil
}
