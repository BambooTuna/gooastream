export PATH := bin:${PATH}

.DEFAULT_GOAL := help
.PHONY: help
help: ## show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' ${MAKEFILE_LIST} | sort | awk 'BEGIN {FS = ":.*?## "}; { \
		printf "\033[36m%-20s\033[0m %s\n", $$1, $$2 \
	}'

.PHONY: test
test:
	go test -v ./...

.PHONY: version-patch-up
version-patch-up:
	awk -F '.' '{printf "%d.%d.%d", $$1, $$2, $$3+1 > "version"}' version

.PHONY: version-minor-up
version-minor-up:
	awk -F '.' '{printf "%d.%d.0", $$1, $$2+1 > "version"}' version

.PHONY: version-major-up
version-major-up:
	awk -F '.' '{printf "%d.0.0", $$1+1 > "version"}' version
