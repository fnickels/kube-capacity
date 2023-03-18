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

package capacity

import (
	"fmt"
	"os"
	"text/tabwriter"
)

const (
	//TableOutput is the constant value for output type table
	TableOutput string = "table"
	//CSVOutput is the constant value for output type csv
	CSVOutput string = "csv"
	//TSVOutput is the constant value for output type csv
	TSVOutput string = "tsv"
	//JSONOutput is the constant value for output type JSON
	JSONOutput string = "json"
	//YAMLOutput is the constant value for output type YAML
	YAMLOutput string = "yaml"
)

// SupportedOutputs returns a string list of output formats supposed by this package
func SupportedOutputs() []string {
	return []string{
		TableOutput,
		CSVOutput,
		TSVOutput,
		JSONOutput,
		YAMLOutput,
	}
}

func printList(cm *clusterMetric,
	showContainers, showPods, showUtil, showPodCount, showNamespace, showAllNodeLabels bool,
	displayNodeLabels, groupByNodeLabels,
	output, sortBy string,
	availableFormat, binpackAnalysis, showPodSummary bool) {

	if output == JSONOutput || output == YAMLOutput {
		lp := &listPrinter{
			cm:                cm,
			showPods:          showPods,
			showUtil:          showUtil,
			showContainers:    showContainers,
			showPodCount:      showPodCount,
			showAllNodeLabels: showAllNodeLabels,
			displayNodeLabels: displayNodeLabels,
			groupByNodeLabels: groupByNodeLabels,
			sortBy:            sortBy,
			binpackAnalysis:   binpackAnalysis,
		}
		lp.Print(output)
	} else if output == TableOutput {
		if showPodSummary {
			pp := &tablePodPrinter{
				cm:                cm,
				showUtil:          showUtil,
				showPodCount:      showPodCount,
				showContainers:    showContainers,
				showNamespace:     showNamespace,
				showAllNodeLabels: showAllNodeLabels,
				displayNodeLabels: displayNodeLabels,
				groupByNodeLabels: groupByNodeLabels,
				sortBy:            sortBy,
				w:                 new(tabwriter.Writer),
				availableFormat:   availableFormat,
				binpackAnalysis:   binpackAnalysis,
			}
			pp.Print()
		} else {
			tp := &tablePrinter{
				cm:                cm,
				showPods:          showPods,
				showUtil:          showUtil,
				showPodCount:      showPodCount,
				showContainers:    showContainers,
				showNamespace:     showNamespace,
				showAllNodeLabels: showAllNodeLabels,
				displayNodeLabels: displayNodeLabels,
				groupByNodeLabels: groupByNodeLabels,
				sortBy:            sortBy,
				w:                 new(tabwriter.Writer),
				availableFormat:   availableFormat,
				binpackAnalysis:   binpackAnalysis,
			}
			tp.Print()
		}
	} else if output == CSVOutput || output == TSVOutput {
		cp := &csvPrinter{
			cm:                cm,
			showPods:          showPods,
			showUtil:          showUtil,
			showPodCount:      showPodCount,
			showContainers:    showContainers,
			showNamespace:     showNamespace,
			showAllNodeLabels: showAllNodeLabels,
			displayNodeLabels: displayNodeLabels,
			groupByNodeLabels: groupByNodeLabels,
			sortBy:            sortBy,
			binpackAnalysis:   binpackAnalysis,
		}
		cp.Print(output)
	} else {
		fmt.Fprintf(os.Stderr, "Called with an unsupported output type: %s", output)
		os.Exit(1)
	}
}
