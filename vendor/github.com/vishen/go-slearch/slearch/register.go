package slearch

import "sync"

var (

	// Map of structured log formatter funcs
	formatters   = map[string]RegisterFunc{}
	formattersMu = sync.Mutex{}
)

type RegisterFunc func(Config) StructuredLogFormatter

func Register(key string, registerFunc RegisterFunc) {
	formattersMu.Lock()
	formatters[key] = registerFunc
	formattersMu.Unlock()
}

func GetAllFormatters() []RegisterFunc {
	formattersList := make([]RegisterFunc, len(formatters))
	i := 0
	for _, f := range formatters {
		formattersList[i] = f
		i++
	}
	return formattersList
}

func getFormatter(key string) (RegisterFunc, bool) {
	formattersMu.Lock()
	defer formattersMu.Unlock()

	slf, ok := formatters[key]
	return slf, ok
}
