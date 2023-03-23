package capacity

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

type ContainerClassificationType string

const (
	NormalContainerClassification    ContainerClassificationType = "normal"
	InitContainerClassification      ContainerClassificationType = "init"
	EphemeralContainerClassification ContainerClassificationType = "ephemeral"
	VoidContainerClassification      ContainerClassificationType = "*"
)

type tablePodPrinter struct {
	cm                        *clusterMetric
	showPods                  bool
	showUtil                  bool
	showPodCount              bool
	showContainers            bool
	showNamespace             bool
	showAllNodeLabels         bool
	showDebug                 bool
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
	appNameLabel    string
	namespace       string
	pod             string
	container       string
	containerType   ContainerClassificationType
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
	podLabels       []string
	groupByLabels   []string
	displayLabels   []string
	remainderLabels []string
	binpack         binAnalysis
}

var tablePodHeaderStrings = tablePodLine{
	appNameLabel:    "APP LABEL",
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
	eniRequests:     "ENI REQUESTS",
	eniLimits:       "ENI LIMITS",
	eniUtil:         "ENI UTIL",
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

	// process Node Label selection elements
	pp.uniquePodLabels, err = processPodLabelSelections(pp.cm)

	// sort pod list (maybe)
	sortedPodAppList := pp.cm.rawPodAppList

	pp.printLine(&tablePodHeaderStrings)

	if len(sortedPodAppList) > 1 {
		pp.printClusterLine()
	}

	for _, pal := range sortedPodAppList {

		pp.printPodAppLine(pal)

		if pp.showPods || pp.showContainers {
			for _, pl := range pal.Items {

				pp.printPodLine(pl)

				if pp.showContainers {
					for _, cc := range pl.Spec.InitContainers {
						pp.printContainerLine(pl, cc, InitContainerClassification)
					}

					for _, cc := range pl.Spec.Containers {
						pp.printContainerLine(pl, cc, NormalContainerClassification)
					}

					for _, cc := range pl.Spec.EphemeralContainers {
						pp.printContainerLine(pl, corev1.Container(cc.EphemeralContainerCommon), EphemeralContainerClassification)
					}
				}
			}
		}
	}

	err = pp.w.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to table: %s", err)
		os.Exit(1)
	}
}

func (pp *tablePodPrinter) printLine(tl *tablePodLine) {
	lineItems := pp.getLineItems(tl)
	if pp.showDebug {
		fmt.Fprintf(os.Stdout, "LineItems: %v\n", lineItems)
	}
	fmt.Fprintf(pp.w, strings.Join(lineItems[:], "\t ")+"\n")
}

func (pp *tablePodPrinter) getLineItems(tl *tablePodLine) []string {

	lineItems := []string{tl.appNameLabel}

	if pp.showContainers || pp.showPods {
		if pp.showNamespace {
			lineItems = append(lineItems, tl.namespace)
		}
		lineItems = append(lineItems, tl.pod)
	}

	if pp.showContainers {
		lineItems = append(lineItems, tl.container)
		lineItems = append(lineItems, string(tl.containerType))
	}

	if pp.showPodCount {
		lineItems = append(lineItems, tl.podCount)
	}

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
		appNameLabel:  VoidValue,
		namespace:     VoidValue,
		pod:           VoidValue,
		container:     VoidValue,
		containerType: VoidContainerClassification,
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

func (pp *tablePodPrinter) printPodAppLine(pal podAppSummary) {

	label := pal.appNameLabel
	if pal.specialNoLabelSet {
		label = "< Not Set >"
	}
	pp.printLine(&tablePodLine{
		appNameLabel:  label,
		namespace:     VoidValue,
		pod:           VoidValue,
		container:     VoidValue,
		containerType: VoidContainerClassification,
		podCount:      stringFormatInt64(pal.podCount),
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

func (pp *tablePodPrinter) printPodLine(pl corev1.Pod) {

	req, limit := resourcehelper.PodRequestsAndLimits(&pl)

	if pp.showDebug {
		fmt.Fprintf(os.Stdout, "Request : %v\n", req)
		fmt.Fprintf(os.Stdout, "Limit   : %v\n", limit)
		if pl.Spec.Overhead != nil {
			fmt.Fprintf(os.Stdout, "Overhead: %v\n", pl.Spec.Overhead)
		}
		fmt.Fprintf(os.Stdout, "LABELS --> %v\n", pl.Labels)
	}

	label := pl.Labels[PodAppNameLabel]

	pp.printLine(&tablePodLine{
		appNameLabel:  label,
		namespace:     pl.GetNamespace(),
		pod:           pl.GetName(),
		container:     VoidValue,
		containerType: VoidContainerClassification,
		podCount:      stringFormatInt64(1),
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

func (pp *tablePodPrinter) printContainerLine(pl corev1.Pod, cl corev1.Container, containerType ContainerClassificationType) {

	label := "-"

	pp.printLine(&tablePodLine{
		appNameLabel:  label,
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
