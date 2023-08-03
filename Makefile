.PHONY: deps build run lint mocks run-mainnet-online run-mainnet-offline run-testnet-online \
	run-testnet-offline check-comments add-license check-license shorten-lines test \
	coverage spellcheck salus build-local coverage-local format check-format

ADDLICENSE_INSTALL=go install github.com/google/addlicense@latest
ADDLICENSE_CMD=addlicense
ADDLICENSE_IGNORE=-ignore ".github/**/*" -ignore ".idea/**/*"
ADDLICENCE_SCRIPT=${ADDLICENSE_CMD} -c "Coinbase, Inc." -l "apache" -v ${ADDLICENSE_IGNORE}
SPELLCHECK_CMD=go run github.com/client9/misspell/cmd/misspell
GOLINES_INSTALL=go install github.com/segmentio/golines@latest
GOLINES_CMD=golines
GOLINT_INSTALL=go get golang.org/x/lint/golint
GOLINT_CMD=golint
GOVERALLS_INSTALL=go install github.com/mattn/goveralls@latest
GOVERALLS_CMD=goveralls
GOIMPORTS_CMD=go run golang.org/x/tools/cmd/goimports
GO_PACKAGES=./services/... ./indexer/... ./thought/... ./configuration/...
GO_FOLDERS=$(shell echo ${GO_PACKAGES} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")
TEST_SCRIPT=go test ${GO_PACKAGES}
LINT_SETTINGS=golint,misspell,gocyclo,gocritic,whitespace,goconst,gocognit,bodyclose,unconvert,lll,unparam
PWD=$(shell pwd)
NOFILE=100000
RECIPIENT=m6cWARU685GQfQos7qNg6isWGCPjvCwyxm

deps:
	go get ./...

build:
	docker build -t rosetta-thought:latest https://github.com/philgrim2/rosetta-thought.git

build-local:
	docker build -t rosetta-thought:latest .

build-release:
	# make sure to always set version with vX.X.X
	docker build -t rosetta-thought:$(version) .;
	docker save rosetta-thought:$(version) | gzip > rosetta-thought-$(version).tar.gz;

run-mainnet-online:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -v "${PWD}/thought-data:/data" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -e "RECIPIENT=${RECIPIENT}" -p 8080:8080 -p 10618:10618 rosetta-thought:latest

run-mainnet-offline:
	docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=MAINNET" -e "PORT=8081" -p 8081:8081 rosetta-thought:latest

run-testnet-online:
	docker run -it --ulimit "nofile=${NOFILE}:${NOFILE}" -v "${PWD}/thought-data:/data" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -e "RECIPIENT=${RECIPIENT}" -p 8080:8080 -p 11618:11618 rosetta-thought:latest

run-testnet-offline:
	docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=TESTNET" -e "PORT=8081" -p 8081:8081 rosetta-thought:latest

train:
	./zstd-train.sh $(network) transaction $(data-directory)

check-comments:
	${GOLINT_INSTALL}
	${GOLINT_CMD} -set_exit_status ${GO_FOLDERS} .
	go mod tidy

lint: | check-comments
	golangci-lint run --timeout 2m0s -v -E ${LINT_SETTINGS},gomnd

add-license:
	${ADDLICENSE_INSTALL}
	${ADDLICENCE_SCRIPT} .

check-license:
	${ADDLICENSE_INSTALL}
	${ADDLICENCE_SCRIPT} -check .

shorten-lines:
	${GOLINES_INSTALL}
	${GOLINES_CMD} -w --shorten-comments ${GO_FOLDERS} .

format:
	gofmt -s -w -l .
	${GOIMPORTS_CMD} -w .

check-format:
	! gofmt -s -l . | read
	! ${GOIMPORTS_CMD} -l . | read

test:
	${TEST_SCRIPT}

coverage:
	${GOVERALLS_INSTALL}
	if [ "${COVERALLS_TOKEN}" ]; then ${TEST_SCRIPT} -coverprofile=c.out -covermode=count; ${GOVERALLS_CMD} -coverprofile=c.out -repotoken ${COVERALLS_TOKEN}; fi

coverage-local:
	${TEST_SCRIPT} -cover

salus:
	docker run --rm -t -v ${PWD}:/home/repo coinbase/salus

spellcheck:
	${SPELLCHECK_CMD} -error .

mocks:
	rm -rf mocks;
	mockery --dir indexer --all --case underscore --outpkg indexer --output mocks/indexer;
	mockery --dir services --all --case underscore --outpkg services --output mocks/services;
	${ADDLICENSE_INSTALL}
	${ADDLICENCE_SCRIPT} .;
