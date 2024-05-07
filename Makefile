SHELL=/bin/bash
.ONESHELL:
.EXPORT_ALL_VARIABLES:

test:
	go test .

build:
	go build ./cmd/jsonpath/
