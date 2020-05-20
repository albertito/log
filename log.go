// Package log implements a simple logger.
//
// It implements an API somewhat similar to "github.com/google/glog" with a
// focus towards simplicity and integration with standard tools such as
// systemd.
//
// There are command line flags (defined using the flag package) to control
// the behaviour of the default logger. By default, it will write to stderr
// without timestamps; this is suitable for systemd (or equivalent) logging.
//
// Command-line flags:
//
//  -alsologtostderr
//        also log to stderr, in addition to the file
//  -logfile string
//        file to log to (enables logtime)
//  -logtime
//        include the time when writing the log to stderr
//  -logtosyslog string
//        log to syslog, with the given tag
//  -v int
//        verbosity level (1 = debug)
package log // import "blitiri.com.ar/go/log"

import (
	"flag"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Flags that control the default logging.
var (
	vLevel = flag.Int("v", 0, "verbosity level (1 = debug)")

	logFile = flag.String("logfile", "",
		"file to log to (enables logtime)")

	logToSyslog = flag.String("logtosyslog", "",
		"log to syslog, with the given tag")

	logTime = flag.Bool("logtime", false,
		"include the time when writing the log to stderr")

	alsoLogToStderr = flag.Bool("alsologtostderr", false,
		"also log to stderr, in addition to the file")
)

// Type of a logging level, to prevent confusion.
type Level int

// Standard logging levels.
const (
	Fatal = Level(-2)
	Error = Level(-1)
	Info  = Level(0)
	Debug = Level(1)
)

var levelToLetter = map[Level]string{
	Fatal: "☠",
	Error: "E",
	Info:  "_",
	Debug: ".",
}

// A Logger represents a logging object that writes logs to a writer.
type Logger struct {
	// Minimum level to log. Messages below this level will be dropped.
	// Note this field is NOT thread safe, if you change it, it is strongly
	// recommended to do so right after creating the logger, and before it is
	// used.
	// The use of this field should be considered EXPERIMENTAL, the API for it
	// could change in the future.
	Level Level

	// File name, if this logger is backed by a file. It's used to implement
	// reopening.
	fname string

	logTime    bool
	callerSkip int
	w          io.WriteCloser
	sync.Mutex
}

// New creates a new Logger, which writes logs to w.
func New(w io.WriteCloser) *Logger {
	return &Logger{
		w:          w,
		callerSkip: 0,
		Level:      Info,
		logTime:    true,
	}
}

// NewFile creates a new Logger, which writes logs to the given file.
func NewFile(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	l := New(f)
	l.logTime = true
	l.fname = path
	return l, nil
}

// NewSyslog creates a new Logger, which writes logs to syslog, using the
// given priority and tag.
func NewSyslog(priority syslog.Priority, tag string) (*Logger, error) {
	w, err := syslog.New(priority, tag)
	if err != nil {
		return nil, err
	}

	l := New(w)
	l.logTime = false
	return l, nil
}

// Close the writer behind the logger.
func (l *Logger) Close() {
	l.w.Close()
}

// Reopen the file behind the logger, if any. This can be used to implement
// log rotation.
//
// Only works for loggers created via NewFile, otherwise it is a no-op.
//
// EXPERIMENTAL, this API could change in the future.
func (l *Logger) Reopen() error {
	if l.fname == "" {
		return nil
	}

	f, err := os.OpenFile(l.fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.Lock()
	l.Close()
	l.w = f
	l.Unlock()
	return nil
}

// V returns true if the logger's level is >= the one given, false otherwise.
// It can be used to decide whether to use or gather debugging information
// only at a certain level, to avoid computing it needlessly.
func (l *Logger) V(level Level) bool {
	return level <= l.Level
}

// Log the message into the logger, at the given level. This is low-level and
// should rarely be needed, but it's available to allow the caller to have
// more complex logic if needed. skip is the number of frames to skip when
// computing the file name and line number.
func (l *Logger) Log(level Level, skip int, format string, a ...interface{}) error {
	if !l.V(level) {
		return nil
	}

	// Message.
	msg := fmt.Sprintf(format, a...)

	// Caller.
	_, file, line, ok := runtime.Caller(1 + l.callerSkip + skip)
	if !ok {
		file = "unknown"
	}
	fl := fmt.Sprintf("%s:%-4d", filepath.Base(file), line)
	if len(fl) > 18 {
		fl = fl[len(fl)-18:]
	}
	msg = fmt.Sprintf("%-18s", fl) + " " + msg

	// Level.
	letter, ok := levelToLetter[level]
	if !ok {
		letter = strconv.Itoa(int(level))
	}
	msg = letter + " " + msg

	// Time.
	if l.logTime {
		msg = time.Now().Format("2006-01-02 15:04:05.000000 ") + msg
	}

	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	l.Lock()
	_, err := l.w.Write([]byte(msg))
	l.Unlock()
	return err
}

// Debugf logs information at a Debug level.
func (l *Logger) Debugf(format string, a ...interface{}) {
	l.Log(Debug, 1, format, a...)
}

// Infof logs information at a Info level.
func (l *Logger) Infof(format string, a ...interface{}) {
	l.Log(Info, 1, format, a...)
}

// Errorf logs information at an Error level. It also returns an error
// constructed with the given message, in case it's useful for the caller.
func (l *Logger) Errorf(format string, a ...interface{}) error {
	l.Log(Error, 1, format, a...)
	return fmt.Errorf(format, a...)
}

// Fatalf logs information at a Fatal level, and then exits the program with a
// non-0 exit code.
func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.Log(Fatal, 1, format, a...)
	// TODO: Log traceback?
	os.Exit(1)
}

// The default logger, used by the top-level functions below.
var Default = &Logger{
	w:          os.Stderr,
	callerSkip: 1,
	Level:      Info,
	logTime:    false,
}

// Initialize the default logger, based on the command-line flags.
func Init() {
	flag.Parse()
	var err error

	if *logToSyslog != "" {
		Default, err = NewSyslog(syslog.LOG_DAEMON|syslog.LOG_INFO, *logToSyslog)
		if err != nil {
			panic(err)
		}
	} else if *logFile != "" {
		Default, err = NewFile(*logFile)
		if err != nil {
			panic(err)
		}
		*logTime = true
	}

	if *alsoLogToStderr && Default.w != os.Stderr {
		Default.w = multiWriteCloser(Default.w, os.Stderr)
	}

	Default.callerSkip = 1
	Default.Level = Level(*vLevel)
	Default.logTime = *logTime
}

// V is a convenient wrapper to Default.V.
func V(level Level) bool {
	return Default.V(level)
}

// Log is a convenient wrapper to Default.Log.
func Log(level Level, skip int, format string, a ...interface{}) {
	Default.Log(level, skip, format, a...)
}

// Debugf is a convenient wrapper to Default.Debugf.
func Debugf(format string, a ...interface{}) {
	Default.Debugf(format, a...)
}

// Infof is a convenient wrapper to Default.Infof.
func Infof(format string, a ...interface{}) {
	Default.Infof(format, a...)
}

// Errorf is a convenient wrapper to Default.Errorf.
func Errorf(format string, a ...interface{}) error {
	return Default.Errorf(format, a...)
}

// Fatalf is a convenient wrapper to Default.Fatalf.
func Fatalf(format string, a ...interface{}) {
	Default.Fatalf(format, a...)
}

// multiWriteCloser creates a WriteCloser that duplicates its writes and
// closes to all the provided writers.
func multiWriteCloser(wc ...io.WriteCloser) io.WriteCloser {
	return mwc(wc)
}

type mwc []io.WriteCloser

func (m mwc) Write(p []byte) (n int, err error) {
	for _, w := range m {
		if n, err = w.Write(p); err != nil {
			return
		}
	}
	return
}
func (m mwc) Close() error {
	for _, w := range m {
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}
