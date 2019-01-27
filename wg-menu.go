package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syst"
	"time"
)

type StoreItem struct {
	item *systray.MenuItem
	title string
	value string
	static bool
	separator bool
	statsRegex string
	onClick func(m *StoreItem)
	onTick  func(m *StoreItem)
}

func (storeItem *StoreItem) ToString() string {
	if storeItem.value == "" {
		return storeItem.title
	}

	return storeItem.title + ": " + storeItem.value
}

func (storeItem *StoreItem) UpdateTitle(title string) {
	storeItem.title = title
	storeItem.item.SetTitle(storeItem.ToString())
}

func (storeItem *StoreItem) UpdateValue(value string) {
	storeItem.value = value
	storeItem.item.SetTitle(storeItem.ToString())
}

type ItemsStore struct {
	items map[string]*StoreItem
	timerQuit chan struct{}
	onTick func(itemsStore *ItemsStore)
}

func NewItemsStore() *ItemsStore {
	return &ItemsStore{items: make(map[string]*StoreItem)}
}

func (itemsStore *ItemsStore) addItems(storeItems []StoreItem) {
	for _, item := range storeItems {
		i := item
		i.item = systray.AddMenuItem(item.title, "")
		itemsStore.items[item.title] = &i

		if !item.static {
			i.item.Disable()
		}

		if item.separator {
			systray.AddSeparator()
		}
	}
}

// TODO: race condition bug
func (itemsStore *ItemsStore) watchClickEvents() {
	// current workaround:
	for _, v := range (*itemsStore).items {
		go func(v *StoreItem) {
			for {
				select {
				case <-v.item.ClickedCh:
					if v.onClick != nil {
						v.onClick(v)
					}
				}
			}
		}(v)
	}
}

// TODO: preference username
// TODO: refactor
func wgRunning() bool {
	name := "/var/run/wireguard"

	f, err := os.Open(name)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	out, err := f.Readdirnames(0);
	if err == io.EOF {
		return false
	}

	for _, v := range out {
		if strings.Contains(v, ".sock") {
			return true
		}
	}

	return false
}

// TODO: SIGNIFICANT performance issues
func (itemsStore *ItemsStore) watchTimerEvents() {
	itemsStore.timerQuit = make(chan struct{})

	// TODO: preferences refresh rate
	ticker := time.NewTicker(30 * time.Second)
	tickerFunc := func() {
		for _, v := range itemsStore.items {
			if !v.static {
				if v.onTick != nil {
					v.onTick(v)
				}
			}
		}
	}
	tickerFunc()

	// TODO: preferences refresh rate
	tickerStatic := time.NewTicker(3 * time.Second)
	tickerStaticFunc := func() {
		if itemsStore.onTick != nil {
			itemsStore.onTick(itemsStore)
		}

		for _, v := range itemsStore.items {
			if v.static {
				if v.onTick != nil {
					v.onTick(v)
				}
			}
		}
	}


	go func() {
		for {
			select {
			case <-tickerStatic.C:
				tickerStaticFunc()
			case <-ticker.C:
				tickerFunc()
			case <-itemsStore.timerQuit:
				tickerStatic.Stop()
				return
			}
		}
	}()
}

func (itemsStore *ItemsStore) toggleNonStatic(hide bool) {
	for _, v := range itemsStore.items {
		if !v.static {
			if hide {
				v.item.Hide()
			} else {
				v.item.Show()
			}
		}
	}
}

func getStats(regex, title string) string {
	// current workaround:
	cmd := exec.Command("wg")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		log.Fatal(errBuf.String())
	}

	re := regexp.MustCompile(regex)
	t := strings.ToLower(title)
	out := re.FindString(outBuf.String())

	return strings.TrimSpace(strings.Replace(out, t, "", 1))
}

func toggleConnection(disconnect bool, user string) {
	var arg string

	if disconnect {
		arg = "down"
	} else {
		arg = "up"
	}

	if err := exec.Command("wg-quick", arg, user).Start(); err != nil {
		log.Println(err)
	}
}

func onReady() {
	// init interface
	itemsStore := NewItemsStore()

	onExitClick := func(i *StoreItem) {
		close(itemsStore.timerQuit)
		systray.Quit()
	}

	itemsStore.addItems([]StoreItem{
		{title: "Connect",  static: true,  onClick: onConnectClick, onTick: onConnectTick, separator: true},
		{title: "Endpoint", static: false, statsRegex: "(\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})\\.(\\d{1,3})", onTick: onDataTick},
		{title: "Real IP",  static: false, onTick: onRealIPTick},
		{title: "Received", static: false, statsRegex: "\\d{0,5}\\.?\\d{0,5}\\s\\wiB\\sreceived", onTick:  onDataTick},
		{title: "Sent",     static: false, statsRegex: "\\d{0,5}\\.?\\d{0,5}\\s\\wiB\\ssent", onTick:  onDataTick, 	separator: true},
		{title: "Preferences", static: true, onClick: onPrefClick},
		{title: "Quit",     static: true,  onClick: onExitClick},
	})

	itemsStore.onTick = onStoreTick

	itemsStore.watchClickEvents()
	itemsStore.watchTimerEvents()
}

func main() {
	systray.Run(onReady, nil)
}