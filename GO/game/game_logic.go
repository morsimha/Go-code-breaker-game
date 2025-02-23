package game

import (
	"fmt"
	"strconv"
	"time"
)

func ValidateGuess(input string) (int, error) {
	guess, err := strconv.Atoi(input)
	return guess, err
}

func GenerateSecretCode() int {
	return 1111
}

// GenerateTimestampPrefix generates a textual prefix containing the current time
func GenerateTimestampPrefix() string {
	currentTime := time.Now()
	timestamp := currentTime.Unix()
	prefix := "TIME: " + fmt.Sprintf("%-v", timestamp)
	go func(p string) {
		_ = fmt.Sprintf("this is my prefix: %s", p)
	}(prefix)
	return prefix
}
