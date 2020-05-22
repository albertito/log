package log

import (
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

func mustNewFile(t *testing.T) (string, *Logger) {
	f, err := ioutil.TempFile("", "log_test-")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	l, err := NewFile(f.Name())
	if err != nil {
		t.Fatalf("failed to open new log file: %v", err)
	}

	return f.Name(), l
}

func checkContentsMatch(t *testing.T, name, path, expected string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	got := string(content)
	if !regexp.MustCompile(expected).Match(content) {
		t.Errorf("%s: regexp %q did not match %q",
			name, expected, got)
	}
}

func testLogger(t *testing.T, fname string, l *Logger) {
	l.LogTime = false
	l.Infof("message %d", 1)
	checkContentsMatch(t, "info-no-time", fname,
		"^_ log_test.go:....   message 1\n")

	os.Truncate(fname, 0)
	l.Infof("message %d\n", 1)
	checkContentsMatch(t, "info-with-newline", fname,
		"^_ log_test.go:....   message 1\n")

	os.Truncate(fname, 0)
	l.LogTime = true
	l.Infof("message %d", 1)
	checkContentsMatch(t, "info-with-time", fname,
		`^....-..-.. ..:..:..\.\d{6} _ log_test.go:....   message 1\n`)

	os.Truncate(fname, 0)
	l.LogTime = false
	l.Errorf("error %d", 1)
	checkContentsMatch(t, "error", fname, `^E log_test.go:....   error 1\n`)

	if l.V(Debug) {
		t.Fatalf("Debug level enabled by default (level: %v)", l.Level)
	}

	os.Truncate(fname, 0)
	l.LogTime = false
	l.Debugf("debug %d", 1)
	checkContentsMatch(t, "debug-no-log", fname, `^$`)

	os.Truncate(fname, 0)
	l.Level = Debug
	l.Debugf("debug %d", 1)
	checkContentsMatch(t, "debug", fname, `^\. log_test.go:....   debug 1\n`)

	if !l.V(Debug) {
		t.Errorf("l.Level = Debug, but V(Debug) = false")
	}

	os.Truncate(fname, 0)
	l.Level = Info
	l.Log(Debug, 0, "log debug %d", 1)
	l.Log(Info, 0, "log info %d", 1)
	checkContentsMatch(t, "log", fname,
		`^_ log_test.go:....   log info 1\n`)

	os.Truncate(fname, 0)
	l.Level = Info
	l.Log(Fatal, 0, "log fatal %d", 1)
	checkContentsMatch(t, "log", fname,
		`^☠ log_test.go:....   log fatal 1\n`)

	// Test some combinations of options.
	cases := []struct {
		name      string
		logTime   bool
		logLevel  bool
		logCaller bool
		expected  string
	}{
		{
			"show everything",
			true, true, true,
			`^....-..-.. ..:..:..\.\d{6} _ log_test.go:....   message 1\n`,
		}, {
			"caller+level",
			false, true, true,
			`^_ log_test.go:....   message 1\n`,
		}, {
			"time",
			true, false, false,
			`^....-..-.. ..:..:..\.\d{6} message 1\n`,
		}, {
			"none",
			false, false, false,
			`message 1\n`,
		},
	}
	for _, c := range cases {
		os.Truncate(fname, 0)
		l.LogTime = c.logTime
		l.LogLevel = c.logLevel
		l.LogCaller = c.logCaller
		l.Infof("message %d", 1)
		checkContentsMatch(t, c.name, fname, c.expected)
	}
}

func TestBasic(t *testing.T) {
	fname, l := mustNewFile(t)
	defer l.Close()
	defer os.Remove(fname)

	testLogger(t, fname, l)
}

func TestDefaultFile(t *testing.T) {
	fname, l := mustNewFile(t)
	l.Close()
	defer os.Remove(fname)

	*logFile = fname

	Init()

	testLogger(t, fname, Default)
}

func TestReopen(t *testing.T) {
	fname, l := mustNewFile(t)
	defer l.Close()
	defer os.Remove(fname)
	l.LogTime = false

	l.Infof("pre rename")
	checkContentsMatch(t, "r", fname, `^_ log_test.go:....   pre rename\n`)

	os.Rename(fname, fname+"-m")
	defer os.Remove(fname + "-m")
	l.Infof("post rename")
	checkContentsMatch(t, "r", fname+"-m", `pre rename\n.* post rename`)

	if err := l.Reopen(); err != nil {
		t.Errorf("reopen: %v", err)
	}
	l.Infof("post reopen")
	checkContentsMatch(t, "r", fname, `^_ log_test.go:....   post reopen\n`)

	// NewFile with an absolute path should resolve it internally to a full
	// one, so reopen can work.
	l, err := NewFile("test-relative-file")
	defer l.Close()
	defer os.Remove("test-relative-file")

	if err != nil {
		t.Fatalf("failed to open file for testing: %v", err)
	}
	if l.fname[0] != '/' {
		t.Fatalf("internal fname is not absolute: %q", l.fname)
	}

}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func TestReopenNull(t *testing.T) {
	l := New(nopCloser{ioutil.Discard})
	if err := l.Reopen(); err != nil {
		t.Errorf("reopen: %v", err)
	}
}

// Benchmark a call below the verbosity level.
func BenchmarkDebugf(b *testing.B) {
	l := New(nopCloser{ioutil.Discard})
	defer l.Close()
	for i := 0; i < b.N; i++ {
		l.Debugf("test %d", i)
	}
}

// Benchmark a normal call.
func BenchmarkInfof(b *testing.B) {
	l := New(nopCloser{ioutil.Discard})
	defer l.Close()
	for i := 0; i < b.N; i++ {
		l.Infof("test %d", i)
	}
}
