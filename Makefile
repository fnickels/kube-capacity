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


#################
# Go Procedures #
#################

##@ Go Support & Helper targets
.PHONY: goversion
goversion: ## go version 
	@go version

.PHONY: goenv
goenv: goversion ## go version & env
	go env

.PHONY: gotidy
gotidy: ## go mod tidy - cleans up go modules
	go mod tidy

.PHONY: govendor
govendor: gotidy ## go mod tidy & vendor - cleans up go modules and creates / updates vendor directory
	go mod vendor
	go mod verify

## Looks for newer versions of modules
.PHONY: goupdatemods
goupdatemods: goenv ## Update Go Modules
	go get -u

.PHONY: refresh
refresh: goupdatemods gotidy govendor ## Update Modules, Tidy up Modules, & Generate vendor dir

.PHONY: fullrefresh
fullrefresh: goclean goupdatemods gotidy govendor ## Clean Go Caches and the 'refresh'

.PHONY: goclean
goclean: goenv ## Clean Go Modules caches
	go clean -r -cache -modcache
