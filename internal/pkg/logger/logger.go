package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/huboh/gwatch/internal/pkg/runner"
)

type Color string
type Colors map[Color]color.Attribute

const (
	Red = Color(iota)
	Cyan
	Blue
	White
	Green
	Yellow
	Magenta
)

var (
	rawColor = Color("raw")
	colorMap = Colors{
		Red:     color.FgHiRed,
		Cyan:    color.FgHiCyan,
		Blue:    color.FgHiBlue,
		White:   color.FgHiWhite,
		Green:   color.FgHiGreen,
		Yellow:  color.FgHiYellow,
		Magenta: color.FgHiMagenta,
	}
)

//*
//*
//* Log Function
//*
//*

type LogFunc func(string, ...any) (n int, err error)

func NewLogFunc(colorName Color) (log LogFunc) {
	return func(format string, v ...any) (n int, err error) {
		format = strings.TrimSpace(strings.ReplaceAll(format, "\n", ""))

		if len(format) == 0 {
			return
		}

		if colorName == rawColor {
			log = fmt.Printf
		} else {
			log = color.New(getColor(colorName)).Printf
		}

		return log(("[gwatch] " + format + "\n"), v...)
	}
}

//*
//*
//* Logger
//*
//*

type logger struct {
	Loggers map[Color]LogFunc
}

func New() *logger {
	loggers := make(map[Color]LogFunc, len(colorMap))

	for name := range colorMap {
		loggers[name] = NewLogFunc(name)
	}

	loggers["default"] = defaultLogger()

	return &logger{
		Loggers: loggers,
	}
}

func (l *logger) Main() LogFunc {
	return l.getLogger(White)
}

func (l *logger) Runner() LogFunc {
	return l.getLogger(Yellow)
}

func (l *logger) Watcher() LogFunc {
	return l.getLogger(Blue)
}

func (l *logger) getLogger(name Color) LogFunc {
	v, ok := l.Loggers[name]

	if !ok {
		return rawLogger()
	}

	return v
}

func getColor(name Color) color.Attribute {
	if v, ok := colorMap[name]; ok {
		return v
	}

	return color.FgWhite
}

func rawLogger() LogFunc {
	return NewLogFunc("raw")
}

func defaultLogger() LogFunc {
	return NewLogFunc("white")
}

func ClearConsole() error {
	args := []string{"clear"}

	if runtime.GOOS == "windows" {
		args = []string{"cmd", "/c", "cls"}
	}

	return runner.NewCommand(args, "").Run(os.Stdout, os.Stderr, nil)
}
