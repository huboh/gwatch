package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/huboh/gwatch/internal/pkg/config"
	"github.com/huboh/gwatch/internal/pkg/runner"
	"github.com/huboh/gwatch/internal/pkg/utils"
	"github.com/huboh/gwatch/internal/pkg/watcher"
)

type Gwatch struct {
	runner    *runner.Runner
	fsWatcher *watcher.Watcher
}

func (g *Gwatch) Kill() {
	if err := g.runner.Kill(); err != nil {
		log.Fatal(err)
	}

	if err := g.fsWatcher.Close(); err != nil {
		log.Fatal(err)
	}
}

func (g *Gwatch) Start() error {
	onBuild := func() {
		fmt.Println("[gwatch] Building...")
	}

	onRunBuild := func() {
		fmt.Println("[gwatch] Running...")
	}

	g.fsWatcher.OnError(func(e error) {
		log.Fatal("watcher error", e)
	})

	g.fsWatcher.OnEvent(watcher.WriteEvent, func(e watcher.Event) {
		if err := g.runner.Launch(onBuild, onRunBuild); err != nil {
			log.Fatal(err)
		}
	})

	g.fsWatcher.Listen(func(configs watcher.Configs) {
		fmt.Println("[gwatch] watching path(s):", strings.Join(configs.RootPaths, ","))
		fmt.Println("[gwatch] watching extension(s):", strings.Join(configs.Exts, ","))

		if err := g.runner.Launch(onBuild, onRunBuild); err != nil {
			log.Fatal(err)
		}
	})

	return nil
}

func watchConfigFile(onChange func()) {
	cfg := config.Default()

	// setup config for config file on a copy of app config
	cfg.Exts = []string{"yml", "yaml"}
	cfg.Paths = []string{config.ConfigPath}
	cfg.Delay = time.Millisecond * 100
	cfg.Recursive = false

	// new watcher for config file
	cfgWatcher := utils.Must(
		watcher.New(watcher.NewConfigs(*cfg)),
	)

	cfgWatcher.OnEvent(watcher.WriteEvent, func(e watcher.Event) {
		onChange()
	})

	// start watch
	cfgWatcher.Listen(nil)
}
