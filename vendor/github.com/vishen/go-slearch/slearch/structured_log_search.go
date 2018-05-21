package slearch

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sync"

	"github.com/pkg/errors"
)

var (
	// Common errors
	ErrNoMatchingKeyValues    = errors.New("no matching key values found")
	ErrNoMatchingPrintValues  = errors.New("no matching print values found")
	ErrInvalidFormat          = errors.New("invalid formatter format")
	ErrNoFormattersRegistered = errors.New("no formatters registered")
)

func isSoftError(err error) bool {
	return err == ErrNoMatchingKeyValues || err == ErrNoMatchingPrintValues
}

type StructuredLogFormatter interface {
	ValidateLine(line []byte) (bool, []byte)
	GetValueFromLine(line []byte, key string) (string, error)
	FormatFoundValues(valuesFound []KV) string
	AppendValues(line []byte, values []KV) string
}

func StructuredLoggingSearch(config Config, in io.Reader, out io.Writer) error {

	var formatters []RegisterFunc

	if config.LogFormatterType == "" {
		formatters = GetAllFormatters()
	} else {
		formatter, ok := getFormatter(config.LogFormatterType)
		if !ok {
			return errors.Errorf("no formatter for '%s' found", config.LogFormatterType)
		}
		formatters = []RegisterFunc{formatter}
	}

	if len(formatters) == 0 {
		return ErrNoFormattersRegistered
	}

	type lineResult struct {
		lineNumber uint64
		original   string
		result     string
		err        error
	}
	resultsChan := make(chan lineResult)

	doneChan := make(chan bool, 1)

	go func() {
		receivedLineResults := map[uint64]lineResult{}
		currentLineNumber := uint64(0)
		foundResults := false

		for lr := range resultsChan {
			receivedLineResults[lr.lineNumber] = lr

			for {
				foundLineResult, ok := receivedLineResults[currentLineNumber]
				if !ok {
					break
				}
				if foundLineResult.result != "" {
					fmt.Fprintln(out, foundLineResult.result)
					foundResults = true
				} else {
					err := foundLineResult.err
					if err != nil {
						if config.Verbose || (!isSoftError(err) && config.LogFormatterType != "") {
							fmt.Fprintf(out, "Error on line %d: %s: %s\n", foundLineResult.lineNumber, err, foundLineResult.original)
						}
					}
				}
				currentLineNumber++
			}
		}

		if !foundResults && !config.Silence {
			fmt.Fprintln(out, "no results found")
		}

		doneChan <- true

	}()

	reader := bufio.NewReader(in)

	// TODO(vishen): Allow configuration to be able to use a max number
	// of goroutines
	wg := sync.WaitGroup{}

	for i := uint64(0); ; i++ {
		text, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		// Strip the \n from the line
		text = text[:len(text)-1]

		// if text is empty just continue
		if len(text) == 0 {
			continue
		}

		wg.Add(1)
		go func(lineNumber uint64, line []byte) {
			defer wg.Done()
			lr := lineResult{
				original:   string(line),
				lineNumber: lineNumber,
			}
			// TODO(vishen): When a formatter is first found, always try that one first
			for _, formatter := range formatters {
				result, err := SearchLine(config, line, formatter)
				if err == nil {
					lr.result = result
					break
				}
				lr.err = err
			}
			resultsChan <- lr
		}(i, text)

	}

	wg.Wait()
	close(resultsChan)

	<-doneChan

	return nil
}

func SearchLine(config Config, line []byte, formatterFunc RegisterFunc) (string, error) {

	formatter := formatterFunc(config)

	validLine, line := formatter.ValidateLine(line)
	if !validLine {
		return "", ErrInvalidFormat
	}

	valuesToPrint := make([]KV, 0, len(config.PrintKeys))

	found := false
	for _, v := range config.MatchOn {
		foundValue, err := formatter.GetValueFromLine(line, v.Key)
		if err != nil {
			return "", err
		}

		if foundValue == "" {
			// If not found in line, check to see if it is in extras
			for _, e := range config.Extras {
				if e.Key == v.Key {
					foundValue = e.Value
					break
				}
			}

			if foundValue == "" {
				continue
			}
		}

		matched := false
		if v.KeyExists && foundValue != "" {
			matched = true
		} else {
			if v.Value != "" {
				matched = foundValue == v.Value
			} else if v.RegexString != "" {
				matched, _ = regexp.MatchString(v.RegexString, foundValue)
			}
		}

		if !matched && config.MatchType == MatchTypeAnd {
			return "", ErrNoMatchingKeyValues
		}

		if matched {
			found = matched
		}

	}

	if !found && len(config.MatchOn) > 0 {
		return "", ErrNoMatchingKeyValues
	}

	for _, pk := range config.PrintKeys {
		pkv, err := formatter.GetValueFromLine(line, pk)
		if err != nil {
			if pkv == "" {
				return "", err
			}
		}
		// If not found in line, check to see if it is in Extras
		for _, e := range config.Extras {
			if e.Key == pk {
				pkv = e.Value
				break
			}
		}
		if pkv == "" {
			continue
		}
		valuesToPrint = append(valuesToPrint, KV{Key: pk, Value: fmt.Sprintf("%s", pkv)})
	}

	// TODO(vishen): It is possible to have config.printKeys that don't match
	// any line, this should NOT print the entire line? Currently it kind of
	// seems alright to default to printing the line if no matching valuesToPrint
	// are found.
	if len(valuesToPrint) == 0 {
		if len(config.PrintKeys) == 0 {
			// TODO(vishen): Do something better for when we have extra values to add
			// to the line, but we don't know the format - maybe add a new formatter method
			// that will parse the line and combine it with the extras if possible
			if len(config.Extras) > 0 {
				return formatter.AppendValues(line, config.Extras), nil
			}
			return string(line), nil

		}
		return "", ErrNoMatchingPrintValues
	} else {
		// Add any extra values
		valuesToPrint = append(valuesToPrint, config.Extras...)
		return formatter.FormatFoundValues(valuesToPrint), nil
	}
}
