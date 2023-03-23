test: kube-capacity
	./kube-capacity -o table --debug -u                 > test-debug.out
	./kube-capacity -o table --pod-summary --pod-count              > test-summary.out
	./kube-capacity -o table --pod-summary --pod-count -u           > test-summary-util.out
	./kube-capacity -o table --pod-summary --pod-count -p -u        > test-summary-pods.out
	./kube-capacity -o table --pod-summary --pod-count -p -c -u     > test-summary-containers.out
	./kube-capacity -o table -b -u       --pod-count                   > test-table.out
	./kube-capacity -o table -b -p -c -u --pod-count --show-all-labels > test-table-full.out
	./kube-capacity -o tsv   -b -p -c -u --pod-count --show-all-labels > test-table-full.tsv
	./kube-capacity -o json  -b -p -c -u --pod-count --show-all-labels > test-table-full.json

help: kube-capacity
	./kube-capacity --help

kube-capacity: go.mod go.sum Makefile main.go pkg/*/*.go
	go build .