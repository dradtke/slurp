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
	Print(v ...interface{})
	Printf(format string, v ...interface{})

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
	l := log.New(os.Stdout, "", Flags)
	return &logger{l, ""}
}

type Printer interface {
	Printf(string, ...interface{})
}

type logger struct {
	Printer
	prefix string
}

func (l *logger) New(prefix string) Log {
	return &logger{l.Printer, l.prefix + prefix}
}

func (l *logger) Print(v ...interface{}) {
	l.Printer.Printf("%s%s", l.prefix, fmt.Sprint(v...))
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.Print(fmt.Sprintf(format, v...))
}

func (l *logger) Info(v ...interface{}) {
	l.Printer.Printf(color.GreenString("%s[INFO] %s ", l.prefix, fmt.Sprint(v...)))
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

func (l *logger) Warn(v ...interface{}) {
	l.Printer.Printf(color.YellowString("%s[WARN] %s ", l.prefix, fmt.Sprint(v...)))
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.Warn(fmt.Sprintf(format, v...))
}

func (l *logger) Error(v ...interface{}) {
	l.Printer.Printf(color.RedString("%s[ERR!] %s ", l.prefix, fmt.Sprint(v...)))
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
}

func (l *logger) Fatal(v ...interface{}) {
	l.Printer.Printf(color.RedString("%s[FATAL] %s ", l.prefix, fmt.Sprint(v...)))
	os.Exit(1)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.Fatal(fmt.Sprintf(format, v...))
}
