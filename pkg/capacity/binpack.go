package capacity

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type binAnalysis struct {
	nodesWellUtilized       string `json:"nodesWellUtilized,omitempty"`
	nodesUnderutilized      string `json:"nodesUnderutilized,omitempty"`
	nodesUnbalanced         string `json:"nodesUnbalanced,omitempty"`
	idleHeadroom            string `json:"idleHeadroomPercentage,omitempty"`
	idleWasteCPU            string `json:"idleCpuWastePercentage,omitempty"`
	idleWasteMEM            string `json:"idlememoryWastePercentage,omitempty"`
	idleWastePODS           string `json:"idlepodcountWastePercentage,omitempty"`
	binpackRequestRatio     string `json:"requestMemoryToCPUCoreRation,omitempty"`
	binpackLimitRatio       string `json:"limitMemoryToCPUCoreRation,omitempty"`
	binpackUtilizationRatio string `json:"utilizationMemoryToCPUCoreRation,omitempty"`
}

var binHeaders = binAnalysis{
	nodesWellUtilized:       "WELL UTILZD NODES",
	nodesUnderutilized:      "UNDER UTILZD NODES",
	nodesUnbalanced:         "UNBALNCD NODES",
	idleHeadroom:            "IDLE HEADROOM %%",
	idleWasteCPU:            "IDLE WASTE CPU %%",
	idleWasteMEM:            "IDLE WASTE MEM %%",
	idleWastePODS:           "IDLE WASTE PODS %%",
	binpackRequestRatio:     "REQUEST MEM/CPU",
	binpackLimitRatio:       "LIMITS MEM/CPU",
	binpackUtilizationRatio: "UTIL MEM/CPU",
}

type analysisData struct {
	headroom struct {
		overall int64
		cpu     int64
		memory  int64
		pods    int64
	}
	memToCpuRatios struct {
		limit       int64
		request     int64
		utilization int64
	}
	waste struct {
		cpu    int64
		memory int64
		pods   int64
	}
}

func getAnalysisData(cpu *resourceMetric, memory *resourceMetric, eni *resourceMetric, podCount *podCount) (result *analysisData) {

	result = &analysisData{}

	result.memToCpuRatios.limit = memoryToCPUCoreRatio(memory.limit, cpu.limit)
	result.memToCpuRatios.request = memoryToCPUCoreRatio(memory.request, cpu.request)
	result.memToCpuRatios.utilization = memoryToCPUCoreRatio(memory.utilization, cpu.utilization)

	cpuReqPercent := cpu.percent(cpu.request)
	memReqPercent := memory.percent(memory.request)
	// eniReqPercent := eni.percent(eni.request)
	podCntPercent := percentRawFunction(float64(podCount.current), float64(podCount.allocatable))

	max := maxOfThree(cpuReqPercent, memReqPercent, podCntPercent)

	result.headroom.overall = headroom(max)
	result.headroom.cpu = headroom(cpuReqPercent)
	result.headroom.memory = headroom(memReqPercent)
	result.headroom.pods = headroom(podCntPercent)

	result.waste.cpu = result.headroom.cpu - result.headroom.overall
	result.waste.memory = result.headroom.memory - result.headroom.overall
	result.waste.pods = result.headroom.pods - result.headroom.overall

	return result
}

func (cm *clusterMetric) getBinAnalysis() binAnalysis {

	x := getAnalysisData(cm.cpu, cm.memory, cm.eni, cm.podCount)

	var results = binAnalysis{
		nodesWellUtilized:       fmt.Sprintf("%d", cm.analysis.nodesWellUtilized),
		nodesUnderutilized:      fmt.Sprintf("%d", cm.analysis.nodesUnderutilized),
		nodesUnbalanced:         fmt.Sprintf("%d", cm.analysis.nodesUnbalanced),
		idleHeadroom:            fmt.Sprintf("%d%%%%", x.headroom.overall),
		idleWasteCPU:            fmt.Sprintf("%d%%%%", x.waste.cpu),
		idleWasteMEM:            fmt.Sprintf("%d%%%%", x.waste.memory),
		idleWastePODS:           fmt.Sprintf("%d%%%%", x.waste.pods),
		binpackRequestRatio:     fmt.Sprintf("%d", x.memToCpuRatios.request),
		binpackLimitRatio:       fmt.Sprintf("%d", x.memToCpuRatios.limit),
		binpackUtilizationRatio: fmt.Sprintf("%d", x.memToCpuRatios.utilization),
	}

	return results
}

func (pal *podAppSummary) getBinAnalysis() binAnalysis {

	x := getAnalysisData(pal.cpu, pal.memory, pal.eni, &podCount{})

	var results = binAnalysis{
		nodesWellUtilized:       "n/a",
		nodesUnderutilized:      "n/a",
		nodesUnbalanced:         "n/a",
		idleHeadroom:            "n/a",
		idleWasteCPU:            "n/a",
		idleWasteMEM:            "n/a",
		idleWastePODS:           "n/a",
		binpackRequestRatio:     fmt.Sprintf("%d", x.memToCpuRatios.request),
		binpackLimitRatio:       fmt.Sprintf("%d", x.memToCpuRatios.limit),
		binpackUtilizationRatio: fmt.Sprintf("%d", x.memToCpuRatios.utilization),
	}

	return results
}

func (nm *nodeMetric) getBinAnalysis() binAnalysis {

	x := getAnalysisData(nm.cpu, nm.memory, nm.eni, nm.podCount)

	var results = binAnalysis{
		nodesWellUtilized:       fmt.Sprintf("%d", nm.analysis.nodesWellUtilized),
		nodesUnderutilized:      fmt.Sprintf("%d", nm.analysis.nodesUnderutilized),
		nodesUnbalanced:         fmt.Sprintf("%d", nm.analysis.nodesUnbalanced),
		idleHeadroom:            fmt.Sprintf("%d%%%%", x.headroom.overall),
		idleWasteCPU:            fmt.Sprintf("%d%%%%", x.waste.cpu),
		idleWasteMEM:            fmt.Sprintf("%d%%%%", x.waste.memory),
		idleWastePODS:           fmt.Sprintf("%d%%%%", x.waste.pods),
		binpackRequestRatio:     fmt.Sprintf("%d", x.memToCpuRatios.request),
		binpackLimitRatio:       fmt.Sprintf("%d", x.memToCpuRatios.limit),
		binpackUtilizationRatio: fmt.Sprintf("%d", x.memToCpuRatios.utilization),
	}

	return results
}

func (pm *podMetric) getBinAnalysis() binAnalysis {

	x := getAnalysisData(pm.cpu, pm.memory, pm.eni, &podCount{})

	var results = binAnalysis{
		nodesWellUtilized:       "",
		nodesUnderutilized:      "",
		nodesUnbalanced:         "",
		idleHeadroom:            "",
		idleWasteCPU:            "",
		idleWasteMEM:            "",
		idleWastePODS:           "",
		binpackRequestRatio:     fmt.Sprintf("%d", x.memToCpuRatios.request),
		binpackLimitRatio:       fmt.Sprintf("%d", x.memToCpuRatios.limit),
		binpackUtilizationRatio: fmt.Sprintf("%d", x.memToCpuRatios.utilization),
	}

	return results
}

func (cm *containerMetric) getBinAnalysis() binAnalysis {

	var results = binAnalysis{
		nodesWellUtilized:       "",
		nodesUnderutilized:      "",
		nodesUnbalanced:         "",
		idleHeadroom:            "",
		idleWasteCPU:            "",
		idleWasteMEM:            "",
		idleWastePODS:           "",
		binpackRequestRatio:     "",
		binpackLimitRatio:       "",
		binpackUtilizationRatio: "",
	}

	return results
}

// binpack analysis

func maxOfThree(a, b, c int64) int64 {

	if a > b {
		if a > c {
			return a
		}
	} else {
		if b > c {
			return b
		}
	}

	return c
}

func headroom(a int64) int64 {
	if a > 100 {
		return 0
	}
	if a < 0 {
		return 100
	}
	return 100 - a
}

func headroomDelta(a, max int64) int64 {
	delta := max - a
	if delta < 0 {
		return 0
	}
	if delta > 100 {
		return 100
	}
	return delta
}

func memoryToCPUCoreRatio(memory, cpu resource.Quantity) int64 {
	if cpu.MilliValue() == 0 {
		return 0
	}
	return formatToMegiBytes(memory) / cpu.Value()
}
