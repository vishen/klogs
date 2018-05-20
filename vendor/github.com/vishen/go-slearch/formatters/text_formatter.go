package formatters

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/vishen/go-slearch/slearch"
)

func init() {
	slearch.Register("text", NewTextFormatter)
}

type textLogFormatter struct {
	config slearch.Config
}

func NewTextFormatter(config slearch.Config) slearch.StructuredLogFormatter {
	return textLogFormatter{config}
}

func (t textLogFormatter) ValidateLine(line []byte) (bool, []byte) {
	trimmedLine := bytes.TrimSpace(line)

	valid := false

	i := bytes.Index(trimmedLine, []byte{'='})
	if i > 0 && i < len(line)-2 {
		valid = !unicode.IsSpace(rune(line[i-1])) && !unicode.IsSpace(rune(line[i+1]))
	}

	return valid, trimmedLine
}

func (t textLogFormatter) GetValueFromLine(line []byte, key string) (string, error) {

	// Loop through the characters in the `line` and find a matching `key`
	// and it's value, take into account some values might be surronded
	// in ' or " and have multiple spaces in the value.
	pos := 0
	for {
		if pos+len(key) > len(line) {
			break
		}

		if string(line[pos:pos+len(key)]) != key {
			pos += 1
			continue
		}

		// If the next character isn't the '=' then we don't
		// have a match
		if line[pos+len(key)] != '=' {
			pos += 1
			continue
		}

		// Eat '='
		pos += len(key) + 1

		var eatUntil byte = ' '
		switch line[pos] {
		case '"':
			eatUntil = '"'
			pos += 1
		case '\'':
			eatUntil = '\''
			pos += 1
		}
		startPos := pos

		// Eat up until the eatUntil character and return value
		for {
			if pos >= len(line) || line[pos] == eatUntil || line[pos] == '\n' {
				// If we are escaping, ignore
				if line[pos-1] != '\\' {
					break
				}
			}
			pos += 1
		}

		if startPos >= pos || pos > len(line) {
			return "", nil
		}

		return string(line[startPos:pos]), nil
	}

	return "", nil
}

func (t textLogFormatter) FormatFoundValues(valuesFound []slearch.KV) string {
	var buffer bytes.Buffer
	for _, v := range valuesFound {
		buffer.WriteString(fmt.Sprintf("%s=\"%s\" ", v.Key, v.Value))
	}
	return buffer.String()
}

func (t textLogFormatter) AppendValues(line []byte, values []slearch.KV) string {
	return fmt.Sprintf("%s %s", line, t.FormatFoundValues(values))
}
