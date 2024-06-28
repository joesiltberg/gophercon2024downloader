package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// resourceSize gets the size of an HTTP resource by issuing an HTTP HEAD request
func resourceSize(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, err
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	return response.ContentLength, nil
}

// getChunk downloads a range of bytes from a URL
// The data received is written to the corresponding range in destination.
func getChunk(ctx context.Context, destination io.WriterAt, url string, offset, size int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	rangeString := fmt.Sprintf("bytes=%d-%d", offset, offset+size-1)
	req.Header.Set("Range", rangeString)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status in range response: %d (expected %d)",
			response.StatusCode, http.StatusPartialContent)
	}

	if response.ContentLength != size {
		return fmt.Errorf("unexpected Content-Length in range response: %d (expected %d)",
			response.ContentLength, size)
	}

	writer := io.NewOffsetWriter(destination, offset)
	written, err := io.Copy(writer, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write response: %s", err.Error())
	} else if written != size {
		return fmt.Errorf("wrote %d, expected to write %d", written, size)
	}

	return nil
}
