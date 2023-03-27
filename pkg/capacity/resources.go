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
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// SupportedSortAttributes lists the valid sorting options
var SupportedSortAttributes = [...]string{
	"cpu.util",
	"cpu.request",
	"cpu.limit",
	"mem.util",
	"mem.request",
	"mem.limit",
	"cpu.util.percentage",
	"cpu.request.percentage",
	"cpu.limit.percentage",
	"mem.util.percentage",
	"mem.request.percentage",
	"mem.limit.percentage",
	"name",
}

// Mebibyte represents the number of bytes in a mebibyte.
const Mebibyte = 1024 * 1024

const CPU = "cpu"
const Memory = "memory"
const ENI = "vpc.amazonaws.com/pod-eni"

type resourceMetric struct {
	resourceType string
	allocatable  resource.Quantity
	utilization  resource.Quantity
	request      resource.Quantity
	limit        resource.Quantity
}

type clusterMetric struct {
	cpu                  *resourceMetric
	memory               *resourceMetric
	eni                  *resourceMetric
	nodeMetrics          map[string]*nodeMetric
	podMetrics           map[string]*podMetric
	podCount             *podCount
	rawPodList           []corev1.Pod
	podAppSelectorLabels []string
	rawPodAppList        []podAppSummary
}

type nodeMetric struct {
	name       string
	cpu        *resourceMetric
	memory     *resourceMetric
	eni        *resourceMetric
	podMetrics map[string]*podMetric
	podCount   *podCount
	nodeLabels map[string]string
}

type podMetric struct {
	name             string
	namespace        string
	node             string
	cpu              *resourceMetric
	memory           *resourceMetric
	eni              *resourceMetric
	containerMetrics map[string]*containerMetric
}

type containerMetric struct {
	name   string
	cpu    *resourceMetric
	memory *resourceMetric
	eni    *resourceMetric
}

type podCount struct {
	current     int64
	allocatable int64
}

type podAppSummary struct {
	appNameKey        string
	appNameLabel      string
	podCount          int64
	specialNoLabelSet bool
	Items             []corev1.Pod
	cpu               *resourceMetric
	memory            *resourceMetric
	eni               *resourceMetric
}

type resourceSummary struct {
	resourceType   string
	utilizationMin resource.Quantity
	utilizationMax resource.Quantity
	utilizationSum resource.Quantity
	requestMin     resource.Quantity
	requestMax     resource.Quantity
	requestSum     resource.Quantity
	limitMin       resource.Quantity
	limitMax       resource.Quantity
	limitSum       resource.Quantity
	count          int
}

func buildClusterMetric(podList *corev1.PodList, pmList *v1beta1.PodMetricsList,
	nodeList *corev1.NodeList, nmList *v1beta1.NodeMetricsList, cr *DisplayCriteria) clusterMetric {

	// get summary list of pods
	appList, appKeys := getPodAppsList(podList, cr.SelectPodLabels)

	// initialize cluster metric
	cm := clusterMetric{
		cpu:         &resourceMetric{resourceType: CPU},
		memory:      &resourceMetric{resourceType: Memory},
		eni:         &resourceMetric{resourceType: ENI},
		nodeMetrics: map[string]*nodeMetric{},
		podMetrics:  map[string]*podMetric{},
		podCount:    &podCount{},
		// pod summary elements
		rawPodList:           podList.Items,
		podAppSelectorLabels: appKeys,
		rawPodAppList:        appList,
	}

	// Iterate over nodes and establish node allocation limits and pod count
	var totalPodAllocatable int64
	var totalPodCurrent int64
	for _, node := range nodeList.Items {
		var tmpPodCount int64
		for _, pod := range podList.Items {
			if pod.Spec.NodeName == node.Name && pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
				tmpPodCount++
			}
		}
		// update cluster pod count and allocation
		totalPodCurrent += tmpPodCount
		totalPodAllocatable += node.Status.Allocatable.Pods().Value()
		// establish resource allocation values
		cm.nodeMetrics[node.Name] = &nodeMetric{
			name: node.Name,
			cpu: &resourceMetric{
				resourceType: CPU,
				allocatable:  node.Status.Allocatable[CPU],
			},
			memory: &resourceMetric{
				resourceType: Memory,
				allocatable:  node.Status.Allocatable[Memory],
			},
			eni: &resourceMetric{
				resourceType: ENI,
				allocatable:  node.Status.Allocatable[ENI],
			},
			podMetrics: map[string]*podMetric{},
			podCount: &podCount{
				current:     tmpPodCount,
				allocatable: node.Status.Allocatable.Pods().Value(),
			},
			nodeLabels: node.GetLabels(),
		}
	}

	cm.podCount.current = totalPodCurrent
	cm.podCount.allocatable = totalPodAllocatable

	// add resource utilization values if specified
	if nmList != nil {
		for _, nm := range nmList.Items {
			cm.nodeMetrics[nm.Name].cpu.utilization = nm.Usage[CPU]
			cm.nodeMetrics[nm.Name].memory.utilization = nm.Usage[Memory]
			cm.nodeMetrics[nm.Name].eni.utilization = nm.Usage[ENI]
		}
	}

	// build pod metrics list if utilization is specified
	podMetrics := map[string]v1beta1.PodMetrics{}
	if pmList != nil {
		for _, pm := range pmList.Items {
			podMetrics[fmt.Sprintf("%s-%s", pm.GetNamespace(), pm.GetName())] = pm
		}
	}

	// Iterate over pods, for active pods update cluster metrics
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			cm.addPodMetric(&pod, podMetrics[fmt.Sprintf("%s-%s", pod.GetNamespace(), pod.GetName())])
		}
	}

	// Iterate over nodes
	for _, node := range nodeList.Items {
		if nm, ok := cm.nodeMetrics[node.Name]; ok {
			// update aggregate cluster metrics from all nodes
			cm.addNodeMetric(nm)
			// When namespace filtering is configured, we want to sum pod
			// utilization instead of relying on node util.
			if nmList == nil {
				nm.addPodUtilization()
			}
		}
	}

	// update applist summary metrics
	for _, app := range appList {
		for _, pod := range app.Items {

			req, limit := resourcehelper.PodRequestsAndLimits(&pod)

			app.cpu.request.Add(req[CPU])
			app.cpu.limit.Add(limit[CPU])

			app.memory.request.Add(req[Memory])
			app.memory.limit.Add(limit[Memory])

			app.eni.request.Add(req[ENI])
			app.eni.limit.Add(limit[ENI])

			if pmList != nil {
				for _, container := range podMetrics[fmt.Sprintf("%s-%s", pod.GetNamespace(), pod.GetName())].Containers {
					app.cpu.utilization.Add(container.Usage[CPU])
					app.memory.utilization.Add(container.Usage[Memory])
					app.eni.utilization.Add(container.Usage[ENI])
				}
			}
		}
	}

	return cm
}

func (rm *resourceMetric) addMetric(m *resourceMetric) {
	rm.allocatable.Add(m.allocatable)
	rm.utilization.Add(m.utilization)
	rm.request.Add(m.request)
	rm.limit.Add(m.limit)
}

func (cm *clusterMetric) addPodMetric(pod *corev1.Pod, podMetrics v1beta1.PodMetrics) {

	req, limit := resourcehelper.PodRequestsAndLimits(pod)
	key := fmt.Sprintf("%s-%s", pod.Namespace, pod.Name)

	// get pointer to node's metrics
	nm := cm.nodeMetrics[pod.Spec.NodeName]

	// build pod's metrics object with limits and requests
	pm := &podMetric{
		name:      pod.Name,
		namespace: pod.Namespace,
		node:      pod.Spec.NodeName,
		cpu: &resourceMetric{
			resourceType: CPU,
			request:      req[CPU],
			limit:        limit[CPU],
		},
		memory: &resourceMetric{
			resourceType: Memory,
			request:      req[Memory],
			limit:        limit[Memory],
		},
		eni: &resourceMetric{
			resourceType: ENI,
			request:      req[ENI],
			limit:        limit[ENI],
		},
		containerMetrics: map[string]*containerMetric{},
	}

	// grab Container limits and requests for each container in the pod
	for _, container := range pod.Spec.Containers {
		pm.containerMetrics[container.Name] = &containerMetric{
			name: container.Name,
			cpu: &resourceMetric{
				resourceType: CPU,
				request:      container.Resources.Requests[CPU],
				limit:        container.Resources.Limits[CPU],
				allocatable:  nm.cpu.allocatable,
			},
			memory: &resourceMetric{
				resourceType: Memory,
				request:      container.Resources.Requests[Memory],
				limit:        container.Resources.Limits[Memory],
				allocatable:  nm.memory.allocatable,
			},
			eni: &resourceMetric{
				resourceType: ENI,
				request:      container.Resources.Requests[ENI],
				limit:        container.Resources.Limits[ENI],
				allocatable:  nm.memory.allocatable,
			},
		}
	}

	// update node's specific pod metrics
	if nm != nil {
		pm.cpu.allocatable = nm.cpu.allocatable
		pm.memory.allocatable = nm.memory.allocatable
		pm.eni.allocatable = nm.eni.allocatable

		nm.podMetrics[key] = pm

		// also add to cluster wide list
		cm.podMetrics[key] = pm

		nm.cpu.request.Add(req[CPU])
		nm.cpu.limit.Add(limit[CPU])
		nm.memory.request.Add(req[Memory])
		nm.memory.limit.Add(limit[Memory])
		nm.eni.request.Add(req[ENI])
		nm.eni.limit.Add(limit[ENI])
	}

	// update pod's utilization data from the container utilization data
	for _, container := range podMetrics.Containers {
		cm := pm.containerMetrics[container.Name]
		if cm != nil {
			pm.containerMetrics[container.Name].cpu.utilization = container.Usage[CPU]
			pm.cpu.utilization.Add(container.Usage[CPU])
			pm.containerMetrics[container.Name].memory.utilization = container.Usage[Memory]
			pm.memory.utilization.Add(container.Usage[Memory])
			pm.containerMetrics[container.Name].eni.utilization = container.Usage[ENI]
			pm.eni.utilization.Add(container.Usage[ENI])
		}
	}
}

func (cm *clusterMetric) addNodeMetric(nm *nodeMetric) {
	cm.cpu.addMetric(nm.cpu)
	cm.memory.addMetric(nm.memory)
	cm.eni.addMetric(nm.eni)
}

func (cm *clusterMetric) getSortedNodeMetrics(groupByLabels []string, sortBy string) []*nodeMetric {

	sortedNodeMetrics := make([]*nodeMetric, len(cm.nodeMetrics))

	i := 0
	for name := range cm.nodeMetrics {
		sortedNodeMetrics[i] = cm.nodeMetrics[name]
		i++
	}

	sort.Slice(sortedNodeMetrics, func(i, j int) bool {
		m1 := sortedNodeMetrics[i]
		m2 := sortedNodeMetrics[j]

		// sort by grouping labels first (if any are specified)
		for _, label := range groupByLabels {
			if m1.nodeLabels[label] != m2.nodeLabels[label] {
				return m1.nodeLabels[label] < m2.nodeLabels[label]
			}
		}
		// if all labels match or if none are defined, sort by specified value

		switch sortBy {
		case "cpu.util":
			return m2.cpu.utilization.MilliValue() < m1.cpu.utilization.MilliValue()
		case "cpu.limit":
			return m2.cpu.limit.MilliValue() < m1.cpu.limit.MilliValue()
		case "cpu.request":
			return m2.cpu.request.MilliValue() < m1.cpu.request.MilliValue()
		case "mem.util":
			return m2.memory.utilization.Value() < m1.memory.utilization.Value()
		case "mem.limit":
			return m2.memory.limit.Value() < m1.memory.limit.Value()
		case "mem.request":
			return m2.memory.request.Value() < m1.memory.request.Value()
		case "cpu.util.percentage":
			return m2.cpu.percent(m2.cpu.utilization) < m1.cpu.percent(m1.cpu.utilization)
		case "cpu.limit.percentage":
			return m2.cpu.percent(m2.cpu.limit) < m1.cpu.percent(m1.cpu.limit)
		case "cpu.request.percentage":
			return m2.cpu.percent(m2.cpu.request) < m1.cpu.percent(m1.cpu.request)
		case "mem.util.percentage":
			return m2.memory.percent(m2.memory.utilization) < m1.memory.percent(m1.memory.utilization)
		case "mem.limit.percentage":
			return m2.memory.percent(m2.memory.limit) < m1.memory.percent(m1.memory.limit)
		case "mem.request.percentage":
			return m2.memory.percent(m2.memory.request) < m1.memory.percent(m1.memory.request)
		default:
			return m1.name < m2.name
		}
	})

	return sortedNodeMetrics
}

func (nm *nodeMetric) getSortedPodMetrics(sortBy string) []*podMetric {
	sortedPodMetrics := make([]*podMetric, len(nm.podMetrics))

	i := 0
	for name := range nm.podMetrics {
		sortedPodMetrics[i] = nm.podMetrics[name]
		i++
	}

	sort.Slice(sortedPodMetrics, func(i, j int) bool {
		m1 := sortedPodMetrics[i]
		m2 := sortedPodMetrics[j]

		switch sortBy {
		case "cpu.util":
			return m2.cpu.utilization.MilliValue() < m1.cpu.utilization.MilliValue()
		case "cpu.limit":
			return m2.cpu.limit.MilliValue() < m1.cpu.limit.MilliValue()
		case "cpu.request":
			return m2.cpu.request.MilliValue() < m1.cpu.request.MilliValue()
		case "mem.util":
			return m2.memory.utilization.Value() < m1.memory.utilization.Value()
		case "mem.limit":
			return m2.memory.limit.Value() < m1.memory.limit.Value()
		case "mem.request":
			return m2.memory.request.Value() < m1.memory.request.Value()
		case "cpu.util.percentage":
			return m2.cpu.percent(m2.cpu.utilization) < m1.cpu.percent(m1.cpu.utilization)
		case "cpu.limit.percentage":
			return m2.cpu.percent(m2.cpu.limit) < m1.cpu.percent(m1.cpu.limit)
		case "cpu.request.percentage":
			return m2.cpu.percent(m2.cpu.request) < m1.cpu.percent(m1.cpu.request)
		case "mem.util.percentage":
			return m2.memory.percent(m2.memory.utilization) < m1.memory.percent(m1.memory.utilization)
		case "mem.limit.percentage":
			return m2.memory.percent(m2.memory.limit) < m1.memory.percent(m1.memory.limit)
		case "mem.request.percentage":
			return m2.memory.percent(m2.memory.request) < m1.memory.percent(m1.memory.request)
		default:
			return m1.name < m2.name
		}
	})

	return sortedPodMetrics
}

func (nm *nodeMetric) addPodUtilization() {
	for _, pm := range nm.podMetrics {
		nm.cpu.utilization.Add(pm.cpu.utilization)
		nm.memory.utilization.Add(pm.memory.utilization)
		nm.eni.utilization.Add(pm.eni.utilization)
	}
}

func (pm *podMetric) getSortedContainerMetrics(sortBy string) []*containerMetric {
	sortedContainerMetrics := make([]*containerMetric, len(pm.containerMetrics))

	i := 0
	for name := range pm.containerMetrics {
		sortedContainerMetrics[i] = pm.containerMetrics[name]
		i++
	}

	sort.Slice(sortedContainerMetrics, func(i, j int) bool {
		m1 := sortedContainerMetrics[i]
		m2 := sortedContainerMetrics[j]

		switch sortBy {
		case "cpu.util":
			return m2.cpu.utilization.MilliValue() < m1.cpu.utilization.MilliValue()
		case "cpu.limit":
			return m2.cpu.limit.MilliValue() < m1.cpu.limit.MilliValue()
		case "cpu.request":
			return m2.cpu.request.MilliValue() < m1.cpu.request.MilliValue()
		case "mem.util":
			return m2.memory.utilization.Value() < m1.memory.utilization.Value()
		case "mem.limit":
			return m2.memory.limit.Value() < m1.memory.limit.Value()
		case "mem.request":
			return m2.memory.request.Value() < m1.memory.request.Value()
		case "cpu.util.percentage":
			return m2.cpu.percent(m2.cpu.utilization) < m1.cpu.percent(m1.cpu.utilization)
		case "cpu.limit.percentage":
			return m2.cpu.percent(m2.cpu.limit) < m1.cpu.percent(m1.cpu.limit)
		case "cpu.request.percentage":
			return m2.cpu.percent(m2.cpu.request) < m1.cpu.percent(m1.cpu.request)
		case "mem.util.percentage":
			return m2.memory.percent(m2.memory.utilization) < m1.memory.percent(m1.memory.utilization)
		case "mem.limit.percentage":
			return m2.memory.percent(m2.memory.limit) < m1.memory.percent(m1.memory.limit)
		case "mem.request.percentage":
			return m2.memory.percent(m2.memory.request) < m1.memory.percent(m1.memory.request)
		default:
			return m1.name < m2.name
		}
	})

	return sortedContainerMetrics
}

func (rm *resourceMetric) requestString(cr *DisplayCriteria) string {
	return rm.resourceString(rm.request, cr)
}

func (rm *resourceMetric) limitString(cr *DisplayCriteria) string {
	return rm.resourceString(rm.limit, cr)
}

func (rm *resourceMetric) utilString(cr *DisplayCriteria) string {
	return rm.resourceString(rm.utilization, cr)
}

// podCountString returns the string representation of podCount struct, example: "15/110 (12%)"
func (pc *podCount) podCountString() string {
	return fmt.Sprintf("%d/%d (%d%%%%)", pc.current, pc.allocatable,
		percentRawFunction(float64(pc.current), float64(pc.allocatable)))
}

func (rm *resourceMetric) resourceString(r resource.Quantity, cr *DisplayCriteria) string {

	// in podsummary do not show % if no allocation is present
	if cr.ShowPodSummary && rm.allocatable.Value() == 0 {
		return fmt.Sprintf("%s", rm.valueFunction()(r))
	}

	if cr.AvailableFormat {
		return fmt.Sprintf("%s/%s", rm.valueAvailableFunction()(r), rm.valueFunction()(rm.allocatable))
	}

	return fmt.Sprintf("%s (%v)", rm.valueFunction()(r), rm.percentFunction()(r))
}

func formatToMegiBytes(actual resource.Quantity) int64 {
	value := actual.Value() / Mebibyte
	// rounding up
	if actual.Value()%Mebibyte != 0 {
		value++
	}
	return value
}

// NOTE: This might not be a great place for closures due to the cyclical nature of how resourceType works. Perhaps better implemented another way.
func (rm resourceMetric) valueFunction() (f func(r resource.Quantity) string) {
	switch rm.resourceType {
	case CPU:
		f = func(r resource.Quantity) string {
			return fmt.Sprintf("%dm", r.MilliValue())
		}
	case Memory:
		f = func(r resource.Quantity) string {
			return fmt.Sprintf("%dMi", formatToMegiBytes(r))
		}
	case ENI:
		f = func(r resource.Quantity) string {
			return stringFormatInt64(r.Value())
		}
	}
	return f
}

func (rm resourceMetric) valueAvailableFunction() (f func(r resource.Quantity) string) {
	switch rm.resourceType {
	case CPU:
		f = func(r resource.Quantity) string {
			return fmt.Sprintf("%dm", rm.allocatable.MilliValue()-r.MilliValue())
		}
	case Memory:
		f = func(r resource.Quantity) string {
			return fmt.Sprintf("%dMi", formatToMegiBytes(rm.allocatable)-formatToMegiBytes(r))
		}
	case ENI:
		f = func(r resource.Quantity) string {
			return stringFormatInt64(rm.allocatable.Value() - r.Value())
		}
	}
	return f
}

func (rm resourceMetric) valueCSVFunction() (f func(r resource.Quantity) string) {
	switch rm.resourceType {
	case CPU:
		f = func(r resource.Quantity) string {
			return stringFormatInt64(r.MilliValue())
		}
	case Memory:
		f = func(r resource.Quantity) string {
			return stringFormatInt64(formatToMegiBytes(r))
		}
	case ENI:
		f = func(r resource.Quantity) string {
			return stringFormatInt64(r.Value())
		}
	}
	return f
}

// NOTE: This might not be a great place for closures due to the cyclical nature of how resourceType works. Perhaps better implemented another way.
func (rm resourceMetric) percentFunction() (f func(r resource.Quantity) string) {
	f = func(r resource.Quantity) string {
		return fmt.Sprintf("%v%%%%", rm.percent(r))
	}
	return f
}

func (rm resourceMetric) percentFunctionWithoutDoubleEscape() (f func(r resource.Quantity) string) {
	f = func(r resource.Quantity) string {
		return fmt.Sprintf("%v%%", rm.percent(r))
	}
	return f
}

func (rm resourceMetric) percent(r resource.Quantity) int64 {
	return percentRawFunction(float64(r.MilliValue()), float64(rm.allocatable.MilliValue()))
}

func percentRawFunction(nominator, denominator float64) int64 {
	if denominator > 0.0 {
		return int64(100.0 * nominator / denominator)
	}
	return 0
}

// For CSV / TSV formatting Helper Functions
// -----------------------------------------

func (rm *resourceMetric) capacityString() string {
	return rm.valueCSVFunction()(rm.allocatable)
}

func (rm *resourceMetric) requestActualString() string {
	return rm.valueCSVFunction()(rm.request)
}

func (rm *resourceMetric) requestPercentageString() string {
	return rm.percentFunction()(rm.request)
}

func (rm *resourceMetric) limitActualString() string {
	return rm.valueCSVFunction()(rm.limit)
}

func (rm *resourceMetric) limitPercentageString() string {
	return rm.percentFunction()(rm.limit)
}

func (rm *resourceMetric) utilActualString() string {
	return rm.valueCSVFunction()(rm.utilization)
}

func (rm *resourceMetric) utilPercentageString() string {
	return rm.percentFunction()(rm.utilization)
}

func (pc *podCount) podCountCurrentString() string {
	return stringFormatInt64(pc.current)
}

func (pc *podCount) podCountAllocatableString() string {
	return stringFormatInt64(pc.allocatable)
}

func (pc *podCount) podCountPercentageString() string {
	return stringFormatInt64(percentRawFunction(float64(pc.current), float64(pc.allocatable)))
}

func stringFormatInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}

/*
Scans all of the pods' labels and returns a list of the unique values set for 'appname'
*/
func getPodAppsList(podList *corev1.PodList, selectPodLabels string) (result []podAppSummary, orderedPodAppKeys []string) {

	// establish 'no label set' entry for pods that do not have an 'appname' label set
	noLabelCount := int64(0)
	noLabelPods := []corev1.Pod{}

	if selectPodLabels != "" {
		orderedPodAppKeys = strings.Split(selectPodLabels, ",")
	} else {
		orderedPodAppKeys = strings.Split(PodAppNameLabelDefaultSelector, ",")
	}

	for _, pod := range podList.Items {
		foundLabel := false
		for k, v := range pod.GetLabels() {
			for _, checkLabel := range orderedPodAppKeys {
				if k == checkLabel {
					foundLabel = true
					found := false
					for i, podapp := range result {
						if v == podapp.appNameLabel && k == podapp.appNameKey {
							found = true
							result[i].podCount++
							result[i].Items = append(result[i].Items, pod)
							break
						}
					}
					// add the label if it did not exist before
					if !found {
						result = append(result, podAppSummary{
							appNameKey:        k,
							appNameLabel:      v,
							podCount:          1,
							specialNoLabelSet: false,
							Items:             []corev1.Pod{pod},
							cpu:               &resourceMetric{resourceType: CPU},
							memory:            &resourceMetric{resourceType: Memory},
							eni:               &resourceMetric{resourceType: ENI},
						})
					}
					break
				}
			}
		}
		if !foundLabel {
			noLabelCount++
			noLabelPods = append(noLabelPods, pod)
		}
	}
	// add 'no label set' entry if they exist
	if noLabelCount > 0 {
		result = append(result, podAppSummary{
			appNameKey:        "",
			appNameLabel:      "",
			podCount:          noLabelCount,
			specialNoLabelSet: true,
			Items:             noLabelPods,
			cpu:               &resourceMetric{resourceType: CPU},
			memory:            &resourceMetric{resourceType: Memory},
			eni:               &resourceMetric{resourceType: ENI},
		})
	}

	return result, orderedPodAppKeys
}
