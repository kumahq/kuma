package accesslog

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	commandWithArgsRE = regexp.MustCompile(`^%(?P<command>[A-Z_]+)(?:\((?P<args>[^\)]*)\))?(?:[:](?P<limit>[0-9]+))?%`)
	newlineRE         = regexp.MustCompile(`[\x00\r\n]`)
	// TODO(yskopets): no idea how the following regexp correlates with the comment
	// "The formatted string may be destined for a header, and should not contain invalid characters {NUL, LR, CF}."
	startTimeNewlineRE = regexp.MustCompile(`%[-_0^#]*[1-9]*n`)
)

var parser = formatParser{}

func ParseFormat(format string) (LogConfigureFormatter, error) {
	return parser.Parse(format)
}

type formatParser struct{}

func (p formatParser) Parse(format string) (_ LogConfigureFormatter, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "format string is not valid")
		}
	}()

	var textLiteralStart = -1
	var formaters []LogEntryFormatter

	for pos := 0; pos < len(format); pos++ {
		if format[pos] == '%' {
			if textLiteralStart > -1 {
				formaters = append(formaters, TextLiteralFormatter(format[textLiteralStart:pos]))
				textLiteralStart = -1
			}

			match := commandWithArgsRE.FindStringSubmatch(format[pos:])
			if match == nil {
				return nil, errors.Errorf("expected a command operator at position %d", pos)
			}
			token, command, args, limit, err := p.splitMatch(match)
			if err != nil {
				return nil, err
			}
			formatter, err := p.parseCommandOperator(token, command, args, limit)
			if err != nil {
				return nil, err
			}
			formaters = append(formaters, formatter)
			pos += len(token) - 1
		} else if textLiteralStart < 0 {
			textLiteralStart = pos
		}
	}

	if textLiteralStart >= 0 {
		formaters = append(formaters, TextLiteralFormatter(format[textLiteralStart:]))
	}

	return &CompositeLogConfigureFormatter{formaters}, nil
}

func (p formatParser) splitMatch(match []string) (token string, command string, args string, limit string, err error) {
	if len(match) != 4 {
		return "", "", "", "", errors.Errorf("expected a command operator that consists of a command, args and limit, got %q", match)
	}
	return match[0], match[1], match[2], match[3], nil
}

func (p formatParser) parseCommandOperator(token, command, args, limit string) (LogEntryFormatter, error) {
	switch command {
	case CMD_REQ:
		header, altHeader, maxLen, err := p.parseHeaderOperator(token, args, limit)
		if err != nil {
			return nil, err
		}
		return &RequestHeaderFormatter{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case CMD_RESP:
		header, altHeader, maxLen, err := p.parseHeaderOperator(token, args, limit)
		if err != nil {
			return nil, err
		}
		return &ResponseHeaderFormatter{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case CMD_TRAILER:
		header, altHeader, maxLen, err := p.parseHeaderOperator(token, args, limit)
		if err != nil {
			return nil, err
		}
		return &ResponseTrailerFormatter{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case CMD_DYNAMIC_METADATA:
		namespace, path, maxLen, err := p.parseDynamicMetadataOperator(token, args, limit)
		if err != nil {
			return nil, err
		}
		return &DynamicMetadataFormatter{FilterNamespace: namespace, Path: path, MaxLength: maxLen}, nil
	case CMD_FILTER_STATE:
		key, maxLen, err := p.parseFilterStateOperator(token, args, limit)
		if err != nil {
			return nil, err
		}
		return &FilterStateFormatter{Key: key, MaxLength: maxLen}, nil
	case CMD_START_TIME:
		format, err := p.parseStartTimeOperator(token, args)
		if err != nil {
			return nil, err
		}
		return StartTimeFormatter(format), nil
	default:
		field, err := p.parseFieldOperator(token, command, args, limit)
		if err != nil {
			return nil, err
		}
		return FieldFormatter(field), nil
	}
}

func (p formatParser) parseHeaderOperator(token, args, limit string) (header string, altHeader string, maxLen int, err error) {
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

func (p formatParser) parseDynamicMetadataOperator(token, args, limit string) (namespace string, path []string, maxLen int, err error) {
	return p.parseOperator(token, args, limit, ":")
}

func (p formatParser) parseFilterStateOperator(token, args, limit string) (key string, maxLen int, err error) {
	key, _, maxLen, err = p.parseOperator(token, args, limit, "")
	if err != nil {
		return "", 0, err
	}
	if key == "" {
		return "", 0, errors.Errorf("expected a non-empty 'key' in the filter state configuration %q", token)
	}
	return key, maxLen, nil
}

func (p formatParser) parseStartTimeOperator(token, args string) (format string, err error) {
	// Validate the input specifier here. The formatted string may be destined for a header, and
	// should not contain invalid characters {NUL, LR, CF}.
	if startTimeNewlineRE.MatchString(args) {
		return "", errors.Errorf("start time format string contains a newline in %q", token)
	}
	return args, nil
}

func (p formatParser) parseFieldOperator(token, command, args, limit string) (field string, err error) {
	if token[1:len(token)-1] != command {
		return "", errors.Errorf(`command "%%%s%%" doesn't support arguments or max length constraint, instead got %q`, command, token)
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
