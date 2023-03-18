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
	"strings"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

type tablePodPrinter struct {
	cm                        *clusterMetric
	showUtil                  bool
	showPodCount              bool
	showContainers            bool
	showNamespace             bool
	showAllNodeLabels         bool
	displayNodeLabels         string
	groupByNodeLabels         string
	sortBy                    string
	w                         *tabwriter.Writer
	availableFormat           bool
	binpackAnalysis           bool
	uniquePodLabels           []string
	uniqueGroupByNodeLabels   []string
	uniqueDisplayNodeLabels   []string
	uniqueRemainderNodeLabels []string
}

type tablePodLine struct {
	namespace       string
	pod             string
	container       string
	containerType   string
	cpuRequests     string
	cpuLimits       string
	cpuUtil         string
	memoryRequests  string
	memoryLimits    string
	memoryUtil      string
	podCount        string
	podLabels       []string
	groupByLabels   []string
	displayLabels   []string
	remainderLabels []string
	binpack         binAnalysis
}

var tablePodHeaderStrings = tablePodLine{
	namespace:       "NAMESPACE",
	pod:             "POD",
	container:       "CONTAINER",
	containerType:   "CONTAINER TYPE",
	cpuRequests:     "CPU REQUESTS",
	cpuLimits:       "CPU LIMITS",
	cpuUtil:         "CPU UTIL",
	memoryRequests:  "MEMORY REQUESTS",
	memoryLimits:    "MEMORY LIMITS",
	memoryUtil:      "MEMORY UTIL",
	podCount:        "POD COUNT",
	podLabels:       []string{},
	groupByLabels:   []string{},
	displayLabels:   []string{},
	remainderLabels: []string{},
	binpack:         binHeaders,
}

func (pp *tablePodPrinter) Print() {

	pp.w.Init(os.Stdout, 0, 8, 2, ' ', 0)

	var err error

	// sort pod list (maybe)
	sortedPodList := pp.cm.rawPodList.Items

	pp.printLine(&tablePodHeaderStrings)

	if len(sortedPodList) > 1 {
		pp.printClusterLine()
	}

	for _, pl := range sortedPodList {

		pp.printPodLine(pl)

		for _, cc := range pl.Spec.InitContainers {
			pp.printContainerLine(pl, cc, true)
		}

		for _, cc := range pl.Spec.Containers {
			pp.printContainerLine(pl, cc, false)
		}

		//		if pp.showContainers {
		//			containerMetrics := pm.getSortedContainerMetrics(pp.sortBy)
		//			for _, containerMetric := range containerMetrics {
		//				pp.printContainerLine(nm.name, nm, pm, containerMetric)
		//			}
		//		}
	}

	err = pp.w.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to table: %s", err)
		os.Exit(1)
	}
}

func (pp *tablePodPrinter) printLine(tl *tablePodLine) {
	lineItems := pp.getLineItems(tl)
	fmt.Fprintf(os.Stdout, "LineItems: %v\n", lineItems)
	fmt.Fprintf(pp.w, strings.Join(lineItems[:], "\t ")+"\n")
}

func (pp *tablePodPrinter) getLineItems(tl *tablePodLine) []string {

	lineItems := []string{
		tl.namespace,
		tl.pod,
		tl.container,
		tl.containerType,
	}

	fmt.Fprintf(os.Stdout, "LineItems: %v\n", lineItems)

	//
	//	// add any 'Group By' Node Labels which are specified here
	//	for _, x := range tl.groupByLabels {
	//		lineItems = append(lineItems, x)
	//	}
	//
	//	// add any 'Display' Node Labels which are specified here
	//	for _, x := range tl.displayLabels {
	//		lineItems = append(lineItems, x)
	//	}
	//
	//	if pp.showContainers || pp.showPods {
	//		if pp.showNamespace {
	//			lineItems = append(lineItems, tl.namespace)
	//		}
	//		lineItems = append(lineItems, tl.pod)
	//	}
	//
	//	if pp.showContainers {
	//		lineItems = append(lineItems, tl.container)
	//	}
	//
	//	lineItems = append(lineItems, tl.cpuRequests)
	//	lineItems = append(lineItems, tl.cpuLimits)
	//
	//	if pp.showUtil {
	//		lineItems = append(lineItems, tl.cpuUtil)
	//	}
	//
	//	lineItems = append(lineItems, tl.memoryRequests)
	//	lineItems = append(lineItems, tl.memoryLimits)
	//
	//	if pp.showUtil {
	//		lineItems = append(lineItems, tl.memoryUtil)
	//	}
	//
	//	if pp.showPodCount {
	//		lineItems = append(lineItems, tl.podCount)
	//	}
	//
	//	if pp.binpackAnalysis {
	//		lineItems = append(lineItems, tl.binpack.idleHeadroom)
	//		lineItems = append(lineItems, tl.binpack.idleWasteCPU)
	//		lineItems = append(lineItems, tl.binpack.idleWasteMEM)
	//		lineItems = append(lineItems, tl.binpack.idleWastePODS)
	//		lineItems = append(lineItems, tl.binpack.binpackRequestRatio)
	//		lineItems = append(lineItems, tl.binpack.binpackLimitRatio)
	//		lineItems = append(lineItems, tl.binpack.binpackUtilizationRatio)
	//	}
	//
	//	// if any remaining Node Labels have been specified to be displayed add them here
	//	for _, x := range tl.remainderLabels {
	//		lineItems = append(lineItems, x)
	//	}
	//
	return lineItems
}

func (pp *tablePodPrinter) printClusterLine() {
	pp.printLine(&tablePodLine{
		//		namespace:       VoidValue,
		//		pod:             VoidValue,
		//		container:       VoidValue,
		//		cpuRequests:     pp.cm.cpu.requestString(pp.availableFormat),
		//		cpuLimits:       pp.cm.cpu.limitString(pp.availableFormat),
		//		cpuUtil:         pp.cm.cpu.utilString(pp.availableFormat),
		//		memoryRequests:  pp.cm.memory.requestString(pp.availableFormat),
		//		memoryLimits:    pp.cm.memory.limitString(pp.availableFormat),
		//		memoryUtil:      pp.cm.memory.utilString(pp.availableFormat),
		//		podCount:        pp.cm.podCount.podCountString(),
		//		groupByLabels:   setMultipleVoids(len(pp.uniqueGroupByNodeLabels)),
		//		displayLabels:   setMultipleVoids(len(pp.uniqueDisplayNodeLabels)),
		//		remainderLabels: setMultipleVoids(len(pp.uniqueRemainderNodeLabels)),
		//		binpack:         pp.cm.getBinAnalysis(),
	})
}

func (pp *tablePodPrinter) printPodLine(pl corev1.Pod) {

	req, limit := resourcehelper.PodRequestsAndLimits(&pl)

	fmt.Fprintf(os.Stdout, "Request : %v\n", req)
	fmt.Fprintf(os.Stdout, "Limit   : %v\n", limit)
	if pl.Spec.Overhead != nil {
		fmt.Fprintf(os.Stdout, "Overhead: %v\n", pl.Spec.Overhead)
	}
	pp.printLine(&tablePodLine{
		namespace:     pl.GetNamespace(),
		pod:           pl.GetName(),
		container:     VoidValue,
		containerType: VoidValue,
		//		cpuRequests:     p.cpu.requestString(pp.availableFormat),
		//		cpuLimits:       pm.cpu.limitString(pp.availableFormat),
		//		cpuUtil:         pm.cpu.utilString(pp.availableFormat),
		//		memoryRequests:  pm.memory.requestString(pp.availableFormat),
		//		memoryLimits:    pm.memory.limitString(pp.availableFormat),
		//		memoryUtil:      pm.memory.utilString(pp.availableFormat),
		//		groupByLabels:   setNodeLabels(pp.uniqueGroupByNodeLabels, nm),
		//		displayLabels:   setNodeLabels(pp.uniqueDisplayNodeLabels, nm),
		//		remainderLabels: setNodeLabels(pp.uniqueRemainderNodeLabels, nm),
		// binpack: pl.getBinAnalysis(),
	})

}

func (pp *tablePodPrinter) printContainerLine(pl corev1.Pod, cl corev1.Container, isInitContainer bool) {

	containerType := "normal"
	if isInitContainer {
		containerType = "init"
	}

	pp.printLine(&tablePodLine{
		namespace:     pl.GetNamespace(),
		pod:           pl.GetName(),
		container:     cl.Name,
		containerType: containerType,

		//		cpuRequests:     cm.cpu.requestString(pp.availableFormat),
		//		cpuLimits:       cm.cpu.limitString(pp.availableFormat),
		//		cpuUtil:         cm.cpu.utilString(pp.availableFormat),
		//		memoryRequests:  cm.memory.requestString(pp.availableFormat),
		//		memoryLimits:    cm.memory.limitString(pp.availableFormat),
		//		memoryUtil:      cm.memory.utilString(pp.availableFormat),
		//		groupByLabels:   setNodeLabels(pp.uniqueGroupByNodeLabels, nm),
		//		displayLabels:   setNodeLabels(pp.uniqueDisplayNodeLabels, nm),
		//		remainderLabels: setNodeLabels(pp.uniqueRemainderNodeLabels, nm),
		//		binpack:         cm.getBinAnalysis(),
	})
}
