.PHONY: test nightly

test:
	go test -race -count=1 .

nightly:
	LAWS_CASES=20000 LAWS_SEED=$$(date +%s) go test -count=1 -v -run 'TestLaw' . | grep -E '^(---|ok|FAIL|PASS|\s+[a-z_]+_test.go:[0-9]+: [0-9]+ cases|\s+[a-z_]+_test.go.*(moved|differs|failed|calculated|incomplete|changed))'
