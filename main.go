package main

import (
    "log"
    "os"
    "os/exec"

    "github.com/getlantern/systray"
    "github.com/krlc/bravo/icon"
)

const (
    configName  = "./bravo.yaml"
)

// Timer tick handlers:

func onEndpointTick (i *StoreItem, stats *StatsStore) {
    i.UpdateValue(stats.Endpoint) // set GUI menu to show endpoint IP
}

func onReceivedTick (i *StoreItem, stats *StatsStore) {
    i.UpdateValue(stats.Received) // set GUI menu to show bytes received
}

func onSentTick (i *StoreItem, stats *StatsStore) {
    i.UpdateValue(stats.Sent) // set GUI menu to show bytes sent
}

func onConnectTick (i *StoreItem, stats *StatsStore) {
    if stats.Alive {
        i.Check()
        i.UpdateTitle("Connected")
    } else {
        i.Uncheck()
        i.UpdateTitle("Connect")
    }
}

func onStoreTick(itemsStore *ItemsStore, stats *StatsStore) {
    itemsStore.toggleDynamic(!stats.Alive) // hide or show statistics

    if stats.Alive {
        systray.SetIcon(icon.DataEnabled)
        systray.SetTooltip("WireGuard is running")
    } else {
        systray.SetIcon(icon.DataDisabled)
        systray.SetTooltip("WireGuard is not running")
    }
}

// GUI click handlers:

func onConnectClick (i *StoreItem) {
    toggleConnection(i.item.Checked()) // on/off vpn connection
}

func onPrefClick(i *StoreItem) {
    err := exec.Command("open", configName).Start()
    if err != nil {
        log.Println(err)
    }
}

// messing up with global space
// TODO: refactor
var conf *ConfigStore

func onReady() {
    // init interface
    itemsStore := NewItemsStore()
    itemsStore.onTick = onStoreTick

    onExitClick := func(i *StoreItem) {
        itemsStore.timerQuit <- struct{}{}
        systray.Quit()
    }

    // add GUI elements
    itemsStore.addItems([]StoreItem{
        {title: "Connect",     elemType: static,  onTick:  onConnectTick, onClick: onConnectClick, separator: true},
        {title: "Endpoint",    elemType: dynamic, onTick:  onEndpointTick},
        {title: "Received",    elemType: dynamic, onTick:  onReceivedTick},
        {title: "Sent",        elemType: dynamic, onTick:  onSentTick, separator: true},
        {title: "Preferences", elemType: static,  onClick: onPrefClick},
        {title: "Quit",        elemType: static,  onClick: onExitClick},
    })

    // init event jobs
    itemsStore.watchClickEvents()
    itemsStore.watchTimerEvents()
}

func main() {
    f, err := os.OpenFile("/var/log/bravo.log", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    log.SetOutput(f)

    conf = NewConfig()
    systray.Run(onReady, nil)
}