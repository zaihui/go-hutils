VERSION=0.0.21
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

TEST_FLAGS = -v -race -failfast

test:
	@go test ${TEST_FLAGS}  -timeout=10s ./...

coverage: test
	@go tool cover -func ${COVERAGE_FILE}
	@COVERAGE=$$(go tool cover -func=${COVERAGE_FILE} | grep total | grep statements |awk '{print $$3}' | sed 's/\%//g'); \
	echo "Current coverage is $${COVERAGE}%, minimal is ${MINIMAL_COVERAGE}."; \
	awk "BEGIN {exit ($${COVERAGE} < ${MINIMAL_COVERAGE})}"

coverage_html: test
	@go tool cover -html ${COVERAGE_FILE}


##################################
#######     Release       ########
##################################
release:
	@bumpversion patch
