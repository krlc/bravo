package main

import (
    "io/ioutil"
    "log"
    "os"
    "path/filepath"

    "github.com/fsnotify/fsnotify"
    "gopkg.in/yaml.v2"
)

type ConfigStore struct {
    WgConfig     string `yaml:"wgConfig"`
    RefreshRate  int64  `yaml:"refreshRate"`

    configChange chan interface{}
    watcher *fsnotify.Watcher
}

func (c *ConfigStore) Update() {
    filePrefix, _ := filepath.Abs(configName)
    file, err := os.OpenFile(filePrefix, os.O_RDONLY, 0444)
    if err != nil {
        log.Println("Error opening config file:", err)
        return
    }

    data, err := ioutil.ReadAll(file)
    if err != nil {
        log.Println("Error reading config file:", err)
        return
    }

    if err := yaml.Unmarshal(data, c); err != nil {
        log.Println("Error marshalling config file:", err)
        return
    }
}

func (c *ConfigStore) Watch() {
    var err error

    c.watcher, err = fsnotify.NewWatcher()
    if err != nil {
        log.Println("Error setting config watcher:", err)
        return
    }

    if err = c.watcher.Add(configName); err != nil {
        log.Println("Error adding config watcher:", err)
        return
    }

    c.configChange = make(chan interface{})

    go func() {
        for {
            select {
            case w := <-c.watcher.Events:
                if w.Op == fsnotify.Write {
                    log.Println("Updating config...")
                    c.configChange <- struct{}{}
                }
            case <-c.watcher.Errors:
                c.Close()
            }
        }
    }()
}

func (c *ConfigStore) Close() {
    //close(c.configChange)
    if err := c.watcher.Close(); err != nil {
        log.Println("Error closing watcher:", err)
    }
}

func NewConfig() *ConfigStore {
    conf := &ConfigStore{}
    conf.Update()
    conf.Watch()

    return conf
}
