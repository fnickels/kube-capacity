package capacity

import (
	"fmt"
	"strings"
)

func (cm *clusterMetric) getUniquePodLabels() (result []string, resultMap map[string]bool) {

	for _, pod := range cm.rawPodList {
		for k, _ := range pod.Labels {
			if !resultMap[k] {
				result = append(result, k)
				resultMap[k] = true
			}
		}
	}

	return result, resultMap
}

func processPodLabelToDisplay(cm *clusterMetric, cr *DisplayCriteria) ([]string, error) {

	badLabels := []string{}
	labelsToDisplay := []string{}

	uniquePodLabels, podLabels := cm.getUniquePodLabels()

	if cr.ShowAllPodLabels {
		// show all pod labels, copy unique list
		labelsToDisplay = uniquePodLabels
	} else if cr.DisplayPodLabels != "" {
		// select which labels to display
		for _, label := range strings.Split(cr.DisplayPodLabels, ",") {
			if podLabels[label] {
				labelsToDisplay = append(labelsToDisplay, label)
			} else {
				badLabels = append(badLabels, label)
			}
		}
	}

	if len(badLabels) > 0 {
		return []string{}, fmt.Errorf("unknown (display) pod label(s): %v", badLabels)
	}

	return labelsToDisplay, nil
}
