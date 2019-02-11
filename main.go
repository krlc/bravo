package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/getlantern/systray"
)

const (
	BlockSize = 6
)

func onPrefClick(gui *GUI) {
	if err := exec.Command("open", gui.config.prefPath).Start(); err != nil {
		log.Println("Error opening preference file:", err)
	}
}

func onExitClick(gui *GUI) {
	gui.exit()
}

func onReady() {
	config := NewConfig()
	setupLog(config)

	gui := NewGUI(config)

	// We cannot add menu items to a certain position,
	// but server list is going to be updated from time to time,
	// therefore, in order to add servers before the last element (exit button),
	// we need to allocate a block of hidden menu items, ready to be used as a server item
	servers := gui.createBlock(BlockSize)
	stats   := gui.createBlock(BlockSize)

	gui.add(guiItem{title: "Preferences", onClick: onPrefClick})
	gui.add(guiItem{title: "Quit", onClick: onExitClick})

	for _, s := range stats.getServers() {
		servers.add(guiItem{
			title:    s.Name,
			value:    s.IP,
			onClick:  servers.connect,
			onTick:   setColor,
		})
	}

	gui.onUpdate = func() {
		for _, s := range stats.getServerChanges() {
			servers.update(guiItem{
				title:    s.Name,
				value:    s.IP,
				onClick:  servers.connect,
				onTick:   setColor,
			})
		}

		for _, s := range stats.getStats() {
			stats.addOrUpdate(guiItem{
				title:    s.Title,
				value:    toReadable(s.Value),
			})
		}
	}
}

func setupLog(config *Config) {
	f, err := os.OpenFile(config.logPath, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	config.onExit(func() { // add file closing procedure
		f.Close()
	})
}

func main() {
	systray.Run(onReady, nil)
}
