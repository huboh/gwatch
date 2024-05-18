package main

import (
	"log"
	"time"

	"github.com/huboh/gwatch/internal/pkg/config"
	"github.com/huboh/gwatch/internal/pkg/runner"
	"github.com/huboh/gwatch/internal/pkg/utils"
	"github.com/huboh/gwatch/internal/pkg/watcher"
)

func main() {
	done := make(chan struct{})
	appConfig := utils.Must(config.New())

	// start gwatch in a goroutine
	go func() {
		for {
			appRnr := runner.New(*appConfig)
			fsWatcher := utils.Must(watcher.New(watcher.NewConfigs(*appConfig)))

			select {
			// kill gwatch
			case <-done:
				if err := appRnr.Kill(); err != nil {
					log.Fatal(err)
				}

				if err := fsWatcher.Close(); err != nil {
					log.Fatal(err)
				}

				// reset channel so we dont close a closed channel
				done = make(chan struct{})

			// start application
			case err := <-utils.AsyncResult(func() error { return startGwatch(fsWatcher, appRnr) }):
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	// watch config file for changes
	watchGwatchConfig(func() {
		// reload gwatch config
		if err := appConfig.Reload(); err != nil {
			log.Fatal("error reloading application due to config change: ", err)
		}

		// signal runner and watcher to exit, then restart.
		utils.CloseSafely(done)
	})
}

func startGwatch(w *watcher.Watcher, r *runner.Runner) error {
	var err error

	defer func() {
		if err = r.Kill(); err != nil {
			return
		}

		if err = w.Close(); err != nil {
			return
		}
	}()

	w.OnError(func(e error) {
		log.Println("watcher error", e)
	})

	w.OnEvent(watcher.WriteEvent, func(e watcher.Event) {
		r.Launch()
	})

	w.Listen(func(paths []string) {
		// for _, v := range paths {
		// 	log.Println("listening for path(s) changes in:", v)
		// }

		r.Launch()
	})

	return err
}

func watchGwatchConfig(onChange func()) {
	cfg := config.Default()

	// setup config for config file on a copy of app config
	cfg.Exts = []string{"yml", "yaml"}
	cfg.Paths = []string{config.ConfigPath}
	cfg.Delay = time.Millisecond
	cfg.Recursive = false

	// new watcher for config file
	cfgWatcher := utils.Must(watcher.New(watcher.NewConfigs(*cfg)))

	defer cfgWatcher.Close()

	cfgWatcher.OnEvent(watcher.WriteEvent, func(e watcher.Event) {
		// detected changes to config file. reload application
		onChange()
	})

	// watch config for changes
	cfgWatcher.Listen(nil)
}
