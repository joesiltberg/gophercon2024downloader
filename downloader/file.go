package downloader

import "os"

// createEmptyFile creates an empty file in given size
// On success a handle to the open file will be returned.
// On failure the file will be deleted if it had been created, and
// an error is returned.
func createEmptyFile(path string, size int64) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	cleanUp := func() {
		file.Close()
		os.Remove(path)
	}

	if size == 0 {
		return file, nil
	}

	_, err = file.Seek(size-1, os.SEEK_SET)
	if err != nil {
		cleanUp()
		return nil, err
	}

	_, err = file.Write([]byte{0})
	if err != nil {
		cleanUp()
		return nil, err
	}

	return file, nil
}
