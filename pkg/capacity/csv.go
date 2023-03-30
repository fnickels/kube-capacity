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
	"io"
	"os"
	"strings"
)

type csvPrinter struct {
	cm                        *clusterMetric
	cr                        *DisplayCriteria
	file                      io.Writer
	separator                 string
	uniqueGroupByNodeLabels   []string
	uniqueDisplayNodeLabels   []string
	uniqueRemainderNodeLabels []string
}

type csvLine struct {
	node                     string
	namespace                string
	pod                      string
	podStatus                string
	container                string
	containerType            ContainerClassificationType
	cpuCapacity              string
	cpuRequests              string
	cpuRequestsPercentage    string
	cpuLimits                string
	cpuLimitsPercentage      string
	cpuUtil                  string
	cpuUtilPercentage        string
	memoryCapacity           string
	memoryRequests           string
	memoryRequestsPercentage string
	memoryLimits             string
	memoryLimitsPercentage   string
	memoryUtil               string
	memoryUtilPercentage     string
	podCountCurrent          string
	podCountAllocatable      string
	groupByLabels            []string
	displayLabels            []string
	remainderLabels          []string
	binpack                  binAnalysis
}

var csvHeaderStrings = csvLine{
	node:                     "NODE",
	namespace:                "NAMESPACE",
	pod:                      "POD",
	podStatus:                "POD STATUS",
	container:                "CONTAINER",
	containerType:            "CONTAINER TYPE",
	cpuCapacity:              "CPU CAPACITY (milli)",
	cpuRequests:              "CPU REQUESTS",
	cpuRequestsPercentage:    "CPU REQUESTS %%",
	cpuLimits:                "CPU LIMITS",
	cpuLimitsPercentage:      "CPU LIMITS %%",
	cpuUtil:                  "CPU UTIL",
	cpuUtilPercentage:        "CPU UTIL %%",
	memoryCapacity:           "MEMORY CAPACITY (Mi)",
	memoryRequests:           "MEMORY REQUESTS",
	memoryRequestsPercentage: "MEMORY REQUESTS %%",
	memoryLimits:             "MEMORY LIMITS",
	memoryLimitsPercentage:   "MEMORY LIMITS %%",
	memoryUtil:               "MEMORY UTIL",
	memoryUtilPercentage:     "MEMORY UTIL %%",
	podCountCurrent:          "POD COUNT CURRENT",
	podCountAllocatable:      "POD COUNT ALLOCATABLE",
	groupByLabels:            []string{},
	displayLabels:            []string{},
	remainderLabels:          []string{},
	binpack:                  binHeaders,
}

func PrintCSV(cm *clusterMetric, cr *DisplayCriteria) {

	cp := &csvPrinter{
		cm:        cm,
		cr:        cr,
		file:      os.Stdout,
		separator: ",",
	}

	if cr.OutputFormat == TSVOutput {
		cp.separator = "\t"
	}

	var err error

	// process Node Label selection elements
	cp.uniqueGroupByNodeLabels,
		cp.uniqueDisplayNodeLabels,
		cp.uniqueRemainderNodeLabels,
		err = processNodeLabelSelections(cp.cm, cp.cr.GroupByNodeLabels, cp.cr.DisplayNodeLabels, cp.cr.ShowAllNodeLabels)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// copy Node Label names to the Header object
	csvHeaderStrings.groupByLabels = cp.uniqueGroupByNodeLabels
	csvHeaderStrings.displayLabels = cp.uniqueDisplayNodeLabels
	csvHeaderStrings.remainderLabels = cp.uniqueRemainderNodeLabels

	// sort first by Group By, then sort criteria
	sortedNodeMetrics := cp.cm.getSortedNodeMetrics(cp.uniqueGroupByNodeLabels, cp.cr.SortBy)

	cp.printLine(&csvHeaderStrings)

	if len(sortedNodeMetrics) > 1 {
		cp.printClusterLine()
	}

	for _, nm := range sortedNodeMetrics {
		cp.printNodeLine(nm.name, nm)

		if cp.cr.ShowPods || cp.cr.ShowContainers {
			podMetrics := nm.getSortedPodMetrics(cp.cr.SortBy)
			for _, pm := range podMetrics {
				cp.printPodLine(nm.name, nm, pm)
				if cp.cr.ShowContainers {
					containerMetrics := pm.getSortedContainerMetrics(cp.cr.SortBy)
					for _, containerMetric := range containerMetrics {
						cp.printContainerLine(nm.name, nm, pm, containerMetric)
					}
				}
			}
		}
	}
}

func (cp *csvPrinter) printLine(cl *csvLine) {
	lineItems := cp.getLineItems(cl)
	fmt.Fprintf(cp.file, strings.Join(lineItems[:], cp.separator)+"\n")
}

func (cp *csvPrinter) getLineItems(cl *csvLine) []string {
	lineItems := []string{CSVStringTerminator + cl.node + CSVStringTerminator}

	// add any 'Group By' Node Labels which are specified here
	for _, x := range cl.groupByLabels {
		lineItems = append(lineItems, CSVStringTerminator+x+CSVStringTerminator)
	}

	// add any 'Display' Node Labels which are specified here
	for _, x := range cl.displayLabels {
		lineItems = append(lineItems, CSVStringTerminator+x+CSVStringTerminator)
	}

	if cp.cr.ShowContainers || cp.cr.ShowPods {
		if cp.cr.ShowNamespace() {
			lineItems = append(lineItems, CSVStringTerminator+cl.namespace+CSVStringTerminator)
		}
		lineItems = append(lineItems, CSVStringTerminator+cl.pod+CSVStringTerminator)
		lineItems = append(lineItems, CSVStringTerminator+cl.podStatus+CSVStringTerminator)
	}

	if cp.cr.ShowContainers {
		lineItems = append(lineItems, CSVStringTerminator+cl.container+CSVStringTerminator)
		lineItems = append(lineItems, CSVStringTerminator+string(cl.containerType)+CSVStringTerminator)
	}

	lineItems = append(lineItems, cl.cpuCapacity)
	lineItems = append(lineItems, cl.cpuRequests)
	lineItems = append(lineItems, cl.cpuRequestsPercentage)
	lineItems = append(lineItems, cl.cpuLimits)
	lineItems = append(lineItems, cl.cpuLimitsPercentage)

	if cp.cr.ShowUtil {
		lineItems = append(lineItems, cl.cpuUtil)
		lineItems = append(lineItems, cl.cpuUtilPercentage)
	}

	lineItems = append(lineItems, cl.memoryCapacity)
	lineItems = append(lineItems, cl.memoryRequests)
	lineItems = append(lineItems, cl.memoryRequestsPercentage)
	lineItems = append(lineItems, cl.memoryLimits)
	lineItems = append(lineItems, cl.memoryLimitsPercentage)

	if cp.cr.ShowUtil {
		lineItems = append(lineItems, cl.memoryUtil)
		lineItems = append(lineItems, cl.memoryUtilPercentage)
	}

	if cp.cr.ShowPodCount {
		lineItems = append(lineItems, cl.podCountCurrent)
		lineItems = append(lineItems, cl.podCountAllocatable)
	}

	if cp.cr.BinpackAnalysis {
		lineItems = append(lineItems, cl.binpack.nodesWellUtilized)
		lineItems = append(lineItems, cl.binpack.nodesUnbalanced)
		lineItems = append(lineItems, cl.binpack.nodesUnderutilized)
		lineItems = append(lineItems, cl.binpack.idleHeadroom)
		lineItems = append(lineItems, cl.binpack.idleWasteCPU)
		lineItems = append(lineItems, cl.binpack.idleWasteMEM)
		lineItems = append(lineItems, cl.binpack.idleWastePODS)
		lineItems = append(lineItems, cl.binpack.binpackRequestRatio)
		lineItems = append(lineItems, cl.binpack.binpackLimitRatio)
		lineItems = append(lineItems, cl.binpack.binpackUtilizationRatio)
	}

	// if any remaining Node Labels have been specified to be displayed add them here
	for _, x := range cl.remainderLabels {
		lineItems = append(lineItems, CSVStringTerminator+x+CSVStringTerminator)
	}

	return lineItems
}

func (cp *csvPrinter) printClusterLine() {
	cp.printLine(&csvLine{
		node:                     VoidValue,
		namespace:                VoidValue,
		pod:                      VoidValue,
		container:                VoidValue,
		cpuCapacity:              cp.cm.cpu.capacityString(),
		cpuRequests:              cp.cm.cpu.requestActualString(),
		cpuRequestsPercentage:    cp.cm.cpu.requestPercentageString(),
		cpuLimits:                cp.cm.cpu.limitActualString(),
		cpuLimitsPercentage:      cp.cm.cpu.limitPercentageString(),
		cpuUtil:                  cp.cm.cpu.utilActualString(),
		cpuUtilPercentage:        cp.cm.cpu.utilPercentageString(),
		memoryCapacity:           cp.cm.memory.capacityString(),
		memoryRequests:           cp.cm.memory.requestActualString(),
		memoryRequestsPercentage: cp.cm.memory.requestPercentageString(),
		memoryLimits:             cp.cm.memory.limitActualString(),
		memoryLimitsPercentage:   cp.cm.memory.limitPercentageString(),
		memoryUtil:               cp.cm.memory.utilActualString(),
		memoryUtilPercentage:     cp.cm.memory.utilPercentageString(),
		podCountCurrent:          cp.cm.podCount.podCountCurrentString(),
		podCountAllocatable:      cp.cm.podCount.podCountAllocatableString(),
		groupByLabels:            setMultipleVoids(len(cp.uniqueGroupByNodeLabels)),
		displayLabels:            setMultipleVoids(len(cp.uniqueDisplayNodeLabels)),
		remainderLabels:          setMultipleVoids(len(cp.uniqueRemainderNodeLabels)),
		binpack:                  cp.cm.getBinAnalysis(),
	})
}

func (cp *csvPrinter) printNodeLine(nodeName string, nm *nodeMetric) {
	cp.printLine(&csvLine{
		node:                     nodeName,
		namespace:                VoidValue,
		pod:                      VoidValue,
		container:                VoidValue,
		cpuCapacity:              nm.cpu.capacityString(),
		cpuRequests:              nm.cpu.requestActualString(),
		cpuRequestsPercentage:    nm.cpu.requestPercentageString(),
		cpuLimits:                nm.cpu.limitActualString(),
		cpuLimitsPercentage:      nm.cpu.limitPercentageString(),
		cpuUtil:                  nm.cpu.utilActualString(),
		cpuUtilPercentage:        nm.cpu.utilPercentageString(),
		memoryCapacity:           nm.memory.capacityString(),
		memoryRequests:           nm.memory.requestActualString(),
		memoryRequestsPercentage: nm.memory.requestPercentageString(),
		memoryLimits:             nm.memory.limitActualString(),
		memoryLimitsPercentage:   nm.memory.limitPercentageString(),
		memoryUtil:               nm.memory.utilActualString(),
		memoryUtilPercentage:     nm.memory.utilPercentageString(),
		podCountCurrent:          nm.podCount.podCountCurrentString(),
		podCountAllocatable:      nm.podCount.podCountAllocatableString(),
		groupByLabels:            setNodeLabels(cp.uniqueGroupByNodeLabels, nm),
		displayLabels:            setNodeLabels(cp.uniqueDisplayNodeLabels, nm),
		remainderLabels:          setNodeLabels(cp.uniqueRemainderNodeLabels, nm),
		binpack:                  nm.getBinAnalysis(),
	})
}

func (cp *csvPrinter) printPodLine(nodeName string, nm *nodeMetric, pm *podMetric) {
	cp.printLine(&csvLine{
		node:                     nodeName,
		namespace:                pm.namespace,
		pod:                      pm.name,
		container:                VoidValue,
		cpuCapacity:              pm.cpu.capacityString(),
		cpuRequests:              pm.cpu.requestActualString(),
		cpuRequestsPercentage:    pm.cpu.requestPercentageString(),
		cpuLimits:                pm.cpu.limitActualString(),
		cpuLimitsPercentage:      pm.cpu.limitPercentageString(),
		cpuUtil:                  pm.cpu.utilActualString(),
		cpuUtilPercentage:        pm.cpu.utilPercentageString(),
		memoryCapacity:           pm.memory.capacityString(),
		memoryRequests:           pm.memory.requestActualString(),
		memoryRequestsPercentage: pm.memory.requestPercentageString(),
		memoryLimits:             pm.memory.limitActualString(),
		memoryLimitsPercentage:   pm.memory.limitPercentageString(),
		memoryUtil:               pm.memory.utilActualString(),
		memoryUtilPercentage:     pm.memory.utilPercentageString(),
		groupByLabels:            setNodeLabels(cp.uniqueGroupByNodeLabels, nm),
		displayLabels:            setNodeLabels(cp.uniqueDisplayNodeLabels, nm),
		remainderLabels:          setNodeLabels(cp.uniqueRemainderNodeLabels, nm),
		binpack:                  pm.getBinAnalysis(),
	})
}

func (cp *csvPrinter) printContainerLine(nodeName string, nm *nodeMetric, pm *podMetric, cm *containerMetric) {
	cp.printLine(&csvLine{
		node:                     nodeName,
		namespace:                pm.namespace,
		pod:                      pm.name,
		container:                cm.name,
		cpuCapacity:              cm.cpu.capacityString(),
		cpuRequests:              cm.cpu.requestActualString(),
		cpuRequestsPercentage:    cm.cpu.requestPercentageString(),
		cpuLimits:                cm.cpu.limitActualString(),
		cpuLimitsPercentage:      cm.cpu.limitPercentageString(),
		cpuUtil:                  cm.cpu.utilActualString(),
		cpuUtilPercentage:        cm.cpu.utilPercentageString(),
		memoryCapacity:           cm.memory.capacityString(),
		memoryRequests:           cm.memory.requestActualString(),
		memoryRequestsPercentage: cm.memory.requestPercentageString(),
		memoryLimits:             cm.memory.limitActualString(),
		memoryLimitsPercentage:   cm.memory.limitPercentageString(),
		memoryUtil:               cm.memory.utilActualString(),
		memoryUtilPercentage:     cm.memory.utilPercentageString(),
		groupByLabels:            setNodeLabels(cp.uniqueGroupByNodeLabels, nm),
		displayLabels:            setNodeLabels(cp.uniqueDisplayNodeLabels, nm),
		remainderLabels:          setNodeLabels(cp.uniqueRemainderNodeLabels, nm),
		binpack:                  cm.getBinAnalysis(),
	})
}
