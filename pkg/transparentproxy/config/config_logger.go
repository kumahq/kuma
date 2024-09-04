package config

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// Logger provides simple logging capabilities directing output to specified
// stdout and stderr writers. It supports retry-aware logging, tracking the
// current and maximum retry counts to adjust log messages accordingly.
type Logger struct {
	// Target writer where standard output logs are written.
	stdout io.Writer
	// Target writer where error output logs are written.
	stderr io.Writer
	// The maximum number of tries for retry operations.
	maxTry int
	// The current try number within the retry operation context.
	try int
	// Default prefixes to be used in log messages.
	defaultPrefixes []string
}

// WithPrefix returns a new Logger instance with the specified prefix added to
// the default prefixes.
func (l Logger) WithPrefix(prefix string) Logger {
	return Logger{
		stdout:          l.stdout,
		stderr:          l.stderr,
		maxTry:          l.maxTry,
		try:             l.try,
		defaultPrefixes: append(l.defaultPrefixes, prefix),
	}
}

// tryPrefix generates a logging prefix based on the current retry attempt.
// It returns a formatted string showing the current attempt and the total
// attempts, or an empty string if no retries are needed (maxTry <= 1).
func (l Logger) tryPrefix() string {
	if l.maxTry <= 1 {
		return ""
	}

	return fmt.Sprintf("%d/%d", l.try, l.maxTry)
}

// Info logs standard messages to stdout, prefixed with a hash.
func (l Logger) Info(a ...any) {
	l.loglnWithPrefixes(l.stdout, a)
}

// Infof logs formatted messages to stdout, prefixed with a hash.
func (l Logger) Infof(format string, a ...any) {
	l.Info(fmt.Sprintf(format, a...))
}

// InfoWithoutPrefix logs messages to stdout without any prefix.
func (l Logger) InfoWithoutPrefix(a ...any) {
	logln(l.stdout, a)
}

// InfoTry logs messages to stdout with retry prefix, showing the attempt number
// if retries are configured.
func (l Logger) InfoTry(a ...any) {
	l.loglnWithPrefixes(l.stdout, a, l.tryPrefix())
}

// Warn logs warning messages to stderr, prefixed with "[WARNING]:".
func (l Logger) Warn(a ...any) {
	l.loglnWithPrefixes(l.stderr, a, "WARNING")
}

// Warnf logs formatted warning messages to stderr, prefixed with "[WARNING]:".
func (l Logger) Warnf(format string, a ...any) {
	l.Warn(fmt.Sprintf(format, a...))
}

// Error logs error messages to stderr, prefixed with a hash.
func (l Logger) Error(a ...any) {
	l.loglnWithPrefixes(l.stderr, a)
}

// Errorf logs formatted error messages to stderr, prefixed with a hash.
func (l Logger) Errorf(format string, a ...any) {
	l.Error(fmt.Sprintf(format, a...))
}

// ErrorTry logs error messages with retry context to stderr, transforming
// multi-line error messages into a single line.
func (l Logger) ErrorTry(err error, a ...any) {
	// Convert multi-line errors to a single line for clarity in logs.
	errInOneLine := strings.ReplaceAll(err.Error(), "\n", "")

	l.loglnWithPrefixes(l.stderr, append(a, errInOneLine), l.tryPrefix())
}

// logln writes a line of output to the specified writer. It directly passes the
// provided arguments to the writer without any formatting or prefix.
func logln(w io.Writer, a []any) {
	fmt.Fprintln(w, a...)
}

// loglnWithPrefixes writes a formatted line to the specified writer,
// prefixing it with a hash (#) and additional prefixes for context.
func (l Logger) loglnWithPrefixes(w io.Writer, args []any, prefixes ...string) {
	finalPrefixes := []any{"#"}
	for _, prefix := range slices.Concat(slices.Clone(l.defaultPrefixes), prefixes) {
		if prefix != "" {
			finalPrefixes = append(finalPrefixes, fmt.Sprintf("[%s]", prefix))
		}
	}

	logln(w, slices.Concat(finalPrefixes, args))
}
