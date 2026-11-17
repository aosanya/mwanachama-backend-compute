.PHONY: test speed-check speed-targets baseline

test:
	go test -race -count=1 .

speed-check:
	SPEED_CHECK=1 go test -count=1 -v -run "TestSpeedCheck" .

speed-targets:
	SPEED_CHECK=1 go test -count=1 -v -run "TestSpeedTargets" .

baseline:
	SPEED_CHECK=1 SPEED_RECORD=1 go test -count=1 -v -run "TestSpeedRecord" .
