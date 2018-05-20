package slearch

type MatchType int

const (
	MatchTypeAnd = MatchType(iota)
	MatchTypeOr
)

type KV struct {
	Key          string
	Value        string
	RegexString  string
	KeyExists    bool
	KeyNotExists bool
}

type Config struct {
	// Defines which 'StructuredLogFormatter' to use
	LogFormatterType string

	// Whether this is an AND or OR matching
	MatchType MatchType

	// Values to match on
	MatchOn []KV

	// Which keys to print for matching records
	PrintKeys []string

	// String to split the key on
	KeySplitString string

	// Will print all debug statements
	Verbose bool

	// Silence will stop normal info statements
	Silence bool

	// Extra key values to print and search on
	Extras []KV
}
