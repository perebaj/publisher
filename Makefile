

## run all tests. Usage `make test` or `make test testcase="TestFunctionName"` to run an isolated tests
.PHONY: test
test:
	if [ -n "$(testcase)" ]; then \
		go test ./... -timeout 10s -race -run="^$(testcase)$$" -v; \
	else \
		go test ./... -timeout 10s -race; \
	fi

## lint the whole project
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## Display help for all targets
.PHONY: help
help:
	@awk '/^.PHONY: / { \
		msg = match(lastLine, /^## /); \
			if (msg) { \
				cmd = substr($$0, 9, 100); \
				msg = substr(lastLine, 4, 1000); \
				printf "  ${GREEN}%-30s${RESET} %s\n", cmd, msg; \
			} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
