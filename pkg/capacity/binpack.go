package capacity

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type binAnalysis struct {
	idleHeadroom            string `json:"idleHeadroomPercentage,omitempty"`
	idleWasteCPU            string `json:"idleCpuWastePercentage,omitempty"`
	idleWasteMEM            string `json:"idlememoryWastePercentage,omitempty"`
	idleWastePODS           string `json:"idlepodcountWastePercentage,omitempty"`
	binpackRequestRatio     string `json:"requestMemoryToCPUCoreRation,omitempty"`
	binpackLimitRatio       string `json:"limitMemoryToCPUCoreRation,omitempty"`
	binpackUtilizationRatio string `json:"utilizationMemoryToCPUCoreRation,omitempty"`
}

var binHeaders = binAnalysis{
	idleHeadroom:            "IDLE HEADROOM %%",
	idleWasteCPU:            "IDLE WASTE CPU %%",
	idleWasteMEM:            "IDLE WASTE MEM %%",
	idleWastePODS:           "IDLE WASTE PODS %%",
	binpackRequestRatio:     "CPU:MEM REQUESTS",
	binpackLimitRatio:       "CPU:MEM LIMITS",
	binpackUtilizationRatio: "CPU:MEM UTIL",
}

func (cm *clusterMetric) getBinAnalysis() binAnalysis {

	cpuReqPercent := cm.cpu.percent(cm.cpu.request)
	memReqPercent := cm.memory.percent(cm.memory.request)
	podCntPercent := percentRawFunction(float64(cm.podCount.current), float64(cm.podCount.allocatable))

	max := maxOfThree(cpuReqPercent, memReqPercent, podCntPercent)
	headroom := headroom(max)

	wasteCPU := headroomDelta(cpuReqPercent, max)
	wasteMEM := headroomDelta(memReqPercent, max)
	wastePODS := headroomDelta(podCntPercent, max)

	limitRatio := memoryToCPUCoreRatio(cm.memory.limit, cm.cpu.limit)
	requestsRatio := memoryToCPUCoreRatio(cm.memory.request, cm.cpu.request)
	utilizationRatio := memoryToCPUCoreRatio(cm.memory.utilization, cm.cpu.utilization)

	var results = binAnalysis{
		idleHeadroom:            fmt.Sprintf("%d%%%%", headroom),
		idleWasteCPU:            fmt.Sprintf("%d%%%%", wasteCPU),
		idleWasteMEM:            fmt.Sprintf("%d%%%%", wasteMEM),
		idleWastePODS:           fmt.Sprintf("%d%%%%", wastePODS),
		binpackRequestRatio:     fmt.Sprintf("%d", requestsRatio),
		binpackLimitRatio:       fmt.Sprintf("%d", limitRatio),
		binpackUtilizationRatio: fmt.Sprintf("%d", utilizationRatio),
	}

	return results
}

func (nm *nodeMetric) getBinAnalysis() binAnalysis {

	cpuReqPercent := nm.cpu.percent(nm.cpu.request)
	memReqPercent := nm.memory.percent(nm.memory.request)
	podCntPercent := percentRawFunction(float64(nm.podCount.current), float64(nm.podCount.allocatable))

	max := maxOfThree(cpuReqPercent, memReqPercent, podCntPercent)
	headroom := headroom(max)

	wasteCPU := headroomDelta(cpuReqPercent, max)
	wasteMEM := headroomDelta(memReqPercent, max)
	wastePODS := headroomDelta(podCntPercent, max)

	limitRatio := memoryToCPUCoreRatio(nm.memory.limit, nm.cpu.limit)
	requestsRatio := memoryToCPUCoreRatio(nm.memory.request, nm.cpu.request)
	utilizationRatio := memoryToCPUCoreRatio(nm.memory.utilization, nm.cpu.utilization)

	var results = binAnalysis{
		idleHeadroom:            fmt.Sprintf("%d%%%%", headroom),
		idleWasteCPU:            fmt.Sprintf("%d%%%%", wasteCPU),
		idleWasteMEM:            fmt.Sprintf("%d%%%%", wasteMEM),
		idleWastePODS:           fmt.Sprintf("%d%%%%", wastePODS),
		binpackRequestRatio:     fmt.Sprintf("%d", requestsRatio),
		binpackLimitRatio:       fmt.Sprintf("%d", limitRatio),
		binpackUtilizationRatio: fmt.Sprintf("%d", utilizationRatio),
	}

	return results
}

func (pm *podMetric) getBinAnalysis() binAnalysis {

	limitRatio := memoryToCPUCoreRatio(pm.memory.limit, pm.cpu.limit)
	requestsRatio := memoryToCPUCoreRatio(pm.memory.request, pm.cpu.request)
	utilizationRatio := memoryToCPUCoreRatio(pm.memory.utilization, pm.cpu.utilization)

	var results = binAnalysis{
		idleHeadroom:            "",
		idleWasteCPU:            "",
		idleWasteMEM:            "",
		idleWastePODS:           "",
		binpackRequestRatio:     fmt.Sprintf("%d", requestsRatio),
		binpackLimitRatio:       fmt.Sprintf("%d", limitRatio),
		binpackUtilizationRatio: fmt.Sprintf("%d", utilizationRatio),
	}

	return results
}

func (cm *containerMetric) getBinAnalysis() binAnalysis {

	var results = binAnalysis{
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
