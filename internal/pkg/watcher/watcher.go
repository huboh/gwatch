package watcher

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/huboh/gwatch/internal/pkg/config"
	"github.com/huboh/gwatch/internal/pkg/utils"
)

//
//
//* Watcher Configs
//
//

type Configs struct {
	// Exts is the list of file extensions to watch for
	Exts []string

	// Paths is the list of directories and subdirectories we are watching
	Paths []string

	// Exclude is the list of directories to Exclude from the watch list
	Exclude []string

	// recursive set the Delay for event handlers execution
	Delay time.Duration

	// RootDir iis the current working directory
	RootDir string

	// Recursive enables watching on the subdirectories of the paths in the watch list
	Recursive bool

	// RootPaths is the list of parent directories to watch from the config
	RootPaths []string
}

func NewConfigs(config config.Config) *Configs {
	var (
		c = &Configs{
			Exts:      config.Exts,
			Paths:     config.Paths,
			Exclude:   config.Exclude,
			RootDir:   config.Root,
			Delay:     config.Delay,
			Recursive: config.Recursive,
			RootPaths: config.Paths,
		}

		// addMatchedDir adds eligible dir to config's paths
		addMatchedDir fs.WalkDirFunc = func(dir string, dirEnt fs.DirEntry, err error) error {
			if dirEnt.IsDir() {
				for _, e := range c.Exclude {
					pattern := filepath.Join(c.RootDir, e)
					isDirOrSub, err := filepath.Match(pattern, dir)

					if err != nil {
						return err
					}

					if isDirOrSub {
						return filepath.SkipDir
					}
				}

				if !slices.Contains(c.Paths, dir) {
					c.Paths = append(c.Paths, dir)
				}
			}

			return nil
		}
	)

	// recursively add eligible pathNames to configs's paths
	if c.Recursive {
		for _, p := range c.Paths {
			if err := filepath.WalkDir(p, addMatchedDir); err != nil {
				panic(err)
			}
		}
	}

	return c
}

//
//
//* Watcher
//
//

type Watcher struct {
	configs         *Configs
	watcher         *fsnotify.Watcher
	eventHandlers   map[EventType][]EventHandler
	eventErrHandler func(error)
}

func New(configs *Configs) (*Watcher, error) {
	var (
		e error

		// new watcher
		w = &Watcher{
			configs:       configs,
			eventHandlers: make(map[EventType][]EventHandler),
		}
	)

	// wrap error incase of error
	defer func() {
		if err := recover(); err != nil {
			if err, isErr := err.(error); isErr {
				e = fmt.Errorf("error creating watcher: %w", err)
			}
		}
	}()

	if w.watcher, e = fsnotify.NewWatcher(); e != nil {
		return nil, e
	}

	if e = w.Watch(w.configs.Paths...); e != nil {
		return nil, e
	}

	return w, nil
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}

func (w *Watcher) Watch(paths ...string) error {
	for _, p := range paths {
		if err := w.watcher.Add(p); err != nil {
			return err
		}
	}

	return nil
}

func (w *Watcher) Listen(onListen func(configs Configs)) {
	defer w.watcher.Close()

	if onListen != nil {
		go onListen(*w.configs)
	}

	var (
		event   *Event
		handler EventHandler

		// execute last handler call after config's delay
		debouncedHandler = utils.Debounce(w.configs.Delay, func() {
			if event != nil && handler != nil {
				go handler(*event)
			}
		})
	)

	for {
		select {
		case err, open := <-w.watcher.Errors:
			if !open {
				return
			}

			go w.eventErrHandler(err)

		case evt, open := <-w.watcher.Events:
			if !open {
				return
			}

			var (
				fsEvent          = NewEvent(EventType(evt.Op), evt.Name)
				handlers, exists = w.eventHandlers[fsEvent.Type]
				extension        = strings.TrimPrefix(filepath.Ext(fsEvent.Path), ".")
			)

			if !exists {
				continue
			}

			stat, err := os.Stat(fsEvent.Path)

			if err != nil {
				go w.eventErrHandler(err)
				return
			}

			// ensure it is a file and we're watching the extension
			//
			//? instead of call IsDir() directly on the stat, we get the Mode() then check if it's a file.
			//? doing this we get the correct file mode for the specific `os`, then check if its a regular file.
			//? because the FileInfo (stat variable) is an interface and the impl might be different depending on the `os` and `filesystem`
			if stat.Mode().IsRegular() && slices.Contains(w.configs.Exts, extension) {
				for _, h := range handlers {
					event = fsEvent
					handler = h

					debouncedHandler()
				}
			}
		}
	}
}

func (w *Watcher) OnError(h func(error)) {
	w.eventErrHandler = h
}

func (w *Watcher) OnEvent(eType EventType, handler EventHandler) {
	// add handler to event handlers list
	w.eventHandlers[eType] = append(w.eventHandlers[eType], handler)
}
