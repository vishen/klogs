package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/vishen/go-slearch/slearch"
)

// slearchConfig returns a slearch usable config from the conbra command
// line arguments
func getSlearchConfig(cmd *cobra.Command, args []string) slearch.Config {

	m, _ := cmd.Flags().GetStringSlice("match")
	r, _ := cmd.Flags().GetStringSlice("regexp")
	k, _ := cmd.Flags().GetStringSlice("print_keys")
	t, _ := cmd.Flags().GetString("type")
	s, _ := cmd.Flags().GetString("search_type")
	d, _ := cmd.Flags().GetString("key_delimiter")
	v, _ := cmd.Flags().GetBool("verbose")

	config := slearch.Config{}

	makeKVSlice := func(values []string, regex bool) []slearch.KV {
		prevKey := ""
		kvs := make([]slearch.KV, 0, len(m))
		for _, v := range values {
			vSplit := strings.SplitN(v, "=", 2)

			var key string
			var value string

			if len(vSplit) == 1 {
				// TODO(): Should error if `prevKey` is empty
				key = strings.TrimSpace(prevKey)
				value = strings.TrimSpace(vSplit[0])
			} else {
				key = strings.TrimSpace(vSplit[0])
				value = strings.TrimSpace(vSplit[1])
			}
			prevKey = key

			kv := slearch.KV{Key: key}
			if regex {
				kv.RegexString = value
			} else {
				kv.Value = value
			}

			kvs = append(kvs, kv)

		}

		return kvs
	}

	config.MatchOn = makeKVSlice(m, false)
	config.MatchOn = append(config.MatchOn, makeKVSlice(r, true)...)
	config.PrintKeys = k
	config.KeySplitString = d
	config.Verbose = v
	config.LogFormatterType = strings.ToLower(t)
	config.Silence = true

	if strings.ToLower(s) == "or" {
		config.MatchType = slearch.MatchTypeOr
	} else {
		config.MatchType = slearch.MatchTypeAnd
	}

	return config
}
