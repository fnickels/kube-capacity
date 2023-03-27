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
)

type tablePrinter struct {
	cm                        *clusterMetric
	cr                        *DisplayCriteria
	w                         *tabwriter.Writer
	uniqueGroupByNodeLabels   []string
	uniqueDisplayNodeLabels   []string
	uniqueRemainderNodeLabels []string
}

type tableLine struct {
	node            string
	namespace       string
	pod             string
	container       string
	cpuRequests     string
	cpuLimits       string
	cpuUtil         string
	memoryRequests  string
	memoryLimits    string
	memoryUtil      string
	eniRequests     string
	eniLimits       string
	eniUtil         string
	podCount        string
	groupByLabels   []string
	displayLabels   []string
	remainderLabels []string
	binpack         binAnalysis
}

var tableHeaderStrings = tableLine{
	node:            "NODE",
	namespace:       "NAMESPACE",
	pod:             "POD",
	container:       "CONTAINER",
	cpuRequests:     "CPU REQUESTS",
	cpuLimits:       "CPU LIMITS",
	cpuUtil:         "CPU UTIL",
	memoryRequests:  "MEMORY REQUESTS",
	memoryLimits:    "MEMORY LIMITS",
	memoryUtil:      "MEMORY UTIL",
	eniRequests:     "ENI REQUESTS",
	eniLimits:       "ENI LIMITS",
	eniUtil:         "ENI UTIL",
	podCount:        "POD COUNT",
	groupByLabels:   []string{},
	displayLabels:   []string{},
	remainderLabels: []string{},
	binpack:         binHeaders,
}

func PrintTableNodeSummary(cm *clusterMetric, cr *DisplayCriteria) {

	tp := &tablePrinter{
		cm: cm,
		cr: cr,
		w:  new(tabwriter.Writer),
	}

	tp.w.Init(os.Stdout, 0, 8, 2, ' ', 0)

	var err error

	// process Node Label selection elements
	tp.uniqueGroupByNodeLabels,
		tp.uniqueDisplayNodeLabels,
		tp.uniqueRemainderNodeLabels,
		err = processNodeLabelSelections(tp.cm, tp.cr.GroupByNodeLabels, tp.cr.DisplayNodeLabels, tp.cr.ShowAllNodeLabels)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// copy Node Label names to the Header object
	tableHeaderStrings.groupByLabels = tp.uniqueGroupByNodeLabels
	tableHeaderStrings.displayLabels = tp.uniqueDisplayNodeLabels
	tableHeaderStrings.remainderLabels = tp.uniqueRemainderNodeLabels

	// sort first by Group By, then sort criteria
	sortedNodeMetrics := tp.cm.getSortedNodeMetrics(tp.uniqueGroupByNodeLabels, tp.cr.SortBy)

	tp.printLine(&tableHeaderStrings)

	if len(sortedNodeMetrics) > 1 {
		tp.printClusterLine()
	}

	for _, nm := range sortedNodeMetrics {
		if tp.cr.ShowPods || tp.cr.ShowContainers {
			tp.printLine(&tableLine{})
		}

		tp.printNodeLine(nm.name, nm)

		if tp.cr.ShowPods || tp.cr.ShowContainers {
			podMetrics := nm.getSortedPodMetrics(tp.cr.SortBy)
			for _, pm := range podMetrics {
				tp.printPodLine(nm.name, nm, pm)
				if tp.cr.ShowContainers {
					containerMetrics := pm.getSortedContainerMetrics(tp.cr.SortBy)
					for _, containerMetric := range containerMetrics {
						tp.printContainerLine(nm.name, nm, pm, containerMetric)
					}
				}
			}
		}
	}

	err = tp.w.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to table: %s", err)
		os.Exit(1)
	}
}

func (tp *tablePrinter) printLine(tl *tableLine) {
	lineItems := tp.getLineItems(tl)
	fmt.Fprintf(tp.w, strings.Join(lineItems[:], "\t ")+"\n")
}

func (tp *tablePrinter) getLineItems(tl *tableLine) []string {

	lineItems := []string{tl.node}

	// add any 'Group By' Node Labels which are specified here
	for _, x := range tl.groupByLabels {
		lineItems = append(lineItems, x)
	}

	// add any 'Display' Node Labels which are specified here
	for _, x := range tl.displayLabels {
		lineItems = append(lineItems, x)
	}

	if tp.cr.ShowContainers || tp.cr.ShowPods {
		if tp.cr.ShowNamespace() {
			lineItems = append(lineItems, tl.namespace)
		}
		lineItems = append(lineItems, tl.pod)
	}

	if tp.cr.ShowContainers {
		lineItems = append(lineItems, tl.container)
	}

	lineItems = append(lineItems, tl.cpuRequests)
	lineItems = append(lineItems, tl.cpuLimits)

	if tp.cr.ShowUtil {
		lineItems = append(lineItems, tl.cpuUtil)
	}

	lineItems = append(lineItems, tl.memoryRequests)
	lineItems = append(lineItems, tl.memoryLimits)

	if tp.cr.ShowUtil {
		lineItems = append(lineItems, tl.memoryUtil)
	}

	lineItems = append(lineItems, tl.eniRequests)
	lineItems = append(lineItems, tl.eniLimits)

	if tp.cr.ShowUtil {
		lineItems = append(lineItems, tl.eniUtil)
	}

	if tp.cr.ShowPodCount {
		lineItems = append(lineItems, tl.podCount)
	}

	if tp.cr.BinpackAnalysis {
		lineItems = append(lineItems, tl.binpack.idleHeadroom)
		lineItems = append(lineItems, tl.binpack.idleWasteCPU)
		lineItems = append(lineItems, tl.binpack.idleWasteMEM)
		lineItems = append(lineItems, tl.binpack.idleWastePODS)
		lineItems = append(lineItems, tl.binpack.binpackRequestRatio)
		lineItems = append(lineItems, tl.binpack.binpackLimitRatio)
		lineItems = append(lineItems, tl.binpack.binpackUtilizationRatio)
	}

	// if any remaining Node Labels have been specified to be displayed add them here
	for _, x := range tl.remainderLabels {
		lineItems = append(lineItems, x)
	}

	return lineItems
}

func (tp *tablePrinter) printClusterLine() {
	tp.printLine(&tableLine{
		node:            VoidValue,
		namespace:       VoidValue,
		pod:             VoidValue,
		container:       VoidValue,
		cpuRequests:     tp.cm.cpu.requestString(tp.cr),
		cpuLimits:       tp.cm.cpu.limitString(tp.cr),
		cpuUtil:         tp.cm.cpu.utilString(tp.cr),
		memoryRequests:  tp.cm.memory.requestString(tp.cr),
		memoryLimits:    tp.cm.memory.limitString(tp.cr),
		memoryUtil:      tp.cm.memory.utilString(tp.cr),
		eniRequests:     tp.cm.eni.requestString(tp.cr),
		eniLimits:       tp.cm.eni.limitString(tp.cr),
		eniUtil:         tp.cm.eni.utilString(tp.cr),
		podCount:        tp.cm.podCount.podCountString(),
		groupByLabels:   setMultipleVoids(len(tp.uniqueGroupByNodeLabels)),
		displayLabels:   setMultipleVoids(len(tp.uniqueDisplayNodeLabels)),
		remainderLabels: setMultipleVoids(len(tp.uniqueRemainderNodeLabels)),
		binpack:         tp.cm.getBinAnalysis(),
	})
}

func (tp *tablePrinter) printNodeLine(nodeName string, nm *nodeMetric) {
	tp.printLine(&tableLine{
		node:            nodeName,
		namespace:       VoidValue,
		pod:             VoidValue,
		container:       VoidValue,
		cpuRequests:     nm.cpu.requestString(tp.cr),
		cpuLimits:       nm.cpu.limitString(tp.cr),
		cpuUtil:         nm.cpu.utilString(tp.cr),
		memoryRequests:  nm.memory.requestString(tp.cr),
		memoryLimits:    nm.memory.limitString(tp.cr),
		memoryUtil:      nm.memory.utilString(tp.cr),
		eniRequests:     nm.eni.requestString(tp.cr),
		eniLimits:       nm.eni.limitString(tp.cr),
		eniUtil:         nm.eni.utilString(tp.cr),
		podCount:        nm.podCount.podCountString(),
		groupByLabels:   setNodeLabels(tp.uniqueGroupByNodeLabels, nm),
		displayLabels:   setNodeLabels(tp.uniqueDisplayNodeLabels, nm),
		remainderLabels: setNodeLabels(tp.uniqueRemainderNodeLabels, nm),
		binpack:         nm.getBinAnalysis(),
	})
}

func (tp *tablePrinter) printPodLine(nodeName string, nm *nodeMetric, pm *podMetric) {
	tp.printLine(&tableLine{
		node:            nodeName,
		namespace:       pm.namespace,
		pod:             pm.name,
		container:       VoidValue,
		cpuRequests:     pm.cpu.requestString(tp.cr),
		cpuLimits:       pm.cpu.limitString(tp.cr),
		cpuUtil:         pm.cpu.utilString(tp.cr),
		memoryRequests:  pm.memory.requestString(tp.cr),
		memoryLimits:    pm.memory.limitString(tp.cr),
		memoryUtil:      pm.memory.utilString(tp.cr),
		eniRequests:     pm.eni.requestString(tp.cr),
		eniLimits:       pm.eni.limitString(tp.cr),
		eniUtil:         pm.eni.utilString(tp.cr),
		groupByLabels:   setNodeLabels(tp.uniqueGroupByNodeLabels, nm),
		displayLabels:   setNodeLabels(tp.uniqueDisplayNodeLabels, nm),
		remainderLabels: setNodeLabels(tp.uniqueRemainderNodeLabels, nm),
		binpack:         pm.getBinAnalysis(),
	})
}

func (tp *tablePrinter) printContainerLine(nodeName string, nm *nodeMetric, pm *podMetric, cm *containerMetric) {
	tp.printLine(&tableLine{
		node:            nodeName,
		namespace:       pm.namespace,
		pod:             pm.name,
		container:       cm.name,
		cpuRequests:     cm.cpu.requestString(tp.cr),
		cpuLimits:       cm.cpu.limitString(tp.cr),
		cpuUtil:         cm.cpu.utilString(tp.cr),
		memoryRequests:  cm.memory.requestString(tp.cr),
		memoryLimits:    cm.memory.limitString(tp.cr),
		memoryUtil:      cm.memory.utilString(tp.cr),
		eniRequests:     cm.eni.requestString(tp.cr),
		eniLimits:       cm.eni.limitString(tp.cr),
		eniUtil:         cm.eni.utilString(tp.cr),
		groupByLabels:   setNodeLabels(tp.uniqueGroupByNodeLabels, nm),
		displayLabels:   setNodeLabels(tp.uniqueDisplayNodeLabels, nm),
		remainderLabels: setNodeLabels(tp.uniqueRemainderNodeLabels, nm),
		binpack:         cm.getBinAnalysis(),
	})
}
