// Copyright 2019 Kube Capacity Authors
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

	"github.com/robscott/kube-capacity/pkg/capacity"
	"github.com/spf13/cobra"
)

var (
	showContainers    bool
	showPods          bool
	showUtil          bool
	showPodCount      bool
	showDebug         bool
	displayNodeLabels string
	groupByNodeLabels string
	showAllNodeLabels bool
	podLabels         string
	nodeLabels        string
	namespaceLabels   string
	namespace         string
	kubeContext       string
	kubeConfig        string
	outputFormat      string
	sortBy            string
	availableFormat   bool
	binpackAnalysis   bool
	showPodSummary    bool
)

var rootCmd = &cobra.Command{
	Use:   "kube-capacity",
	Short: "kube-capacity provides an overview of the resource requests, limits, and utilization in a Kubernetes cluster.",
	Long:  "kube-capacity provides an overview of the resource requests, limits, and utilization in a Kubernetes cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.ParseFlags(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if err := validateOutputType(outputFormat); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if showContainers {
			showPods = true
		}

		capacity.FetchAndPrint(
			showContainers, showPods, showUtil, showPodCount, showAllNodeLabels,
			availableFormat, binpackAnalysis, showPodSummary, showDebug,
			podLabels, nodeLabels, displayNodeLabels, groupByNodeLabels,
			namespaceLabels, namespace, kubeContext, kubeConfig, outputFormat, sortBy)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&showContainers,
		"containers", "c", false, "includes containers in output (forces --pods)")
	rootCmd.PersistentFlags().BoolVarP(&showPods,
		"pods", "p", false, "includes pods in output")
	rootCmd.PersistentFlags().BoolVarP(&showUtil,
		"util", "u", false, "includes resource utilization in output")
	rootCmd.PersistentFlags().BoolVarP(&showPodCount,
		"pod-count", "", false, "includes pod count per node in output")
	rootCmd.PersistentFlags().BoolVarP(&availableFormat,
		"available", "a", false, "includes quantity available instead of percentage used (ignored with csv or tsv output types)")
	rootCmd.PersistentFlags().StringVarP(&podLabels,
		"pod-labels", "l", "", "labels to filter pods with")
	rootCmd.PersistentFlags().StringVarP(&displayNodeLabels,
		"display-node-labels", "", "", "comma separated list of node label(s) to display")
	rootCmd.PersistentFlags().StringVarP(&groupByNodeLabels,
		"group-by-node-labels", "", "", "comma separated list of node label(s) to group by")
	rootCmd.PersistentFlags().BoolVarP(&showAllNodeLabels,
		"show-all-labels", "", false, "show all node labels")
	rootCmd.PersistentFlags().StringVarP(&nodeLabels,
		"node-labels", "", "", "labels to filter nodes with")
	rootCmd.PersistentFlags().StringVarP(&namespaceLabels,
		"namespace-labels", "", "", "labels to filter namespaces with")
	rootCmd.PersistentFlags().BoolVarP(&binpackAnalysis,
		"binpack-analysis", "b", false, "add node binpack analysis fields")
	rootCmd.PersistentFlags().BoolVarP(&showPodSummary,
		"pod-summary", "", false, "generate alternate report of pods")
	rootCmd.PersistentFlags().BoolVarP(&showDebug,
		"debug", "d", false, "Show debug data")
	rootCmd.PersistentFlags().StringVarP(&namespace,
		"namespace", "n", "", "only include pods from this namespace")
	rootCmd.PersistentFlags().StringVarP(&kubeContext,
		"context", "", "", "context to use for Kubernetes config")
	rootCmd.PersistentFlags().StringVarP(&kubeConfig,
		"kubeconfig", "", "", "kubeconfig file to use for Kubernetes config")
	rootCmd.PersistentFlags().StringVarP(&sortBy,
		"sort", "", "name",
		fmt.Sprintf("attribute to sort results by (supports: %v)", capacity.SupportedSortAttributes))

	rootCmd.PersistentFlags().StringVarP(&outputFormat,
		"output", "o", capacity.TableOutput,
		fmt.Sprintf("output format for information (supports: %v)", capacity.SupportedOutputs()))
}

// Execute is the primary entrypoint for this CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func validateOutputType(outputType string) error {
	for _, format := range capacity.SupportedOutputs() {
		if format == outputType {
			return nil
		}
	}
	return fmt.Errorf("Unsupported Output Type. We only support: %v", capacity.SupportedOutputs())
}
