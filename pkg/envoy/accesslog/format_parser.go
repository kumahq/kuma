package accesslog

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	commandWithArgsRE  = regexp.MustCompile(`%([A-Z]|_)+(\([^\)]*\))?(:[0-9]+)?(%)`)
	newlineRE          = regexp.MustCompile(`\n`)
	startTimeNewlineRE = regexp.MustCompile(`%[-_0^#]*[1-9]*n`)
)

const (
	requestHeaderPrefix   = "REQ("
	responseHeaderPrefix  = "RESP("
	responseTrailerPrefix = "TRAILER("
	dynamicMetadataPrefix = "DYNAMIC_METADATA("
	filterStatePrefix     = "FILTER_STATE("
	startTimePrefix       = "START_TIME"

	requestHeaderPrefixLen   = len(requestHeaderPrefix)
	responseHeaderPrefixLen  = len(responseHeaderPrefix)
	responseTrailerPrefixLen = len(responseTrailerPrefix)
	dynamicMetadataPrefixLen = len(dynamicMetadataPrefix)
	filterStatePrefixLen     = len(filterStatePrefix)
	startTimePrefixLen       = len(startTimePrefix)
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
			if textLiteralStart >= 0 {
				formaters = append(formaters, TextLiteralFormatter(format[textLiteralStart:pos]))
				textLiteralStart = -1
			}

			match := string(commandWithArgsRE.Find([]byte(format[pos:])))
			if len(match) == 0 {
				return nil, errors.Errorf("expected a command operator at position %d", pos)
			}
			token := match[1 : len(match)-1]
			formatter, err := p.parseFormatToken(token)
			if err != nil {
				return nil, err
			}
			formaters = append(formaters, formatter)
			pos += len(match) - 1
		} else if textLiteralStart < 0 {
			textLiteralStart = pos
		}
	}

	if textLiteralStart >= 0 {
		formaters = append(formaters, TextLiteralFormatter(format[textLiteralStart:]))
	}

	return &CompositeLogConfigureFormatter{formaters}, nil
}

func (p formatParser) parseFormatToken(token string) (LogEntryFormatter, error) {
	switch {
	case strings.HasPrefix(token, requestHeaderPrefix):
		header, altHeader, maxLen, err := p.parseCommandHeader(token, requestHeaderPrefixLen)
		if err != nil {
			return nil, err
		}
		return &RequestHeaderFormatter{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case strings.HasPrefix(token, responseHeaderPrefix):
		header, altHeader, maxLen, err := p.parseCommandHeader(token, responseHeaderPrefixLen)
		if err != nil {
			return nil, err
		}
		return &ResponseHeaderFormatter{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case strings.HasPrefix(token, responseTrailerPrefix):
		header, altHeader, maxLen, err := p.parseCommandHeader(token, responseTrailerPrefixLen)
		if err != nil {
			return nil, err
		}
		return &ResponseTrailerFormatter{HeaderFormatter{Header: header, AltHeader: altHeader, MaxLength: maxLen}}, nil
	case strings.HasPrefix(token, dynamicMetadataPrefix):
		namespace, path, maxLen, err := p.parseCommand(token, dynamicMetadataPrefixLen, ":")
		if err != nil {
			return nil, err
		}
		return &DynamicMetadataFormatter{FilterNamespace: namespace, Path: path, MaxLength: maxLen}, nil
	case strings.HasPrefix(token, filterStatePrefix):
		key, maxLen, err := p.parseFilterState(token, filterStatePrefixLen)
		if err != nil {
			return nil, err
		}
		return &FilterStateFormatter{Key: key, MaxLength: maxLen}, nil
	case strings.HasPrefix(token, startTimePrefix):
		args, err := p.parseStartTime(token, startTimePrefixLen)
		if err != nil {
			return nil, err
		}
		return StartTimeFormatter(args), nil
	default:
		return FieldFormatter(token), nil
	}
}

func (p formatParser) parseCommandHeader(token string, prefixLen int) (header string, altHeader string, maxLen int, err error) {
	header, altHeaders, maxLen, err := p.parseCommand(token, prefixLen, "?")
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
	if newlineRE.Match([]byte(header)) || newlineRE.Match([]byte(altHeader)) {
		return "", "", 0, errors.Errorf("header name contains a newline in %q", token)
	}
	return header, altHeader, maxLen, nil
}

func (p formatParser) parseCommand(token string, prefixLen int, separator string) (name string, extraNames []string, maxLen int, err error) {
	bracketPos := strings.Index(token, ")")
	if bracketPos == -1 {
		return "", nil, 0, errors.Errorf("expected ')' in the end of %q", token)
	}
	if bracketPos != len(token)-1 {
		// Closing bracket should be either last one or followed by ':' to denote limitation.
		if token[bracketPos+1] != ':' {
			return "", nil, 0, errors.Errorf("incorrect position of ')' in %q", token)
		}

		maxLenArg := token[bracketPos+2:]
		maxLen, err = strconv.Atoi(maxLenArg)
		if err != nil {
			return "", nil, 0, errors.Errorf("length must be an integer, instead got %q in %q", maxLenArg, token)
		}
	}
	nameArg := token[prefixLen:bracketPos]
	if separator != "" {
		names := strings.Split(nameArg, separator)
		name = names[0]
		extraNames = names[1:]
	} else {
		name = nameArg
	}
	return name, extraNames, maxLen, err
}

func (p formatParser) parseFilterState(token string, prefixLen int) (key string, maxLen int, err error) {
	key, _, maxLen, err = p.parseCommand(token, prefixLen, "")
	if err != nil {
		return "", 0, err
	}
	if key == "" {
		return "", 0, errors.Errorf("expected a non-empty 'key' in the filter state configuration %q", token)
	}
	return key, maxLen, nil
}

func (p formatParser) parseStartTime(token string, prefixLen int) (args string, err error) {
	if len(token) != prefixLen {
		if token[prefixLen] != '(' {
			return "", errors.Errorf("expected a '(' in %q at position %d", token, prefixLen)
		}
		if token[len(token)-1] != ')' {
			return "", errors.Errorf("expected a ')' in %q at position %d", token, len(token)-1)
		}
		args = token[prefixLen+1 : len(token)-1]
	}
	// Validate the input specifier here. The formatted string may be destined for a header, and
	// should not contain invalid characters {NUL, LR, CF}.
	if startTimeNewlineRE.Match([]byte(args)) {
		return "", errors.Errorf("start time format string contains a newline in %q", token)
	}
	return args, nil
}
