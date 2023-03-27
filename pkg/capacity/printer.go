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
)

func printList(cm *clusterMetric, cr *DisplayCriteria) {

	if cr.OutputFormat == JSONOutput || cr.OutputFormat == YAMLOutput {
		PrintList(cm, cr)
	} else if cr.OutputFormat == TableOutput {
		if cr.ShowPodSummary {
			PrintTablePodSummary(cm, cr)
		} else {
			PrintTableNodeSummary(cm, cr)
		}
	} else if cr.OutputFormat == CSVOutput || cr.OutputFormat == TSVOutput {
		PrintCSV(cm, cr)
	} else {
		fmt.Fprintf(os.Stderr, "Called with an unsupported output type: %s", cr.OutputFormat)
		os.Exit(1)
	}
}
