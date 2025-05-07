package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Global random generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func ValidateGuess(input string) (int, error) {
	// Trim any whitespace
	trimmedInput := strings.TrimSpace(input)

	// Check that input contains exactly 4 digits
	if len(trimmedInput) != 4 {
		return 0, errors.New("invalid input: must contain exactly 4 digits")
	}

	// Check that all characters are digits
	for _, char := range trimmedInput {
		if char < '0' || char > '9' {
			return 0, errors.New("invalid input: must contain only digits")
		}
	}

	// Parse the validated input
	// We'll use Atoi to get an integer for easier comparison, but
	// we maintain the original string format for display purposes
	guess, err := strconv.Atoi(trimmedInput)
	if err != nil {
		return 0, errors.New("error converting input to number")
	}

	return guess, nil
}

func GenerateSecretCode() int {
	// Generate a random 4-digit number (1000-9999)
	num := rng.Intn(9000) + 1000

	// Calculate the sum of digits
	sum := sumOfDigits(num)

	// Adjust the number based on whether the sum is odd or even
	if sum%2 == 0 {
		// If sum is even, reverse the number
		num = reverseNumber(num)
	} else {
		// If sum is odd, increment each digit by 1 (wrap 9 to 0)
		num = incrementDigits(num)
	}

	// Check if the number is a palindrome
	if isPalindrome(num) {
		// If it's a palindrome, replace all digits with 7
		num = allSevens(num)
	}
	// Print the chosen number for debugging purposes
	fmt.Printf("Generated number: %d\n", num)

	return num
}

// Helper function to calculate the sum of digits
func sumOfDigits(num int) int {
	sum := 0
	n := num
	for n > 0 {
		sum += n % 10
		n /= 10
	}
	return sum
}

// Helper function to reverse a number while preserving leading zeros
func reverseNumber(num int) int {
	// Convert to string with padded zeros to ensure 4 digits
	strNum := fmt.Sprintf("%04d", num)

	// Reverse the string
	reversed := ""
	for i := len(strNum) - 1; i >= 0; i-- {
		reversed += string(strNum[i])
	}

	// Convert back to int
	result, _ := strconv.Atoi(reversed)
	return result
}

// Helper function to increment each digit by 1 (wrap 9 to 0)
func incrementDigits(num int) int {
	// Convert to string with padded zeros to ensure 4 digits
	strNum := fmt.Sprintf("%04d", num)

	// Increment each digit
	result := ""
	for i := 0; i < len(strNum); i++ {
		digit := int(strNum[i] - '0')
		newDigit := (digit + 1) % 10
		result += strconv.Itoa(newDigit)
	}

	// Convert back to int
	resultNum, _ := strconv.Atoi(result)
	return resultNum
}

// Helper function to check if a number is a palindrome
func isPalindrome(num int) bool {
	// Convert to string with padded zeros to ensure 4 digits
	strNum := fmt.Sprintf("%04d", num)

	// Check if it reads the same forward and backward
	for i := 0; i < len(strNum)/2; i++ {
		if strNum[i] != strNum[len(strNum)-1-i] {
			return false
		}
	}
	return true
}

// Helper function to replace all digits with 7
func allSevens(num int) int {
	// Count number of digits
	digits := len(fmt.Sprintf("%d", num))

	// Create a number with all 7s
	result := 0
	for i := 0; i < digits; i++ {
		result = result*10 + 7
	}

	return result
}

// GenerateTimestampPrefix generates a textual prefix containing the current time
func GenerateTimestampPrefix() string {
	currentTime := time.Now()
	timestamp := currentTime.Unix()
	prefix := "TIME: " + fmt.Sprintf("%-v", timestamp)+ " "
	return prefix
}
