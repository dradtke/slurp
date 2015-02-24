package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

var Rate = time.Millisecond * 300

type Log interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Warn(v ...interface{})
	Warnf(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})

	ReadProgress(io.Reader, string, int64) io.ReadCloser
	Counter(string, int) *Counter

	New(string) Log
}

var Flags = log.Ltime

func New() Log {
	l := log.New(os.Stdout, " ", Flags)
	return &logger{l, ""}
}

type Printer interface {
	Printf(string, ...interface{})
}

type logger struct {
	printer Printer
	prefix string
}

func (l *logger) New(prefix string) Log {
	return &logger{l.printer, l.prefix + prefix}
}

func (l *logger) Info(v ...interface{}) {
	l.printer.Printf("[INFO] %s%s ", l.prefix, fmt.Sprint(v...))
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

func (l *logger) Warn(v ...interface{}) {
	l.printer.Printf(color.YellowString("[WARN] %s%s ", l.prefix, fmt.Sprint(v...)))
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.Warn(fmt.Sprintf(format, v...))
}

func (l *logger) Error(v ...interface{}) {
	l.printer.Printf(color.RedString("[ERR!] %s%s ", l.prefix, fmt.Sprint(v...)))
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
}

func (l *logger) Fatal(v ...interface{}) {
	l.printer.Printf(color.RedString("[FATAL] %s%s ", l.prefix, fmt.Sprint(v...)))
	os.Exit(1)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.Fatal(fmt.Sprintf(format, v...))
}
