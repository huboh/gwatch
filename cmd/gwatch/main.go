package main

import (
	"log"

	"github.com/huboh/gwatch/internal/pkg/config"
	"github.com/huboh/gwatch/internal/pkg/runner"
	"github.com/huboh/gwatch/internal/pkg/utils"
	"github.com/huboh/gwatch/internal/pkg/watcher"
)

func main() {
	done := make(chan struct{})
	gwatchCfg := utils.Must(config.New())

	// start gwatch in a goroutine
	go func() {
		for {
			gwatch := &Gwatch{
				runner:    runner.New(*gwatchCfg),
				fsWatcher: utils.Must(watcher.New(watcher.NewConfigs(*gwatchCfg))),
			}

			select {
			// kill gwatch
			case <-done:
				gwatch.Kill()

				// reset channel so we dont close a closed channel
				done = make(chan struct{})

			// start gwatch
			case <-utils.AsyncResult(gwatch.Start):
			}
		}
	}()

	// watch config file for changes
	watchConfigFile(func() {
		// signal gwatch to exit, so it can restart.
		defer utils.CloseSafely(done)

		// reload gwatch config on change
		if err := gwatchCfg.Reload(); err != nil {
			log.Fatal("error restarting gwatch due to config change: ", err)
		}
	})
}
