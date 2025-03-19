package utils

import (
	"errors"
	"strings"
	"time"
)

func FormatStreamAndRoute(ds string) (string, string, error) {
	streamAndRoute := strings.Split(ds, "_")
	if len(streamAndRoute) != 2 {
		return "", "", errors.New("Data stream passed in does not have correct formatting: %s")
	}

	datastream := streamAndRoute[0]
	route := streamAndRoute[1]

	return datastream, route, nil
}

func FormatDateString(inputDate string) (string, error) {
	parsedDate, err := time.Parse(time.RFC3339, inputDate)
	if err != nil {
		return "", err
	}

	isoDateString := parsedDate.Format("2006-01-02T15:04:05Z")

	cleanedDateString := strings.ReplaceAll(strings.ReplaceAll(isoDateString, ":", ""), "-", "")

	return cleanedDateString, nil
}
