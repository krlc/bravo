package main

import (
    "time"

    "github.com/getlantern/systray"
)

const (
    static = iota
    dynamic
)

type StoreItem struct {
    item *systray.MenuItem

    title     string
    value     string
    elemType  int
    separator bool

    onClick func(m *StoreItem)
    onTick  func(m *StoreItem, s *StatsStore)
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

func (storeItem *StoreItem) Check() {
    storeItem.item.Check()
}

func (storeItem *StoreItem) Uncheck() {
    storeItem.item.Uncheck()
}


type ItemsStore struct {
    items 	  map[string]*StoreItem
    timerQuit chan struct{}
    onTick 	  func(itemsStore *ItemsStore, stats *StatsStore)
}

func NewItemsStore() *ItemsStore {
    return &ItemsStore{
        items: 	   make(map[string]*StoreItem),
        timerQuit: make(chan struct{}),
    }
}

func (itemsStore *ItemsStore) addItems(storeItems []StoreItem) {
    for _, item := range storeItems {
        i := item
        i.item = systray.AddMenuItem(item.title, "")
        itemsStore.items[item.title] = &i

        if item.elemType == dynamic {
            i.item.Disable()
        }

        if item.separator {
            systray.AddSeparator()
        }
    }
}

func (itemsStore *ItemsStore) toggleDynamic(hide bool) {
    for _, v := range itemsStore.items {
        if v.elemType == dynamic {
            if hide {
                v.item.Hide()
            } else {
                v.item.Show()
            }
        }
    }
}

// TODO: race condition bug
func (itemsStore *ItemsStore) watchClickEvents() {
    // current workaround:
    for _, v := range (*itemsStore).items {
        go func(v *StoreItem) {
            for {
                <-v.item.ClickedCh

                if v.onClick != nil {
                    v.onClick(v)
                }
            }
        }(v)
    }
}

func (itemsStore *ItemsStore) watchTimerEvents() {
    // acquiring WireGuard stats
    stats := &StatsStore{}
    stats.Update()

    ticker := time.NewTicker(time.Duration(conf.RefreshRate) * time.Second)
    tickerFunc := func() {
        if itemsStore.onTick != nil {
            itemsStore.onTick(itemsStore, stats)
        }

        for _, v := range itemsStore.items {
            if v.onTick != nil {
                v.onTick(v, stats)
            }
        }
    }
    tickerFunc() // first init tick

    go func() {
        for {
            select {
            case <-conf.configChange:
                conf.Update() // gather new config
                ticker.Stop()
                ticker = time.NewTicker(time.Duration(conf.RefreshRate) * time.Second)
            case <-ticker.C:
                stats.Update() // gather WireGuard stats
                tickerFunc()   // update GUI
            case <-itemsStore.timerQuit:
                ticker.Stop()
                stats.Close() // close wireguardctrl client
                conf.Close()
                return
            }
        }
    }()
}
