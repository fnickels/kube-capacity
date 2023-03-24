package capacity

type DisplayCriteria struct {
	showContainers    bool
	showPods          bool
	showUtil          bool
	showPodCount      bool
	showDebug         bool
	displayNodeLabels string
	groupByNodeLabels string
	showAllNodeLabels bool
	showAllPodLabels  bool
	selectPodLabels   string
	podLabels         string
	nodeLabels        string
	namespaceLabels   string
	namespace         string
	kubeContext       string
	kubeConfig        string
	outputFormat      string
	sortBy            string
	availableFormat   bool
	binpackAnalysis   bool
	showPodSummary    bool
}
