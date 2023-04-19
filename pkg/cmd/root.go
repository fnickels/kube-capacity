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

	"github.com/fnickels/kube-capacity/pkg/capacity"
	"github.com/spf13/cobra"
)

var criteria capacity.DisplayCriteria

var rootCmd = &cobra.Command{
	Use:   "kube-capacity",
	Short: "kube-capacity provides an overview of the resource requests, limits, and utilization in a Kubernetes cluster.",
	Long:  "kube-capacity provides an overview of the resource requests, limits, and utilization in a Kubernetes cluster.",
	Run: func(cmd *cobra.Command, args []string) {

		if err := cmd.ParseFlags(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if err := validateOutputType(criteria.OutputFormat); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if err := validateSortBy(criteria.SortBy); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if criteria.ShowContainers {
			criteria.ShowPods = true
		}

		capacity.FetchAndPrint(&criteria)
	},
}

func init() {
	criteria.ShowDebug = true

	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowContainers,
		"containers", "c", false, "includes containers in output (forces --pods)")
	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowPods,
		"pods", "p", false, "includes pods in output")
	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowUtil,
		"util", "u", false, "includes resource utilization in output")
	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowPodCount,
		"pod-count", "", false, "includes pod count per node in output")
	rootCmd.PersistentFlags().BoolVarP(&criteria.AvailableFormat,
		"available", "a", false, "includes quantity available instead of percentage used (ignored with csv or tsv output types)")
	rootCmd.PersistentFlags().StringVarP(&criteria.DisplayNodeLabels,
		"display-node-labels", "", "", "comma separated list of node label(s) to display")
	rootCmd.PersistentFlags().StringVarP(&criteria.GroupByNodeLabels,
		"group-by-node-labels", "", "", "comma separated list of node label(s) to group by")
	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowAllNodeLabels,
		"show-all-node-labels", "", false, "show all node labels")

	rootCmd.PersistentFlags().StringVarP(&criteria.SelectPodLabels,
		"select-pod-labels", "", "", "comma separated list of pod label(s) to identify pod families (default: '"+capacity.PodAppNameLabelDefaultSelector+"')")
	rootCmd.PersistentFlags().StringVarP(&criteria.DisplayPodLabels,
		"display-pod-labels", "", "", "comma separated list of node label(s) to display")
	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowAllPodLabels,
		"show-all-pod-labels", "", false, "show all pod labels")

	rootCmd.PersistentFlags().StringVarP(&criteria.Filters.PodLabels,
		"pod-labels", "l", "", "labels to filter pods with")
	rootCmd.PersistentFlags().StringVarP(&criteria.Filters.NodeLabels,
		"node-labels", "", "", "labels to filter nodes with")
	rootCmd.PersistentFlags().StringVarP(&criteria.Filters.NamespaceLabels,
		"namespace-labels", "", "", "labels to filter namespaces with")
	rootCmd.PersistentFlags().StringVarP(&criteria.Filters.Namespace,
		"namespace", "n", "", "only include pods from this namespace")

	rootCmd.PersistentFlags().BoolVarP(&criteria.BinpackAnalysis,
		"binpack-analysis", "b", false, "add node binpack analysis fields")
	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowPodSummary,
		"pod-summary", "", false, "generate alternate report of pods")
	rootCmd.PersistentFlags().BoolVarP(&criteria.ShowDebug,
		"debug", "d", false, "Show debug data")
	rootCmd.PersistentFlags().StringVarP(&criteria.KubeContext,
		"context", "", "", "context to use for Kubernetes config")
	rootCmd.PersistentFlags().StringVarP(&criteria.KubeConfig,
		"kubeconfig", "", "", "kubeconfig file to use for Kubernetes config")
	rootCmd.PersistentFlags().StringVarP(&criteria.SortBy,
		"sort", "", "name",
		fmt.Sprintf("attribute to sort results by (supports: %v)", capacity.SupportedSortAttributes))

	rootCmd.PersistentFlags().StringVarP(&criteria.OutputFormat,
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

func validateSortBy(sortBy string) error {
	for _, attribute := range capacity.SupportedSortAttributes {
		if attribute == sortBy {
			return nil
		}
	}
	return fmt.Errorf("Unsupported Sort expression. We only support: %v", capacity.SupportedSortAttributes)
}
