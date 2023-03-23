package capacity

const VoidValue = "*"
const CSVStringTerminator = "\""

const PodAppNameLabel = "appname"

func setMultipleVoids(n int) []string {

	if n <= 0 {
		return []string{}
	}

	voids := make([]string, n)

	for i := range voids {
		voids[i] = VoidValue
	}

	return voids
}

func setNodeLabels(labelNames []string, nm *nodeMetric) []string {

	if len(labelNames) <= 0 {
		return []string{}
	}

	labels := make([]string, len(labelNames))

	for i, label := range labelNames {
		labels[i] = nm.nodeLabels[label]
	}

	return labels
}
