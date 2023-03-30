package capacity

type SelectionFilters struct {
	PodLabels       string
	NodeLabels      string
	NamespaceLabels string
	Namespace       string
}

type DisplayCriteria struct {
	ShowContainers    bool
	ShowPods          bool
	ShowUtil          bool
	ShowPodCount      bool
	ShowDebug         bool
	DisplayNodeLabels string
	GroupByNodeLabels string
	ShowAllNodeLabels bool
	DisplayPodLabels  string
	ShowAllPodLabels  bool
	SelectPodLabels   string
	KubeContext       string
	KubeConfig        string
	OutputFormat      string
	SortBy            string
	AvailableFormat   bool
	BinpackAnalysis   bool
	ShowPodSummary    bool
	Filters           SelectionFilters
}

func (cr *DisplayCriteria) ShowNamespace() bool {
	return cr.Filters.Namespace == ""
}
