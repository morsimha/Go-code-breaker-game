package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ValidateGuess(input string) (int, error) {
	// Trimming any whitespace
	trimmedInput := strings.TrimSpace(input)
	
	// Check that input contains exactly 4 digits
	if len(trimmedInput) != 4 {
		return 0, errors.New("invalid input: must contain exactly 4 digits")
	}
	
	// Checking that all characters are actually digits
	for _, char := range trimmedInput {
		if char < '0' || char > '9' {
			return 0, errors.New("invalid input: must contain only digits")
		}
	}
	
	// Convert validated input to integer
	guess, err := strconv.Atoi(trimmedInput)
	if err != nil {
		return 0, errors.New("error converting input to number")
	}
	
	return guess, nil
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
