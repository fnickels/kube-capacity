package capacity

func (cm *clusterMetric) analyzeCluster(cr *DisplayCriteria) {

	cm.analysis.nodesUnbalanced = 0
	cm.analysis.nodesUnderutilized = 0
	cm.analysis.nodesWellUtilized = 0

	// Look at each node
	for _, nm := range cm.nodeMetrics {

		nm.analysis.nodesUnbalanced = 0
		nm.analysis.nodesUnderutilized = 0
		nm.analysis.nodesWellUtilized = 0

		nodeData := getAnalysisData(nm.cpu, nm.memory, nm.eni, nm.podCount)

		if nodeData.headroom.overall > 20 {
			// under utilized node
			cm.analysis.nodesUnderutilized++
			nm.analysis.nodesUnderutilized++

		} else if nodeData.waste.cpu > 20 || nodeData.waste.memory > 20 {
			// unbalanced node
			cm.analysis.nodesUnbalanced++
			nm.analysis.nodesUnbalanced++
		} else {
			// well balanced node
			cm.analysis.nodesWellUtilized++
			nm.analysis.nodesWellUtilized++
		}
	}
}
