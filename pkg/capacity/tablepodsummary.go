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
	cm                         *clusterMetric
	cr                         *DisplayCriteria
	w                          *tabwriter.Writer
	uniquePodAppSelectorLabels []string
	uniquePodLabels            []string
	uniqueGroupByNodeLabels    []string
	uniqueDisplayNodeLabels    []string
	uniqueRemainderNodeLabels  []string
}

type tablePodLine struct {
	appNameLabel    string
	namespace       string
	pod             string
	podStatus       string
	node            string
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
	podStatus:       "POD STATUS",
	node:            "NODE",
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

func PrintTablePodSummary(cm *clusterMetric, cr *DisplayCriteria) {

	pp := &tablePodPrinter{
		cm: cm,
		cr: cr,
		w:  new(tabwriter.Writer),
	}

	pp.w.Init(os.Stdout, 0, 8, 2, ' ', 0)

	var err error

	// process Node Label selection elements
	pp.uniquePodLabels, err = processPodLabelToDisplay(pp.cm)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// copy Pod Label names to the Header object
	tablePodHeaderStrings.podLabels = pp.uniquePodLabels

	// sort pod list (maybe)
	sortedPodAppList := pp.cm.rawPodAppList

	pp.printLine(&tablePodHeaderStrings)

	if len(sortedPodAppList) > 1 {
		pp.printClusterLine()
	}

	for _, pal := range sortedPodAppList {

		pp.printPodAppLine(&pal)

		if cr.ShowPods || cr.ShowContainers {
			for _, pl := range pal.Items {

				key := fmt.Sprintf("%s-%s", pl.Namespace, pl.Name)
				pm := cm.podMetrics[key]

				pp.printPodLine(&pl, pm, &pal)

				if cr.ShowContainers {
					for _, cc := range pl.Spec.InitContainers {
						pp.printContainerLine(&pl, &pal, &cc, InitContainerClassification)
					}

					for _, cc := range pl.Spec.Containers {
						pp.printContainerLine(&pl, &pal, &cc, NormalContainerClassification)
					}

					for _, cc := range pl.Spec.EphemeralContainers {
						ce := corev1.Container(cc.EphemeralContainerCommon)
						pp.printContainerLine(&pl, &pal, &ce, EphemeralContainerClassification)
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
	if pp.cr.ShowDebug {
		fmt.Fprintf(os.Stdout, "LineItems: %v\n", lineItems)
	}
	fmt.Fprintf(pp.w, strings.Join(lineItems[:], "\t ")+"\n")
}

func (pp *tablePodPrinter) getLineItems(tl *tablePodLine) []string {

	lineItems := []string{tl.appNameLabel}

	if pp.cr.ShowNamespace() {
		lineItems = append(lineItems, tl.namespace)
	}

	if pp.cr.ShowContainers || pp.cr.ShowPods {
		lineItems = append(lineItems, tl.pod)
		lineItems = append(lineItems, tl.podStatus)
		lineItems = append(lineItems, tl.node)
	}

	if pp.cr.ShowContainers {
		lineItems = append(lineItems, tl.container)
		lineItems = append(lineItems, string(tl.containerType))
	}

	if pp.cr.ShowPodCount {
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

	lineItems = append(lineItems, tl.cpuRequests)
	lineItems = append(lineItems, tl.cpuLimits)

	if pp.cr.ShowUtil {
		lineItems = append(lineItems, tl.cpuUtil)
	}

	lineItems = append(lineItems, tl.memoryRequests)
	lineItems = append(lineItems, tl.memoryLimits)

	if pp.cr.ShowUtil {
		lineItems = append(lineItems, tl.memoryUtil)
	}

	// lineItems = append(lineItems, tl.eniRequests)
	// lineItems = append(lineItems, tl.eniLimits)
	//
	// if pp.cr.ShowUtil {
	// 	lineItems = append(lineItems, tl.eniUtil)
	// }

	if pp.cr.BinpackAnalysis {
		//	lineItems = append(lineItems, tl.binpack.idleHeadroom)
		//	lineItems = append(lineItems, tl.binpack.idleWasteCPU)
		//	lineItems = append(lineItems, tl.binpack.idleWasteMEM)
		//	lineItems = append(lineItems, tl.binpack.idleWastePODS)
		lineItems = append(lineItems, tl.binpack.binpackRequestRatio)
		lineItems = append(lineItems, tl.binpack.binpackLimitRatio)
		lineItems = append(lineItems, tl.binpack.binpackUtilizationRatio)
	}

	// if Pod Labels have been specified to be displayed add them here
	for _, x := range tl.podLabels {
		lineItems = append(lineItems, x)
	}

	return lineItems
}

func (pp *tablePodPrinter) printClusterLine() {

	pcSum := int64(0)

	for _, ps := range pp.cm.rawPodAppList {
		pcSum += ps.podCount
	}

	pp.printLine(&tablePodLine{
		appNameLabel:   VoidValue,
		namespace:      VoidValue,
		pod:            VoidValue,
		podStatus:      VoidValue,
		node:           VoidValue,
		container:      VoidValue,
		containerType:  VoidContainerClassification,
		podCount:       stringFormatInt64(pcSum),
		podLabels:      sliceFilledWithString(len(pp.uniquePodLabels), VoidValue),
		cpuRequests:    pp.cm.cpu.requestString(pp.cr),
		cpuLimits:      pp.cm.cpu.limitString(pp.cr),
		cpuUtil:        pp.cm.cpu.utilString(pp.cr),
		memoryRequests: pp.cm.memory.requestString(pp.cr),
		memoryLimits:   pp.cm.memory.limitString(pp.cr),
		memoryUtil:     pp.cm.memory.utilString(pp.cr),
		eniRequests:    pp.cm.eni.requestString(pp.cr),
		eniLimits:      pp.cm.eni.limitString(pp.cr),
		eniUtil:        pp.cm.eni.utilString(pp.cr),
		//		groupByLabels:   setMultipleVoids(len(pp.uniqueGroupByNodeLabels)),
		//		displayLabels:   setMultipleVoids(len(pp.uniqueDisplayNodeLabels)),
		//		remainderLabels: setMultipleVoids(len(pp.uniqueRemainderNodeLabels)),
		binpack: pp.cm.getBinAnalysis(),
	})
}

func (pp *tablePodPrinter) printPodAppLine(pal *podAppSummary) {

	// get pod labels from across all related pods
	labelList := make([]string, len(pp.uniquePodLabels))
	for i, labelName := range pp.uniquePodLabels {
		listOfValues := map[string]bool{}
		for _, pod := range pal.Items {
			labelValue := pod.Labels[labelName]
			if _, ok := listOfValues[labelValue]; ok {
				listOfValues[labelValue] = true
			}
		}

		values := make([]string, len(listOfValues))
		j := 0
		for k := range listOfValues {
			values[j] = k
			j++
		}
		labelList[i] = strings.Join(values, ",")
	}

	pp.printLine(&tablePodLine{
		appNameLabel:   pal.setAppLabel(),
		namespace:      pal.getNamespacesUsed(),
		pod:            VoidValue,
		podStatus:      VoidValue,
		node:           VoidValue,
		container:      VoidValue,
		containerType:  VoidContainerClassification,
		podCount:       stringFormatInt64(pal.podCount),
		podLabels:      labelList,
		cpuRequests:    pal.cpu.requestString(pp.cr),
		cpuLimits:      pal.cpu.limitString(pp.cr),
		cpuUtil:        pal.cpu.utilString(pp.cr),
		memoryRequests: pal.memory.requestString(pp.cr),
		memoryLimits:   pal.memory.limitString(pp.cr),
		memoryUtil:     pal.memory.utilString(pp.cr),
		eniRequests:    pal.eni.requestString(pp.cr),
		eniLimits:      pal.eni.limitString(pp.cr),
		eniUtil:        pal.eni.utilString(pp.cr),
		//		groupByLabels:   setNodeLabels(pp.uniqueGroupByNodeLabels, nm),
		//		displayLabels:   setNodeLabels(pp.uniqueDisplayNodeLabels, nm),
		//		remainderLabels: setNodeLabels(pp.uniqueRemainderNodeLabels, nm),
		binpack: pal.getBinAnalysis(),
	})

}

func (pp *tablePodPrinter) printPodLine(pl *corev1.Pod, pm *podMetric, pal *podAppSummary) {

	req, limit := resourcehelper.PodRequestsAndLimits(pl)

	if pp.cr.ShowDebug && pm == nil {
		fmt.Fprintf(os.Stdout, "pod : %v-%v\n", pl.GetNamespace(), pl.GetName())
		fmt.Fprintf(os.Stdout, "pod : %v\n", pl)
		fmt.Fprintf(os.Stdout, "pod : %v\n", pl.Status)
		fmt.Fprintf(os.Stdout, "pod : %v\n", pl.Status.Phase)
		fmt.Fprintf(os.Stdout, "pod metrics: %v\n", pm)
		if pm != nil {
			fmt.Fprintf(os.Stdout, "pod metrics Node: %v\n", pm.node)
		}
		fmt.Fprintf(os.Stdout, "Request : %v\n", req)
		fmt.Fprintf(os.Stdout, "Limit   : %v\n", limit)
		if pl.Spec.Overhead != nil {
			fmt.Fprintf(os.Stdout, "Overhead: %v\n", pl.Spec.Overhead)
		}
		fmt.Fprintf(os.Stdout, "LABELS --> %v\n", pl.Labels)
	}

	labelList := make([]string, len(pp.uniquePodLabels))
	for i, labelName := range pp.uniquePodLabels {
		labelList[i] = pl.Labels[labelName]
	}

	a1 := ""
	a2 := ""
	a3 := ""
	a4 := ""
	a5 := ""
	a6 := ""
	nodename := ""
	z := binAnalysis{}

	if pm != nil {
		a1 = pm.cpu.requestString(pp.cr)
		a2 = pm.cpu.limitString(pp.cr)
		a3 = pm.cpu.utilString(pp.cr)
		a4 = pm.memory.requestString(pp.cr)
		a5 = pm.memory.limitString(pp.cr)
		a6 = pm.memory.utilString(pp.cr)
		nodename = pm.node
		z = pm.getBinAnalysis()
	}

	pp.printLine(&tablePodLine{
		appNameLabel:   pal.setAppLabel(),
		namespace:      pl.GetNamespace(),
		pod:            pl.GetName(),
		podStatus:      string(pl.Status.Phase),
		node:           nodename,
		container:      VoidValue,
		containerType:  VoidContainerClassification,
		podCount:       stringFormatInt64(1),
		podLabels:      labelList,
		cpuRequests:    a1,
		cpuLimits:      a2,
		cpuUtil:        a3,
		memoryRequests: a4,
		memoryLimits:   a5,
		memoryUtil:     a6,
		//		groupByLabels:   setNodeLabels(pp.uniqueGroupByNodeLabels, nm),
		//		displayLabels:   setNodeLabels(pp.uniqueDisplayNodeLabels, nm),
		//		remainderLabels: setNodeLabels(pp.uniqueRemainderNodeLabels, nm),
		binpack: z,
	})

}

func (pp *tablePodPrinter) printContainerLine(pl *corev1.Pod, pal *podAppSummary, cl *corev1.Container, containerType ContainerClassificationType) {

	pp.printLine(&tablePodLine{
		appNameLabel:  "-",
		namespace:     pl.GetNamespace(),
		pod:           pl.GetName(),
		container:     cl.Name,
		containerType: containerType,
		podLabels:     sliceFilledWithString(len(pp.uniquePodLabels), ""),

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

func (pal *podAppSummary) setAppLabel() string {
	if pal.specialNoLabelSet {
		return "< Not Set >"
	}
	return pal.appNameKey + ":" + pal.appNameLabel
}

func (pal *podAppSummary) getNamespacesUsed() string {
	namespaces := map[string]bool{}
	for _, pod := range pal.Items {
		if _, ok := namespaces[pod.Namespace]; !ok {
			namespaces[pod.Namespace] = true
		}
	}
	keys := make([]string, 0, len(namespaces))
	for k := range namespaces {
		keys = append(keys, k)
	}
	return strings.Join(keys, ",")
}
