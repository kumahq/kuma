package config

import (
	"fmt"
	"io"
	"strings"
)

// logln writes a line of output to the specified writer. It directly passes the
// provided arguments to the writer without any formatting or prefix.
func logln(w io.Writer, a []any) {
	fmt.Fprintln(w, a...)
}

// loglnWithHashPrefix writes a formatted line to the specified writer,
// prefixing it with a hash (#) to denote the line as a comment.
func loglnWithHashPrefix(w io.Writer, a []any) {
	logln(w, append([]any{"#"}, a...))
}

// Logger provides simple logging capabilities directing output to specified
// stdout and stderr writers. It supports retry-aware logging, tracking the
// current and maximum retry counts to adjust log messages accordingly.
type Logger struct {
	stdout io.Writer // Target writer where standard output logs are written.
	stderr io.Writer // Target writer where error output logs are written.
	maxTry int       // The maximum number of tries for retry operations.
	try    int       // The current try number within the retry operation ctx.
}

// tryPrefix generates a logging prefix based on the current retry attempt.
// It returns a formatted string showing the current attempt and the total
// attempts, or nil if no retries are needed (maxTry <= 1).
func (l Logger) tryPrefix() []any {
	if l.maxTry <= 1 {
		return nil
	}

	return []any{fmt.Sprintf("[%d/%d]", l.try, l.maxTry)}
}

// Info logs standard messages to stdout, prefixed with a hash.
func (l Logger) Info(a ...any) {
	loglnWithHashPrefix(l.stdout, a)
}

// InfoWithoutPrefix logs messages to stdout without any prefix.
func (l Logger) InfoWithoutPrefix(a ...any) {
	logln(l.stdout, a)
}

// InfoTry logs messages to stdout with retry prefix, showing the attempt number
// if retries are configured.
func (l Logger) InfoTry(a ...any) {
	l.Info(append(l.tryPrefix(), a...)...)
}

// Warn logs warning messages to stderr, prefixed with "[WARNING]:"
func (l Logger) Warn(a ...any) {
	loglnWithHashPrefix(l.stderr, append([]any{"[WARNING]:"}, a...))
}

// Error logs error messages to stderr, prefixed with a hash.
func (l Logger) Error(a ...any) {
	loglnWithHashPrefix(l.stderr, a)
}

// ErrorTry logs error messages with retry context to stderr, transforming
// multi-line error messages into a single line.
func (l Logger) ErrorTry(err error, a ...any) {
	result := append(l.tryPrefix(), a...)
	// Convert multi-line errors to a single line for clarity in logs.
	errInOneLine := strings.ReplaceAll(err.Error(), "\n", "")

	l.Error(append(result, errInOneLine)...)
}
