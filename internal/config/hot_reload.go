package config

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/singl3focus/load_balancer-go/pkg/logger"
)

type UpdateCallback func(cfg *Config)

func WatchConfig(path string, lg logger.Logger, cfg *Config, callback UpdateCallback) {
    watcher, _ := fsnotify.NewWatcher()
    defer watcher.Close()

    absPath, _ := filepath.Abs(path)
    dir := filepath.Dir(absPath)

    if err := watcher.Add(absPath); err != nil {
        panic(err)
    }
    if err := watcher.Add(dir); err != nil {
        panic(err)
    }

    for {
        select {
        case event, ok := <-watcher.Events:
            lg.Debug("event received", "(info)", event)

            if !ok {
                return
            }
            if event.Has(fsnotify.Write) && event.Name == absPath {
                newCfg := MustLoadCfg(false)
                *cfg = *newCfg
                callback(newCfg)
            }
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            lg.Error("config watcher error", "(describe)", err)
        }
    }
}