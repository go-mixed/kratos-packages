package utils

import "regexp"

var dateRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var timeRegexp = regexp.MustCompile(`^[012][0-9]:[0-5][0-9]:[0-5][0-9]$`)
var dateTimeRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} [012][0-9]:[0-5][0-9]:[0-5][0-9]$`)

// IsTime checks if the string contains a time
func IsTime(date string) bool {
	return timeRegexp.MatchString(date)
}

// IsDate checks if the string contains a date
func IsDate(date string) bool {
	return dateRegexp.MatchString(date)
}

// IsDateTime checks if the string is a date time
func IsDateTime(date string) bool {
	return dateTimeRegexp.MatchString(date)
}

// IsDateTimeOrDate checks if the string is a date time or date
func IsDateTimeOrDate(date string) bool {
	return IsDateTime(date) || IsDate(date)
}
