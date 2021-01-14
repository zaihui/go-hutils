VERSION=0.0.3
COVERAGE_FILE=.coverage

##################################
#######       Setup       ########
##################################
.PHONY: setup

setup:
	@go mod download

##################################
#######        Tool       ########
##################################
.PHONY: fmt lint clean

fmt:
	@gofmt -d -w -e .

lint:
	@golangci-lint run ./...

clean:
	@git clean -fdx ${COVERAGE_FILE}

##################################
#######     Coverage      ########
##################################
.PHONY: test coverage coverage_html

TEST_FLAGS = -v -race -failfast -covermode=atomic

test:
	@go test ${TEST_FLAGS} -coverprofile=${COVERAGE_FILE} -coverpkg=./... -timeout=10s ./...

coverage: test
	@go tool cover -func ${COVERAGE_FILE}

coverage_html: test
	@go tool cover -html ${COVERAGE_FILE}


##################################
#######     Release       ########
##################################
release:
	@bumpversion patch
