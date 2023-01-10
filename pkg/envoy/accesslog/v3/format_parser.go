package v3

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	commandWithArgsRE = regexp.MustCompile(`^%(?P<command>[A-Z_0-9]+)(?:\((?P<args>[^\)]*)\))?(?:[:](?P<limit>[0-9]+))?%`)
	newlineRE         = regexp.MustCompile(`[\x00\r\n]`)
	// TODO(yskopets): no idea how the following regexp correlates with the comment
	// "The formatted string may be destined for a header, and should not contain invalid characters {NUL, LR, CF}."
	startTimeNewlineRE = regexp.MustCompile(`%[-_0^#]*[1-9]*n`)
)

// ValidateFormat validates whether a given format string is valid.
func ValidateFormat(format string) error {
	_, err := ParseFormat(format)
	return err
}

var parser = formatParser{}

// ParseFormat parses a given format string.
//
// The returned object can be used for multiple purposes, i.e.
//  1. To verify that access log format string is valid
//  2. To adjust configuration of `envoy.access_loggers.http_grpc` and `envoy.tcp_grpc_access_log`
//     according to the format string, e.g. to capture additional HTTP headers
//  3. To format a given HTTP or TCP log entry according to the format string
//  4. To bind `%KUMA_*%` placeholders to concrete context-dependent values
//  5. To render back into the format string, e.g.
//     after `%KUMA_*%` placeholders have been bound to concrete context-dependent values
func ParseFormat(format string) (*AccessLogFormat, error) {
	return parser.Parse(format)
}

type formatParser struct{}

func (p formatParser) Parse(format string) (_ *AccessLogFormat, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "format string is not valid")
		}
	}()

	var textLiteralStart = -1
	var fragments []AccessLogFragment

	for pos := 0; pos < len(format); pos++ {
		if format[pos] == '%' {
			if textLiteralStart > -1 {
				fragments = append(fragments, TextSpan(format[textLiteralStart:pos]))
				textLiteralStart = -1
			}

			tail := format[pos:]
			match := commandWithArgsRE.FindStringSubmatch(tail)
			if match == nil {
				return nil, errors.Errorf("expected a command operator to start at position %d, instead got: %q", pos+1, tail)
			}
			token, command, args, limit, err := p.splitMatch(match)
			if err != nil {
				return nil, err
			}
			operator, err := p.parseCommandOperator(token, command, args, limit)
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, operator)
			pos += len(token) - 1
		} else if textLiteralStart < 0 {
			textLiteralStart = pos
		}
	}

	if textLiteralStart >= 0 {
		fragments = append(fragments, TextSpan(format[textLiteralStart:]))
	}

	return &AccessLogFormat{Fragments: fragments}, nil
}

func (p formatParser) splitMatch(match []string) (token string, command string, args string, limit string, err error) {
	if len(match) != 4 {
		return "", "", "", "", errors.Errorf("expected a command operator that consists of a command, args and limit, got %q", match)
	}
	return match[0], match[1], match[2], match[3], nil
}

func (p formatParser) parseCommandOperator(token, command, args, limit string) (AccessLogFragment, error) {
	switch command {
	case CMD_REQ:
		header, altHeader, maxLen, err := p.parseHeaderOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		return &RequestHeaderOperator{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case CMD_RESP:
		header, altHeader, maxLen, err := p.parseHeaderOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		return &ResponseHeaderOperator{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case CMD_TRAILER:
		header, altHeader, maxLen, err := p.parseHeaderOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		return &ResponseTrailerOperator{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case CMD_DYNAMIC_METADATA:
		namespace, path, maxLen, err := p.parseDynamicMetadataOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		return &DynamicMetadataOperator{FilterNamespace: namespace, Path: path, MaxLength: maxLen}, nil
	case CMD_FILTER_STATE:
		key, maxLen, err := p.parseFilterStateOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		return &FilterStateOperator{Key: key, MaxLength: maxLen}, nil
	case CMD_START_TIME:
		format, err := p.parseStartTimeOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		return StartTimeOperator(format), nil
	default:
		field, err := p.parseFieldOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		if CommandOperatorDescriptor(field).IsPlaceholder() {
			return Placeholder(field), nil
		}
		return FieldOperator(field), nil
	}
}

func (p formatParser) parseHeaderOperator(token, command, args, limit string) (header string, altHeader string, maxLen int, err error) {
	if p.hasNoArguments(token, command, args, limit) {
		return "", "", 0, errors.Errorf(`command %q requires a header and optional alternative header names as its arguments, instead got %q`, CommandOperatorDescriptor(command), token)
	}
	header, altHeaders, maxLen, err := p.parseOperator(token, args, limit, "?")
	if err != nil {
		return "", "", 0, err
	}
	if len(altHeaders) > 1 {
		return "", "", 0, errors.Errorf("more than 1 alternative header specified in %q", token)
	}
	if len(altHeaders) > 0 {
		altHeader = altHeaders[0]
	}
	// The main and alternative header should not contain invalid characters {NUL, LR, CF}.
	if newlineRE.MatchString(header) || newlineRE.MatchString(altHeader) {
		return "", "", 0, errors.Errorf("header name contains a newline in %q", token)
	}
	// apparently, Envoy allows both `Header` and `AltHeader` to be empty
	return strings.ToLower(header), strings.ToLower(altHeader), maxLen, nil // Envoy emits log entries with all headers in lower case
}

func (p formatParser) parseDynamicMetadataOperator(token, command, args, limit string) (namespace string, path []string, maxLen int, err error) {
	namespace, path, maxLen, err = p.parseOperator(token, args, limit, ":")
	if err != nil {
		return "", nil, 0, err
	}
	if p.hasNoArguments(token, command, args, limit) {
		return "", nil, 0, errors.Errorf(`command %q requires a filter namespace and optional path as its arguments, instead got %q`, CommandOperatorDescriptor(command), token)
	}
	return namespace, path, maxLen, err
}

func (p formatParser) parseFilterStateOperator(token, command, args, limit string) (key string, maxLen int, err error) {
	key, _, maxLen, err = p.parseOperator(token, args, limit, "")
	if err != nil {
		return "", 0, err
	}
	if p.hasNoArguments(token, command, args, limit) || key == "" {
		return "", 0, errors.Errorf(`command %q requires a key as its argument, instead got %q`, CommandOperatorDescriptor(command), token)
	}
	return key, maxLen, nil
}

func (p formatParser) parseStartTimeOperator(token, command, args, limit string) (format string, err error) {
	// Validate the input specifier here. The formatted string may be destined for a header, and
	// should not contain invalid characters {NUL, LR, CF}.
	if startTimeNewlineRE.MatchString(args) {
		return "", errors.Errorf("start time format string contains a newline in %q", token)
	}
	return args, nil
}

func (p formatParser) parseFieldOperator(token, command, args, limit string) (field string, err error) {
	if token[1:len(token)-1] != command {
		return "", errors.Errorf(`command %q doesn't support arguments or max length constraint, instead got %q`, CommandOperatorDescriptor(command), token)
	}
	return command, nil
}

func (p formatParser) parseOperator(token, args, limit string, separator string) (firstArg string, otherArgs []string, maxLen int, err error) {
	if limit != "" {
		maxLen, err = strconv.Atoi(limit)
		if err != nil {
			return "", nil, 0, errors.Errorf("length must be an integer, instead got %q in %q", limit, token)
		}
	}
	if separator != "" {
		allArgs := strings.Split(args, separator)
		firstArg, otherArgs = allArgs[0], allArgs[1:]
	} else {
		firstArg = args
	}
	return firstArg, otherArgs, maxLen, err
}

func (p formatParser) hasNoArguments(token, command, args, limit string) bool {
	return args == "" &&
		((limit == "" && token[1:len(token)-1] == command) ||
			(limit != "" && token[1:len(token)-1-(len(limit)+1)] == command))
}
