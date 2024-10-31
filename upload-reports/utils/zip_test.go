package utils

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateZipArchive(t *testing.T) {
	bufOne := bytes.NewBufferString("This is the content of the first file.")
	bufTwo := bytes.NewBufferString("This is the content of the second file.")

	bufOneName := "file1.txt"
	bufTwoName := "file2.txt"

	zipBuffer, err := CreateZipArchive(bufOne, bufTwo, bufOneName, bufTwoName)
	assert.NoError(t, err)
	assert.NotNil(t, zipBuffer)

	verifyZipContents(t, zipBuffer, bufOneName, "This is the content of the first file.")
	verifyZipContents(t, zipBuffer, bufTwoName, "This is the content of the second file.")
}

func verifyZipContents(t *testing.T, zipBuffer *bytes.Buffer, expectedFileName string, expectedContent string) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBuffer.Bytes()), int64(zipBuffer.Len()))
	assert.NoError(t, err)

	var found bool
	for _, file := range zipReader.File {
		if file.Name == expectedFileName {
			found = true

			// Open the file within the ZIP
			rc, err := file.Open()
			assert.NoError(t, err)
			defer rc.Close()

			// Read the content of the file
			fileContent, err := io.ReadAll(rc)
			assert.NoError(t, err)
			assert.Equal(t, expectedContent, string(fileContent))
		}
	}

	assert.True(t, found, "Expected file %s not found in ZIP archive", expectedFileName)
}
