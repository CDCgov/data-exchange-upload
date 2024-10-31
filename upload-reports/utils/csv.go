package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

func CreateCSV(data [][]string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return &buf, nil
}

func SaveCsvToFile(csvData *bytes.Buffer, outputPath string, filename string) error {
	fullPath := filepath.Join(outputPath, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create the file %v: %v", file, err)
	}
	defer file.Close()

	_, err = file.Write(csvData.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", fullPath, err)
	}

	fmt.Printf("CSV successfully saved to file: %s\n", fullPath)
	return nil
}
