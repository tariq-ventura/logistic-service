package validations

import "strconv"

func ParsePositiveInt(value string, fallback int) int {
	number, err := strconv.Atoi(value)

	if err != nil || number < 1 {
		return fallback
	}

	return number
}
