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
	"encoding/json"
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

type listNodeMetric struct {
	Name            string              `json:"name"`
	CPU             *listResourceOutput `json:"cpu,omitempty"`
	Memory          *listResourceOutput `json:"memory,omitempty"`
	Pods            []*listPod          `json:"pods,omitempty"`
	PodCount        string              `json:"podCount,omitempty"`
	NodeLabels      map[string]string   `json:"nodelabels,omitempty"`
	binpackAnalysis binAnalysis         `json:"binpackAnalysis,omitempty"`
}

type listPod struct {
	Name            string              `json:"name"`
	Namespace       string              `json:"namespace"`
	CPU             *listResourceOutput `json:"cpu"`
	Memory          *listResourceOutput `json:"memory"`
	Containers      []listContainer     `json:"containers,omitempty"`
	binpackAnalysis binAnalysis         `json:"binpackAnalysis,omitempty"`
}

type listContainer struct {
	Name   string              `json:"name"`
	CPU    *listResourceOutput `json:"cpu"`
	Memory *listResourceOutput `json:"memory"`
}

type listResourceOutput struct {
	Requests       string `json:"requests"`
	RequestsPct    string `json:"requestsPercent"`
	Limits         string `json:"limits"`
	LimitsPct      string `json:"limitsPercent"`
	Utilization    string `json:"utilization,omitempty"`
	UtilizationPct string `json:"utilizationPercent,omitempty"`
}

type listClusterMetrics struct {
	Nodes         []*listNodeMetric  `json:"nodes"`
	ClusterTotals *listClusterTotals `json:"clusterTotals"`
}

type listClusterTotals struct {
	CPU             *listResourceOutput `json:"cpu"`
	Memory          *listResourceOutput `json:"memory"`
	PodCount        string              `json:"podCount,omitempty"`
	binpackAnalysis binAnalysis         `json:"binpackAnalysis,omitempty"`
}

type listPrinter struct {
	cm                      *clusterMetric
	showPods                bool
	showContainers          bool
	showUtil                bool
	showPodCount            bool
	showAllNodeLabels       bool
	showDebug               bool
	displayNodeLabels       string
	groupByNodeLabels       string
	sortBy                  string
	binpackAnalysis         bool
	uniqueGroupByNodeLabels []string
	uniqueDisplayNodeLabels []string
}

func (lp listPrinter) Print(outputType string) {

	var err error

	// process Node Label selection elements
	lp.uniqueGroupByNodeLabels,
		lp.uniqueDisplayNodeLabels,
		_,
		err = processNodeLabelSelections(lp.cm, lp.groupByNodeLabels, lp.displayNodeLabels, false)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	listOutput := lp.buildListClusterMetrics()

	jsonRaw, err := json.MarshalIndent(listOutput, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error Marshalling JSON: %v\n", err)
		os.Exit(1)
	} else {
		if outputType == JSONOutput {
			fmt.Fprintf(os.Stdout, "%s", jsonRaw)
		} else {
			// This is a strange approach, but the k8s YAML package
			// already marshalls to JSON before converting to YAML,
			// this just allows us to follow the same code path.
			yamlRaw, err := yaml.JSONToYAML(jsonRaw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error Converting JSON to Yaml: %v\n", err)
				os.Exit(1)
			} else {
				fmt.Fprintf(os.Stdout, "%s", yamlRaw)
			}
		}
	}
}

func (lp *listPrinter) buildListClusterMetrics() listClusterMetrics {
	var response listClusterMetrics

	response.ClusterTotals = &listClusterTotals{
		CPU:    lp.buildListResourceOutput(lp.cm.cpu),
		Memory: lp.buildListResourceOutput(lp.cm.memory),
	}

	if lp.showPodCount {
		response.ClusterTotals.PodCount = lp.cm.podCount.podCountString()
	}

	if lp.binpackAnalysis {
		response.ClusterTotals.binpackAnalysis = lp.cm.getBinAnalysis()
	}

	for _, nodeMetric := range lp.cm.getSortedNodeMetrics(lp.uniqueGroupByNodeLabels, lp.sortBy) {
		var node listNodeMetric
		node.Name = nodeMetric.name
		node.CPU = lp.buildListResourceOutput(nodeMetric.cpu)
		node.Memory = lp.buildListResourceOutput(nodeMetric.memory)

		// Node Labels
		if lp.showAllNodeLabels {
			node.NodeLabels = nodeMetric.nodeLabels
		} else {
			mapOfNodeLabels := map[string]string{}
			for _, label := range append(lp.uniqueGroupByNodeLabels, lp.uniqueDisplayNodeLabels...) {
				mapOfNodeLabels[label] = nodeMetric.nodeLabels[label]
			}
			node.NodeLabels = mapOfNodeLabels
		}

		if lp.showPodCount {
			node.PodCount = nodeMetric.podCount.podCountString()
		}

		if lp.binpackAnalysis {
			node.binpackAnalysis = nodeMetric.getBinAnalysis()
		}

		if lp.showPods || lp.showContainers {
			for _, podMetric := range nodeMetric.getSortedPodMetrics(lp.sortBy) {
				var pod listPod
				pod.Name = podMetric.name
				pod.Namespace = podMetric.namespace
				pod.CPU = lp.buildListResourceOutput(podMetric.cpu)
				pod.Memory = lp.buildListResourceOutput(podMetric.memory)

				if lp.binpackAnalysis {
					pod.binpackAnalysis = podMetric.getBinAnalysis()
				}

				if lp.showContainers {
					for _, containerMetric := range podMetric.getSortedContainerMetrics(lp.sortBy) {
						pod.Containers = append(pod.Containers, listContainer{
							Name:   containerMetric.name,
							Memory: lp.buildListResourceOutput(containerMetric.memory),
							CPU:    lp.buildListResourceOutput(containerMetric.cpu),
						})
					}
				}
				node.Pods = append(node.Pods, &pod)
			}
		}
		response.Nodes = append(response.Nodes, &node)
	}

	return response
}

func (lp *listPrinter) buildListResourceOutput(item *resourceMetric) *listResourceOutput {
	valueCalculator := item.valueFunction()
	percentCalculator := item.percentFunctionWithoutDoubleEscape()

	out := listResourceOutput{
		Requests:    valueCalculator(item.request),
		RequestsPct: percentCalculator(item.request),
		Limits:      valueCalculator(item.limit),
		LimitsPct:   percentCalculator(item.limit),
	}

	if lp.showUtil {
		out.Utilization = valueCalculator(item.utilization)
		out.UtilizationPct = percentCalculator(item.utilization)
	}
	return &out
}
