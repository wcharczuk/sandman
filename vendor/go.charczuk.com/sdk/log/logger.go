package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const defaultFlags = FlagError | FlagWarn | FlagInfo

// New returns a new logger.
func New(opts ...Option) *Logger {
	l := new(Logger)
	l.flags = defaultFlags.Flag
	l.output = os.Stdout
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Option mutates a logger.
type Option func(l *Logger)

// OptConfig configures the logger based on a config type.
func OptConfig(cfg Config) Option {
	return func(l *Logger) {
		if cfg.Disabled {
			l.output = io.Discard
			l.flags = FlagDisabled.Flag
			return
		}
		l.skipSource = cfg.SkipSource
		l.skipTime = cfg.SkipTime
		l.flags = cfg.Flags.Flag
		for key, value := range cfg.DefaultAttrs {
			l.defaultAttrs = append(l.defaultAttrs, String(key, value))
		}
	}
}

// OptOutput sets the logger output.
func OptOutput(output io.Writer) Option {
	return func(l *Logger) {
		l.output = output
	}
}

// OptFlags sets the flags provider.
func OptFlags(flags FlagProvider) Option {
	return func(l *Logger) {
		l.flags = flags.Flag
	}
}

// OptSkipSource skips the source segment in the log output line.
func OptSkipSource(skipSource bool) Option {
	return func(l *Logger) {
		l.skipSource = skipSource
	}
}

// OptSkipTime skips the time segment in the log output line.
func OptSkipTime(skipTime bool) Option {
	return func(l *Logger) {
		l.skipTime = skipTime
	}
}

// OptAttrs sets the default attributes for log output lines.
func OptAttrs(attrs ...any) Option {
	return func(l *Logger) {
		l.defaultAttrs = argsToAttrSlice(attrs)
	}
}

// Logger can be used to write to output.
type Logger struct {
	output       io.Writer
	flags        func() Flag
	group        string
	skipSource   bool
	skipTime     bool
	sourceSkip   int
	defaultAttrs []Attr
}

// Error writes an error message to the logger.
func (l *Logger) Error(msg string, attrs ...any) {
	if l == nil {
		return
	}
	if l.flags()&FlagError == 0 {
		return
	}
	l.writeOutput("ERROR", msg, attrs...)
}

// Err writes an error to the logger.
func (l *Logger) Err(err error, attrs ...any) {
	if l == nil {
		return
	}
	if l.flags()&FlagError == 0 {
		return
	}
	l.writeOutput("ERROR", fmt.Sprintf("%+v", err), attrs...)
}

// Info writes an info message to the logger.
func (l *Logger) Info(msg string, attrs ...any) {
	if l == nil {
		return
	}
	if l.flags()&FlagInfo == 0 {
		return
	}
	l.writeOutput("INFO", msg, attrs...)
}

// Warn writes a warning message to the logger.
func (l *Logger) Warn(msg string, attrs ...any) {
	if l == nil {
		return
	}
	if l.flags()&FlagWarn == 0 {
		return
	}
	l.writeOutput("WARN", msg, attrs...)
}

// Debug writes a debug message to the logger.
func (l *Logger) Debug(msg string, attrs ...any) {
	if l == nil {
		return
	}
	if l.flags()&FlagDebug == 0 {
		return
	}
	l.writeOutput("DEBUG", msg, attrs...)
}

// Silly writes a silly message to the logger.
func (l *Logger) Silly(msg string, attrs ...any) {
	if l == nil {
		return
	}
	if l.flags()&FlagSilly == 0 {
		return
	}
	l.writeOutput("SILLY", msg, attrs...)
}

// WithGroup returns a new logger with a given group.
func (l *Logger) WithGroup(group string) *Logger {
	l2 := l.clone()
	l2.group = group
	return l2
}

// WithSourceSkip returns a new logger with a given source stack frame skip.
func (l *Logger) WithSourceSkip(skip int) *Logger {
	l2 := l.clone()
	l2.sourceSkip = skip
	return l2
}

// WithAttrs returns a new logger and adds new attrs to the logger.
func (l *Logger) WithAttrs(attr ...any) *Logger {
	l2 := l.clone()
	l2.defaultAttrs = append(l2.defaultAttrs, argsToAttrSlice(attr)...)
	return l2
}

func (l *Logger) clone() *Logger {
	l2 := *l
	return &l2
}

func (l *Logger) writeOutput(level, msg string, args ...any) {
	buf := new(bytes.Buffer)
	if !l.skipTime {
		buf.WriteString(l.formatTimestamp(time.Now().UTC()) + " ")
	}
	buf.WriteString(level)
	if l.group != "" {
		buf.WriteString(" [" + l.group + "]: ")
	} else {
		buf.WriteString(": ")
	}
	buf.WriteString(msg)
	attrs := argsToAttrSlice(args)
	attrs = append(l.defaultAttrs, attrs...)
	if !l.skipSource {
		buf.WriteString(" ")
		buf.WriteString(String("source", l.formatPC(l.getPC())).String())
	}
	if len(attrs) > 0 {
		buf.WriteString(" ")
	}
	for index, attr := range attrs {
		buf.WriteString(attr.String())
		if index < len(attrs)-1 {
			buf.WriteString(" ")
		}
	}
	buf.WriteRune('\n')
	l.output.Write(buf.Bytes())
}

func (l *Logger) formatTimestamp(ts time.Time) string {
	return ts.Format(time.RFC3339Nano)
}

func (l *Logger) formatPC(pc uintptr) string {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()
	return fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
}

func (l *Logger) getPC() (pc uintptr) {
	if !l.skipSource {
		var pcs [1]uintptr
		// there is a min-skip here (3) based on:
		// [getPC, writeOutput, Info|Debug|Warn|Error|Silly ]
		// then the user can add extra skips themselves
		runtime.Callers(l.sourceSkip+4, pcs[:])
		pc = pcs[0]
	}
	return
}
