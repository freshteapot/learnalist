package remind

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func DefaultNowUTC() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func DefaultWhenNextWithLastActiveOffset() (string, string) {
	now := time.Now().UTC()
	whenNext := now.Format(time.RFC3339Nano)
	count := 5
	lastActive := now.Add(time.Duration(-count) * time.Minute).Format(time.RFC3339Nano)
	return whenNext, lastActive
}

func ParseAndValidateTimeOfDay(input string) (string, error) {
	parts := strings.Split(input, ":")

	if len(parts) == 2 {
		input = fmt.Sprintf("%s:%s:00", parts[0], parts[1])
	}

	return input, ValidateTimeOfDay(input)
}

func ValidateTimeOfDay(input string) error {
	fail := errors.New("fail")
	parts := strings.Split(input, ":")

	if len(parts) != 3 {
		return fail
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return fail
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return fail
	}

	second, err := strconv.Atoi(parts[2])
	if err != nil {
		return fail
	}

	// make sure the number is no larger than 2 in length, ugly code
	if len(parts[0]) > 2 {
		return fail
	}

	if len(parts[1]) > 2 {
		return fail
	}

	if len(parts[2]) > 2 {
		return fail
	}

	// Hour first
	if hour < 0 {
		return fail
	}

	if hour > 23 {
		return fail
	}

	// Minute
	if minute < 0 {
		return fail
	}
	if minute > 59 {
		return fail
	}

	// Second
	if second < 0 {
		return fail
	}
	if second > 59 {
		return fail
	}
	return nil
}
