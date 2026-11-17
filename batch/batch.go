package batch

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/function"
)

type plan struct {
	rows      int
	chunkSize int
	workers   int
	chunks    int
}

func (c *Compiled) RunBatch(ctx context.Context, rows []function.Row, options Options) (Result, error) {
	started := time.Now()
	planned, refusal := planBatch(len(rows), options)
	if refusal != nil {
		return Result{}, refusal
	}
	results := make([]RowResult, len(rows))
	var next, spent atomic.Int64
	var wait sync.WaitGroup
	for worker := 0; worker < planned.workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for {
				index := int(next.Add(1))
				if index > planned.chunks {
					return
				}
				c.runChunk(ctx, rows, results, index, planned, options, &spent)
			}
		}()
	}
	wait.Wait()
	for i := range results {
		results[i].Row = i + 1
	}
	return Result{Rows: results, Summary: c.summarize(results, planned, time.Since(started))}, nil
}

func planBatch(rows int, options Options) (plan, *BatchError) {
	maxRows := options.MaxRows
	if maxRows <= 0 {
		maxRows = DefaultMaxRows
	}
	if rows > maxRows {
		return plan{}, &BatchError{Code: CodeTooManyRows, Limit: maxRows, Got: rows}
	}
	chunkLimit := min(MaxChunkSize, maxRows)
	chunkSize := options.ChunkSize
	if chunkSize == 0 {
		chunkSize = min(DefaultChunkSize, chunkLimit)
	}
	if chunkSize < 1 || chunkSize > chunkLimit {
		return plan{}, &BatchError{Code: CodeBadChunkSize, Limit: chunkLimit, Got: chunkSize}
	}
	workers := options.Workers
	if workers == 0 {
		workers = DefaultWorkers
	}
	if workers < 1 || workers > MaxWorkers {
		return plan{}, &BatchError{Code: CodeBadWorkers, Limit: MaxWorkers, Got: workers}
	}
	return plan{rows: rows, chunkSize: chunkSize, workers: workers, chunks: (rows + chunkSize - 1) / chunkSize}, nil
}

func (c *Compiled) runChunk(ctx context.Context, rows []function.Row, results []RowResult, index int, planned plan, options Options, spent *atomic.Int64) {
	if options.beforeChunk != nil {
		options.beforeChunk(index)
	}
	from := (index - 1) * planned.chunkSize
	to := min(index*planned.chunkSize, planned.rows)
	refusal, refused := refusalOf(ctx, options, spent)
	for position := from; position < to; position++ {
		if refused {
			results[position] = refusal
			continue
		}
		result, cost := c.runRow(Merge(options.Shared, rows[position]))
		results[position] = result
		spent.Add(int64(cost))
	}
	if options.afterChunk != nil {
		options.afterChunk(index)
	}
}

func refusalOf(ctx context.Context, options Options, spent *atomic.Int64) (RowResult, bool) {
	if ctx.Err() != nil {
		return notRun(CodeBatchCancelled, batchCancelSentence), true
	}
	if options.Cost > 0 && spent.Load() >= int64(options.Cost) {
		return notRun(CodeBatchLimitReached, batchLimitSentence), true
	}
	return RowResult{}, false
}

func (c *Compiled) summarize(results []RowResult, planned plan, duration time.Duration) Summary {
	summary := Summary{
		FormulaHash:    c.hash,
		Rows:           len(results),
		ProblemsByCode: map[string]int{},
		ChunkSize:      planned.chunkSize,
		Workers:        planned.workers,
		Duration:       duration,
	}
	for _, found := range results {
		switch {
		case found.StoppedAt == StageBatch:
			summary.NotRun++
		case len(found.Problems) > 0:
			summary.Stopped++
		default:
			summary.Valued++
		}
		for _, one := range found.Problems {
			summary.ProblemsByCode[one.Code]++
		}
	}
	return summary
}
