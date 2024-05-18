package watcher

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/huboh/gwatch/internal/pkg/config"
)

//
//
//* Watcher Configs
//
//

type WatcherConfigs struct {
	// exts is the list of file extensions to watch for
	exts []string

	// paths is the list of directories we are watching
	paths []string

	// exclude is the list of directories to exclude from the watch list
	exclude []string

	// rootDir
	rootDir string

	// recursive enables watching on the subdirectories of the paths in the watch list
	recursive bool
}

func NewConfigs(config config.Config) *WatcherConfigs {
	var (
		c = &WatcherConfigs{
			exts:      config.Exts,
			paths:     config.Paths,
			exclude:   config.Exclude,
			rootDir:   config.Root,
			recursive: config.Recursive,
		}

		// addMatchedDir adds eligible dir to config's paths
		addMatchedDir fs.WalkDirFunc = func(dir string, dirEnt fs.DirEntry, err error) error {
			if dirEnt.IsDir() {
				for _, e := range c.exclude {
					pattern := filepath.Join(c.rootDir, e)
					isDirOrSub, err := filepath.Match(pattern, dir)

					if err != nil {
						return err
					}

					if isDirOrSub {
						return filepath.SkipDir
					}
				}

				if !slices.Contains(c.paths, dir) {
					c.paths = append(c.paths, dir)
				}
			}

			return nil
		}
	)

	// recursively add eligible pathNames to configs's paths
	if c.recursive {
		for _, p := range c.paths {
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
	configs         *WatcherConfigs
	watcher         *fsnotify.Watcher
	eventHandlers   map[EventType][]EventHandler
	eventErrHandler func(error)
}

func New(configs *WatcherConfigs) (*Watcher, error) {
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

	if e = w.Watch(w.configs.paths...); e != nil {
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

func (w *Watcher) Listen(onListen func(paths []string)) {
	defer w.watcher.Close()

	if onListen != nil {
		go onListen(w.watcher.WatchList())
	}

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
				event            = NewEvent(EventType(evt.Op), evt.Name)
				handlers, exists = w.eventHandlers[event.Type]
				extension        = strings.TrimPrefix(filepath.Ext(event.Path), ".")
			)

			if !exists {
				continue
			}

			stat, err := os.Stat(event.Path)
			if err != nil {
				go w.eventErrHandler(err)
				return
			}

			// ensure it is a file and we're watching the extension
			//
			//? instead of call IsDir() directly on the stat, we get the Mode() then check if it's a file.
			//? doing this we get the correct file mode for the specific `os`, then check if its a regular file.
			//? because the FileInfo (stat variable) is an interface and the impl might be different depending on the `os` and `filesystem`
			if stat.Mode().IsRegular() && slices.Contains(w.configs.exts, extension) {
				for _, handler := range handlers {
					go handler(*event)
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
