// Package runner — T-005-44: worker pool implementation.
//
// This file contains the workerPool type which manages N goroutines
// that consume story execution requests from a channel and report
// their results back to the scheduler via a completion channel.
// The model follows ADR 0019 §"Worker-pool model".
package runner

import (
	"context"
	"sync"
)

// workItem is a single unit of work dispatched from the scheduler
// to a worker goroutine. It carries the node identifier and the
// execution function that the worker calls.
type workItem struct {
	// nodeID is the DAG node identifier for this work item.
	nodeID string

	// execute is the function the worker calls to run the story.
	// It receives the parent context and returns the StoryResult.
	// The function must not mutate scheduler state; results are
	// returned via completionCh.
	execute func(ctx context.Context) StoryResult
}

// completionItem is the report a worker sends back to the scheduler
// after finishing a work item.
type completionItem struct {
	// nodeID is the DAG node identifier that was processed.
	nodeID string

	// result is the terminal StoryResult for the node.
	result StoryResult
}

// workerPool manages a fixed pool of worker goroutines.
type workerPool struct {
	// workCh is the channel through which the scheduler feeds
	// work items to workers.
	workCh chan workItem

	// completionCh is the channel through which workers report
	// completed items back to the scheduler.
	completionCh chan completionItem

	// wg tracks live worker goroutines so that the scheduler can
	// wait for all workers to exit before returning.
	wg sync.WaitGroup
}

// newWorkerPool creates and starts workerCount worker goroutines.
// Workers read from workCh and write to completionCh until workCh
// is closed.
//
// When workerCount is zero or negative, a single worker is started
// (sequential execution mode, default per ADR 0019).
func newWorkerPool(
	ctx context.Context,
	workerCount int,
) *workerPool {
	if workerCount <= 0 {
		workerCount = 1
	}

	// Buffer the completion channel so workers never block writing
	// results even if the scheduler is briefly occupied.
	pool := &workerPool{
		workCh:       make(chan workItem, workerCount),
		completionCh: make(chan completionItem, workerCount),
		wg:           sync.WaitGroup{},
	}

	pool.wg.Add(workerCount)

	for range workerCount {
		go pool.runWorker(ctx)
	}

	return pool
}

// runWorker is the goroutine body for a single worker. It reads
// work items from workCh, executes them, and writes the result to
// completionCh. It exits when workCh is closed or ctx is cancelled.
func (wp *workerPool) runWorker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case item, ok := <-wp.workCh:
			if !ok {
				return
			}

			result := item.execute(ctx)

			wp.completionCh <- completionItem{
				nodeID: item.nodeID,
				result: result,
			}
		}
	}
}

// submit sends a work item to the pool. It blocks until a worker
// picks up the item or ctx is cancelled.
//
// Submit returns false if ctx was cancelled before the item could
// be submitted; the caller should stop dispatching in that case.
func (wp *workerPool) submit(
	ctx context.Context,
	item workItem,
) bool {
	select {
	case <-ctx.Done():
		return false

	case wp.workCh <- item:
		return true
	}
}

// close signals all workers to stop by closing workCh, then waits
// for all goroutines to exit. close must be called exactly once,
// after the scheduler has finished dispatching all work items.
func (wp *workerPool) close() {
	close(wp.workCh)
	wp.wg.Wait()
	close(wp.completionCh)
}
