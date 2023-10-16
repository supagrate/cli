package utils

import (
	"strconv"
)

func ParseCount(count string) int {
	i, err := strconv.Atoi(count)

	if err != nil {
		return -1
	}

	return i
}
