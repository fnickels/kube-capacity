package capacity

const VoidValue = "*"
const CSVStringTerminator = "\""

const PodAppNameLabelDefaultSelector = "appname,app,name,app.kubernetes.io/name,k8s-app"

const (
	//TableOutput is the constant value for output type table
	TableOutput string = "table"
	//CSVOutput is the constant value for output type csv
	CSVOutput string = "csv"
	//TSVOutput is the constant value for output type csv
	TSVOutput string = "tsv"
	//JSONOutput is the constant value for output type JSON
	JSONOutput string = "json"
	//YAMLOutput is the constant value for output type YAML
	YAMLOutput string = "yaml"
)

// SupportedOutputs returns a string list of output formats supposed by this package
func SupportedOutputs() []string {
	return []string{
		TableOutput,
		CSVOutput,
		TSVOutput,
		JSONOutput,
		YAMLOutput,
	}
}

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

func sliceFilledWithString(size int, str string) []string {
	data := make([]string, size)
	for i := 0; i < size; i++ {
		data[i] = str
	}
	return data
}
