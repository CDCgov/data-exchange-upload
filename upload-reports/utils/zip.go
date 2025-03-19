package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
)

func CreateZipArchive(bufOne, bufTwo *bytes.Buffer, bufOneName, bufTwoName string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	bufOneFile, err := zipWriter.Create(bufOneName)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s file in zip: %v", bufOneName, err)
	}
	_, err = bufOneFile.Write(bufOne.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write %s data to zip: %v", bufOneName, err)
	}

	bufTwoFile, err := zipWriter.Create(bufTwoName)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s file in zip: %v", bufTwoName, err)
	}
	_, err = bufTwoFile.Write(bufTwo.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write %s data to zip: %v", bufTwoName, err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %v", err)
	}

	return buf, nil
}
