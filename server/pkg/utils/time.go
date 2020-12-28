package utils

import "time"

// function to return current time stamp in UTC
func nowUTC() time.Time {
	return time.Now().UTC()
}
