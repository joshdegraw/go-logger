package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	gray    = 37
)

func getModuleName(entry *logrus.Entry) string {
	mod := entry.Data["module"]
	switch mod.(type) {
	case string:
		return mod.(string)
	case fmt.Stringer:
		return mod.(fmt.Stringer).String()
	default:
		return fmt.Sprintf("%v", mod)
	}
}

func printEntry(buf *bytes.Buffer, entry *logrus.Entry,
	colored bool, modLength int) {
	var colorPfx, colorSfx string
	if colored {
		var levelColor int
		switch entry.Level {
		case logrus.DebugLevel:
			levelColor = gray
		case logrus.WarnLevel:
			levelColor = yellow
		case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			levelColor = red
		default:
			levelColor = blue
		}
		colorPfx = "\x1b[" + strconv.Itoa(levelColor) + "m"
		colorSfx = "\x1b[0m"
	}
	buf.WriteString(entry.Time.Format("2006-01-02T15:04:05.000") + " ")
	modStr := getModuleName(entry)
	if len(modStr) > modLength {
		modStr = modStr[:modLength]
	} else {
		modStr = strings.Repeat(" ", modLength-len(modStr)) + modStr
	}
	buf.WriteString("[" + modStr + "] ")
	levelText := strings.ToUpper(entry.Level.String())[0:4]
	buf.WriteString(colorPfx + levelText + colorSfx + "  ")
	buf.WriteString(entry.Message + "\n")
}

type consoleHook struct {
	logger *Logger
	sync.Mutex
}

func newConsoleHook(logger *Logger) *consoleHook {
	hook := &consoleHook{logger: logger}
	return hook
}

func (hook *consoleHook) Fire(entry *logrus.Entry) error {
	var buf bytes.Buffer
	isColorTerminal := hook.logger.isTerminal && (runtime.GOOS != "windows")
	printEntry(&buf, entry, isColorTerminal, hook.logger.GetModuleLength())
	hook.Lock()
	defer hook.Unlock()
	var out io.Writer
	if entry.Level <= logrus.WarnLevel {
		// Warnings and above forward to Stderr...
		out = os.Stderr
	} else {
		// ...all others to Stdout
		out = os.Stdout
	}
	_, err := io.Copy(out, &buf)
	if err != nil {
		return fmt.Errorf("Failed to write to log, %v\n", err)
	}
	return nil
}

func (hook *consoleHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
