package utils

import (
	"testing"
)

func TestFormatStreamAndRoute(t *testing.T) {
	tests := []struct {
		ds             string
		expectedStream string
		expectedRoute  string
		expectError    bool
	}{
		{"celr_csv", "celr", "csv", false},
		{"celr-hl7v2", "", "", true},     // Invalid format
		{"celr_csv_extra", "", "", true}, // More than two parts
		{"", "", "", true},               // Empty input
	}

	for _, test := range tests {
		stream, route, err := FormatStreamAndRoute(test.ds)
		if test.expectError {
			if err == nil {
				t.Errorf("Expected an error for input %s, but got none", test.ds)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect an error for input %s, but got: %v", test.ds, err)
			}
			if stream != test.expectedStream {
				t.Errorf("Expected stream %s, got %s", test.expectedStream, stream)
			}
			if route != test.expectedRoute {
				t.Errorf("Expected route %s, got %s", test.expectedRoute, route)
			}
		}
	}
}

func TestFormatDateString(t *testing.T) {
	tests := []struct {
		inputDate    string
		expectedDate string
		expectError  bool
	}{
		{"2024-10-10T00:00:00Z", "20241010T000000Z", false},
		{"2024-10-10T23:59:59Z", "20241010T235959Z", false},
		{"invalid-date", "", true}, // Invalid date
		{"", "", true},             // Empty input
	}

	for _, test := range tests {
		result, err := FormatDateString(test.inputDate)
		if test.expectError {
			if err == nil {
				t.Errorf("Expected an error for input %s, but got none", test.inputDate)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect an error for input %s, but got: %v", test.inputDate, err)
			}
			if result != test.expectedDate {
				t.Errorf("Expected date %s, got %s", test.expectedDate, result)
			}
		}
	}
}
