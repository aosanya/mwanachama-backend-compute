.PHONY: build test vet clean fuzz fuzz-nightly

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf bin/

fuzz:
	GOMEMLIMIT=1GiB go test -fuzz=FuzzParse -fuzztime=60s ./hostile_test

fuzz-nightly:
	GOMEMLIMIT=1GiB go test -fuzz=FuzzParse -fuzztime=30m ./hostile_test
