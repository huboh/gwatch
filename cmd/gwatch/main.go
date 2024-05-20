package main

import (
	"fmt"
	"log"

	"github.com/huboh/gwatch/internal/pkg/config"
	"github.com/huboh/gwatch/internal/pkg/runner"
	"github.com/huboh/gwatch/internal/pkg/utils"
	"github.com/huboh/gwatch/internal/pkg/watcher"
)

func main() {
	done := make(chan struct{})
	gwatchCfg := utils.Must(config.New())

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

				// reset channel so we don't close a closed channel
				done = make(chan struct{})

			// start gwatch
			case <-utils.AsyncResult(gwatch.Start):
			}
		}
	}()

	watchConfigFile(func() {
		// signal gwatch to restart.
		defer utils.CloseSafely(done)

		// reload gwatch config
		if err := gwatchCfg.Reload(); err != nil {
			log.Fatal("error reloading gwatch config: ", err)
		}

		fmt.Println("[gwatch] restarting gwatch due to changes to config file")
	})
}
