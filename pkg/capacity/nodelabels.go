package capacity

import (
	"fmt"
	"strings"
)

func (cm *clusterMetric) getUniqueNodeLabels() (result []string, resultMap map[string]bool) {

	resultMap = map[string]bool{}

	for _, node := range cm.nodeMetrics {
		for k, _ := range node.nodeLabels {
			if !resultMap[k] {
				result = append(result, k)
				resultMap[k] = true
			}
		}
	}

	return result, resultMap
}

func processNodeLabelSelections(cm *clusterMetric, groupBy, display string, showAll bool) ([]string, []string, []string, error) {

	// if nothing is called for exit
	if groupBy == "" && display == "" && !showAll {
		// only populate 'uniqueNodeLabels' if one of the the selection criteria is set
		return []string{}, []string{}, []string{}, nil
	}

	badLabels := []string{}
	groupByLabels := []string{}
	displayLabels := []string{}
	remainderLabels := []string{}

	groupMap := map[string]bool{}
	displayMap := map[string]bool{}

	uniqueNodeLabels, nodeLabels := cm.getUniqueNodeLabels()

	// process any 'group by' labels 1st
	if groupBy != "" {
		for _, label := range strings.Split(groupBy, ",") {
			if nodeLabels[label] {
				groupByLabels = append(groupByLabels, label)
				groupMap[label] = true
			} else {
				badLabels = append(badLabels, label)
			}
		}
		if len(badLabels) > 0 {
			return []string{}, []string{}, []string{}, fmt.Errorf("unknown (group by) node label(s): %v", badLabels)
		}
	}

	// process any 'display' node labels next
	if display != "" {
		for _, label := range strings.Split(display, ",") {
			if nodeLabels[label] {
				if !groupMap[label] {
					displayLabels = append(displayLabels, label)
					displayMap[label] = true
				}
			} else {
				badLabels = append(badLabels, label)
			}
		}
		if len(badLabels) > 0 {
			return []string{}, []string{}, []string{}, fmt.Errorf("unknown (display) node label(s): %v", badLabels)
		}
	}

	// pick up all node labels not previously selected if 'showAll' is specified
	if showAll {
		for _, label := range uniqueNodeLabels {
			// check to see if already picked up in the 'group by' or 'display' sets
			if (!groupMap[label]) && (!displayMap[label]) {
				remainderLabels = append(remainderLabels, label)
			}
		}
	}

	return groupByLabels, displayLabels, remainderLabels, nil
}
