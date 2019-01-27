package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syst"
	"wg-menu/icon"
)

func onDataTick (i *StoreItem) {
	i.UpdateValue(getStats(i.statsRegex, i.title))
}

func onConnectClick (i *StoreItem) {
	// TODO: preference username
	toggleConnection(i.item.Checked(), "bravo")
}

func onConnectTick (i *StoreItem) {
	if wgRunning() {
		i.item.Check()
		i.UpdateTitle("Connected")
	} else {
		i.item.Uncheck()
		i.UpdateTitle("Connect")
	}
}

func onRealIPTick(i *StoreItem) {
	resp, err := http.Get("https://ipv4.icanhazip.com/")
	if err != nil {
		log.Println(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	err = resp.Body.Close()
	if err != nil {
		log.Println(err)
	}

	i.UpdateValue(strings.TrimSuffix(string(body), "\n"))
}

func onStoreTick(itemsStore *ItemsStore) {
	if wgRunning() {
		systray.SetIcon(icon.DataEnabled)
		systray.SetTooltip("WireGuard is running")
		itemsStore.toggleNonStatic(false)
	} else {
		systray.SetIcon(icon.DataDisabled)
		systray.SetTooltip("WireGuard is not running")
		itemsStore.toggleNonStatic(true)
	}
}

func onPrefClick(i *StoreItem) {
	// TODO: preference username config
	err := exec.Command("open", "/usr/local/etc/wg-menu/wg-menu.yaml").Start()
	if err != nil {
		log.Println(err)
	}
}