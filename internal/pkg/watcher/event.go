package watcher

import (
	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path string
	Type EventType
}

func NewEvent(t EventType, p string) *Event {
	return &Event{
		Path: p,
		Type: t,
	}
}

type EventType uint32

type EventHandler func(Event)

const (
	// ChmodEvent is emitted when a File attributes was changed.
	ChmodEvent = EventType(fsnotify.Chmod)

	// WriteEvent is emitted when a pathname was written to; this does *not* mean the write has finished.
	WriteEvent = EventType(fsnotify.Write)

	// CreateEvent is emitted when a path(dir or file) was created
	CreateEvent = EventType(fsnotify.Create)

	// RemoveEvent is emitted when a path was removed; any watches on it will be removed.
	RemoveEvent = EventType(fsnotify.Remove)

	// RenameEvent is emitted when a path was renamed to something else;
	// any watched on it will be removed.
	RenameEvent = EventType(fsnotify.Rename)
)

func (e EventType) String() string {
	return fsnotify.Op(e).String()
}
