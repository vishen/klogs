package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/vishen/go-slearch/slearch"
)

func init() {
	slearch.Register("json", NewJSONFormatter)
}

type jsonLogFormatter struct {
	config slearch.Config
}

func NewJSONFormatter(config slearch.Config) slearch.StructuredLogFormatter {
	return jsonLogFormatter{config}
}

func (j jsonLogFormatter) ValidateLine(line []byte) (bool, []byte) {
	trimedLine := bytes.TrimSpace(line)
	return trimedLine[0] == '{' && trimedLine[len(trimedLine)-1] == '}', trimedLine
}

func (j jsonLogFormatter) GetValueFromLine(line []byte, key string) (string, error) {
	keySplit := searchableKey(key, j.config.KeySplitString)
	vs, _, _, _ := jsonparser.Get(line, keySplit...)
	return fmt.Sprintf("%s", vs), nil
}

func (j jsonLogFormatter) FormatFoundValues(valuesFound []slearch.KV) string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i, v := range valuesFound {
		buffer.WriteString(fmt.Sprintf("\"%s\":\"%s\"", v.Key, v.Value))
		if i != len(valuesFound)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()

}

func (j jsonLogFormatter) AppendValues(line []byte, values []slearch.KV) string {
	var buffer bytes.Buffer

	buffer.Write(line[:len(line)-2])
	buffer.WriteString(", ")

	for i, v := range values {
		buffer.WriteString(fmt.Sprintf("\"%s\":\"%s\"", v.Key, v.Value))
		if i != len(values)-1 {
			buffer.WriteString(", ")
		}
	}

	buffer.WriteString("}")
	return buffer.String()

}

func searchableKey(key, splitKeyOnString string) []string {
	if splitKeyOnString == "" {
		return []string{key}
	}
	return strings.Split(key, splitKeyOnString)
}
