.PHONY: build test test-race vet fmt-check comment-scan ci clean fuzz fuzz-nightly nightly

NIGHTLY_FUZZTIME ?= 30m

build:
	go build ./...

test:
	go test ./...

test-race:
	go test -race -count=1 ./...

vet:
	go vet ./...

fmt-check:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "$$out"; exit 1; fi

comment-scan:
	go run ./cmd/commentscan .

ci: vet fmt-check comment-scan test-race

clean:
	rm -rf bin/

fuzz:
	GOMEMLIMIT=1GiB go test -fuzz=FuzzParse -fuzztime=60s ./hostile_test

fuzz-nightly:
	GOMEMLIMIT=1GiB go test -fuzz=FuzzParse -fuzztime=$(NIGHTLY_FUZZTIME) ./hostile_test

nightly:
	$(MAKE) -C laws nightly
	HOSTILE_SEED=$$(date +%s) go test -count=1 -race -v -run Generated ./hostile_test
	$(MAKE) fuzz-nightly
