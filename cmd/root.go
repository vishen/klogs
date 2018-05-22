// Copyright © 2018 Jonathan Pentecost <pentecostjonathan@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "klogs",
	Short: "Kubernetes earch structured logs",
	Long:  `Read stuctured logs from Kubernetes and filter out lines based on exact or regex matches. Currently only supports JSON and text logs.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting klogs...")
		logs(cmd, args)
		fmt.Println("Done")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("kubeconfig", "", "Path to kubernetes config")
	rootCmd.Flags().String("kubecontext", "", "Kubernetes context to use")
	rootCmd.Flags().StringP("namespace", "n", "", "the kubernetes namespace to filter on")
	rootCmd.Flags().StringP("selector", "l", "", "kubernetes selector (label query) to filter on. eg: app=api")
	rootCmd.Flags().StringSliceP("containers", "c", []string{}, "kubernetes selector (label query) to filter on")

	rootCmd.Flags().StringP("type", "t", "", "the log type to use: 'json' or 'text'. If unspecified it will attempt to use all log types")
	rootCmd.Flags().StringP("search_type", "s", "and", "the search type to use: 'and' or 'or'")
	rootCmd.Flags().StringP("key_delimiter", "d", "", "the string to split the key on for complex key queries")
	rootCmd.Flags().StringSliceP("match", "m", []string{}, "key and value to match on. eg: label1=value1")
	rootCmd.Flags().StringSliceP("regexp", "r", []string{}, "key and value to regex match on. eg: label1=value*")
	rootCmd.Flags().StringSliceP("key_exists", "e", []string{}, "print lines that have these keys")
	rootCmd.Flags().StringSliceP("print_keys", "p", []string{}, "keys to print if a match is found")
	rootCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
