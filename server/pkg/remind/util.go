package remind

import (
	"errors"
	"strconv"
	"strings"
)

func ValidateTimeOfDay(input string) error {
	fail := errors.New("fail")
	parts := strings.Split(input, ":")

	if len(parts) != 2 {
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

	if len(parts[0]) > 2 {
		return fail
	}

	if len(parts[1]) > 2 {
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
	return nil
}
