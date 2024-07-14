package config

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// logln writes a line of output to the specified writer. It directly passes the
// provided arguments to the writer without any formatting or prefix.
//
// Args:
//   - w (io.Writer): The writer where the line will be written.
//   - a ([]any): The arguments to write to the writer.
func logln(w io.Writer, a []any) {
	fmt.Fprintln(w, a...)
}

// loglnWithPrefixes writes a formatted line to the specified writer,
// prefixing it with a hash (#) and additional prefixes for context.
//
// Args:
//   - w (io.Writer): The writer where the line will be written.
//   - a ([]any): The arguments to write to the writer, prefixed with a hash.
//   - p (...string): Additional prefixes to add for context.
func (l Logger) loglnWithPrefixes(w io.Writer, a []any, p ...string) {
	prefixes := []any{"#"}
	for _, prefix := range slices.Concat(slices.Clone(l.defaultPrefixes), p) {
		if prefix != "" {
			prefixes = append(prefixes, fmt.Sprintf("[%s]", prefix))
		}
	}

	logln(w, slices.Concat(prefixes, a))
}

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
//
// Args:
//   - prefix (string): The prefix to add to the logger.
//
// Returns:
//   - Logger: A new Logger instance with the added prefix.
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
//
// Returns:
//   - string: The formatted prefix for the current retry attempt, or an empty
//     string if no retries.
func (l Logger) tryPrefix() string {
	if l.maxTry <= 1 {
		return ""
	}

	return fmt.Sprintf("%d/%d", l.try, l.maxTry)
}

// Info logs standard messages to stdout, prefixed with a hash.
//
// Args:
//   - a (...any): The arguments to log to stdout.
func (l Logger) Info(a ...any) {
	l.loglnWithPrefixes(l.stdout, a)
}

// Infof logs formatted messages to stdout, prefixed with a hash.
//
// Args:
//   - format (string): The format string for the log message.
//   - a (...any): The arguments to format and log to stdout.
func (l Logger) Infof(format string, a ...any) {
	l.Info(fmt.Sprintf(format, a...))
}

// InfoWithoutPrefix logs messages to stdout without any prefix.
//
// Args:
//   - a (...any): The arguments to log to stdout.
func (l Logger) InfoWithoutPrefix(a ...any) {
	logln(l.stdout, a)
}

// InfoTry logs messages to stdout with retry prefix, showing the attempt number
// if retries are configured.
//
// Args:
//   - a (...any): The arguments to log to stdout with retry context.
func (l Logger) InfoTry(a ...any) {
	l.loglnWithPrefixes(l.stdout, a, l.tryPrefix())
}

// Warn logs warning messages to stderr, prefixed with "[WARNING]:".
//
// Args:
//   - a (...any): The arguments to log as a warning to stderr.
func (l Logger) Warn(a ...any) {
	l.loglnWithPrefixes(l.stderr, a, "WARNING")
}

// Warnf logs formatted warning messages to stderr, prefixed with "[WARNING]:".
//
// Args:
//   - format (string): The format string for the warning message.
//   - a (...any): The arguments to format and log as a warning to stderr.
func (l Logger) Warnf(format string, a ...any) {
	l.Warn(fmt.Sprintf(format, a...))
}

// Error logs error messages to stderr, prefixed with a hash.
//
// Args:
//   - a (...any): The arguments to log as an error to stderr.
func (l Logger) Error(a ...any) {
	l.loglnWithPrefixes(l.stderr, a)
}

// Errorf logs formatted error messages to stderr, prefixed with a hash.
//
// Args:
//   - format (string): The format string for the error message.
//   - a (...any): The arguments to format and log as an error to stderr.
func (l Logger) Errorf(format string, a ...any) {
	l.Error(fmt.Sprintf(format, a...))
}

// ErrorTry logs error messages with retry context to stderr, transforming
// multi-line error messages into a single line.
//
// Args:
//   - err (error): The error to log.
//   - a (...any): Additional context to log with the error message.
func (l Logger) ErrorTry(err error, a ...any) {
	// Convert multi-line errors to a single line for clarity in logs.
	errInOneLine := strings.ReplaceAll(err.Error(), "\n", "")

	l.loglnWithPrefixes(l.stderr, append(a, errInOneLine), l.tryPrefix())
}
