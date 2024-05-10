SHELL=/bin/bash
.ONESHELL:
.EXPORT_ALL_VARIABLES:

test:
	go test . -v

build:
	go build ./cmd/jsonpath/
