package batch

import (
	"fmt"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type Stage string

const (
	StageValidate  Stage = "validate"
	StageCalculate Stage = "calculate"
	StageBatch     Stage = "batch"
)

const (
	CodeTooManyRows       = "too-many-rows"
	CodeBadChunkSize      = "bad-chunk-size"
	CodeBadWorkers        = "bad-workers"
	CodeBatchLimitReached = "batch-limit-reached"
	CodeBatchCancelled    = "batch-cancelled"
)

const (
	RuleCalculation = "calculation"
	RuleBatch       = "batch"
)

const (
	DefaultChunkSize = 100
	MaxChunkSize     = 10000
	DefaultWorkers   = 1
	MaxWorkers       = 64
	DefaultMaxRows   = 10000
)

type RowResult struct {
	Row       int
	Value     formula.Value
	Problems  []problem.Problem
	StoppedAt Stage
}

func (r RowResult) HasValue() bool {
	return len(r.Problems) == 0
}

type Options struct {
	ChunkSize int
	Workers   int
	Shared    function.Row
	MaxRows   int
	Cost      int

	beforeChunk func(index int)
	afterChunk  func(index int)
}

type Summary struct {
	FormulaHash    string
	Rows           int
	Valued         int
	Stopped        int
	NotRun         int
	ProblemsByCode map[string]int
	ChunkSize      int
	Workers        int
	Duration       time.Duration
}

type Result struct {
	Rows    []RowResult
	Summary Summary
}

type BatchError struct {
	Code  string
	Limit int
	Got   int
}

func (e *BatchError) Error() string {
	switch e.Code {
	case CodeTooManyRows:
		return fmt.Sprintf("This batch has %d rows and at most %d are allowed.", e.Got, e.Limit)
	case CodeBadChunkSize:
		return fmt.Sprintf("The chunk size is %d and it must be from 1 to %d.", e.Got, e.Limit)
	case CodeBadWorkers:
		return fmt.Sprintf("The number of workers is %d and it must be from 1 to %d.", e.Got, e.Limit)
	}
	return fmt.Sprintf("The batch was refused with the code %s.", e.Code)
}
