package capacity

func (cm *clusterMetric) getUniquePodLabels() (result []string) {

	for _, pod := range cm.rawPodList {
		for k, _ := range pod.Labels {
			found := false
			for _, b := range result {
				if k == b {
					found = true
					break
				}
			}
			if !found {
				result = append(result, k)
			}
		}
	}

	return result
}

func processPodLabelToDisplay(cm *clusterMetric) ([]string, error) {

	// if nothing is called for exit
	//	if groupBy == "" && display == "" && !showAll {
	//		// only populate 'uniqueNodeLabels' if one of the the selection criteria is set
	//		return []string{}, []string{}, []string{}, nil
	//	}

	//  badLabels := []string{}
	//  displayLabels := []string{}
	// remainderLabels := []string{}
	//
	uniquePodLabels := cm.getUniquePodLabels()
	//
	//	// process any 'group by' labels 1st
	//	if groupBy != "" {
	//		unknownLabels := false
	//		for _, label := range strings.Split(groupBy, ",") {
	//			found := false
	//			for _, existingLabel := range uniqueNodeLabels {
	//				if label == existingLabel {
	//					groupByLabels = append(groupByLabels, label)
	//					found = true
	//					break
	//				}
	//			}
	//			if !found {
	//				badLabels = append(badLabels, label)
	//				unknownLabels = true
	//			}
	//		}
	//		if unknownLabels {
	//			return []string{}, []string{}, []string{}, fmt.Errorf("unknown (group by) node label(s): %v", badLabels)
	//		}
	//	}
	//
	//	// process any 'display' node labels next
	//	if display != "" {
	//		unknownLabels := false
	//		for _, label := range strings.Split(display, ",") {
	//			found := false
	//			for _, existingLabel := range uniqueNodeLabels {
	//				if label == existingLabel {
	//					// check to see if already picked up in group by set
	//					inGroupBy := false
	//					for _, groupedLabel := range groupByLabels {
	//						if label == groupedLabel {
	//							inGroupBy = true
	//							break
	//						}
	//					}
	//					// if not add it to the display
	//					if !inGroupBy {
	//						displayLabels = append(displayLabels, label)
	//					}
	//					found = true
	//					break
	//				}
	//			}
	//			if !found {
	//				badLabels = append(badLabels, label)
	//				unknownLabels = true
	//			}
	//		}
	//		if unknownLabels {
	//			return []string{}, []string{}, []string{}, fmt.Errorf("unknown (display) node label(s): %v", badLabels)
	//		}
	//	}
	//
	//	// pick up all node labels not previously selected if 'showAll' is specified
	//	if showAll {
	//		for _, label := range uniqueNodeLabels {
	//			// check to see if already picked up in the 'group by' or 'display' sets
	//			found := false
	//			for _, check := range append(groupByLabels, displayLabels...) {
	//				if label == check {
	//					found = true
	//					break
	//				}
	//			}
	//			// if not, add it to the remainder list
	//			if !found {
	//				remainderLabels = append(remainderLabels, label)
	//			}
	//		}
	//	}

	return uniquePodLabels, nil
}
