package downloader

import (
	"context"
	"io"
	"os"
	"sync"
)

// A downloadTask is a job for one go routine to do.
type downloadTask struct {
	offset int64
	size   int64
}

// The downloader function is a worker in a worker pool
// The worker will get tasks from a channel and do the tasks until there are
// no remaining tasks. If a task fails the context will be cancelled so the
// other workers stop performing more requests.
func downloader(ctx context.Context, cancel context.CancelCauseFunc, destination io.WriterAt, url string, tasks <-chan downloadTask, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		if ctx.Err() != nil {
			// If the context is cancelled we will continue through all
			// tasks from the channel, but we won't keep trying to download
			// the remaining tasks.
			continue
		}
		err := getChunk(ctx, destination, url, task.offset, task.size)
		if err != nil {
			cancel(err)
		}
	}
}

// ParallelDownload downloads a file from a URL in parallel
func ParallelDownload(ctx context.Context, destination, url string, numGoRoutines int, chunkSize int64) error {
	cancellableCtx, cancel := context.WithCancelCause(ctx)

	totalSize, err := resourceSize(cancellableCtx, url)
	if err != nil {
		return err
	}

	file, err := createEmptyFile(destination, totalSize)
	if err != nil {
		return err
	}

	defer file.Close()

	// Create the worker pool and a channel so we can send tasks to the workers
	var wg sync.WaitGroup
	tasks := make(chan downloadTask)
	for range numGoRoutines {
		wg.Add(1)
		go downloader(cancellableCtx, cancel, file, url, tasks, &wg)
	}

	// Send all tasks to the task channel
	for position := int64(0); position < totalSize; position += chunkSize {
		task := downloadTask{
			offset: position,
			size:   min(chunkSize, totalSize-position),
		}
		tasks <- task
	}
	close(tasks)
	wg.Wait()

	if err := context.Cause(cancellableCtx); err != nil {
		file.Close()
		os.Remove(destination)
		return err
	}

	return nil
}
