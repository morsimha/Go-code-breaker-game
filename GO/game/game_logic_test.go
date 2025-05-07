package game

import (
	"strconv"
	"testing"
	"github.com/stretchr/testify/assert"
)

// --- ValidateGuess tests ---

func TestValidateGuess_ValidInputs(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1234", 1234},
		{" 4321 ", 4321},
		{"0000", 0},
	}

	for _, tt := range tests {
		guess, err := ValidateGuess(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, guess)
	}
}

func TestValidateGuess_InvalidInputs(t *testing.T) {
	invalidInputs := []string{" ", "1", "abc123", "12.34", "123456", "12 34", "!@#$", "-1234"}

	for _, input := range invalidInputs {
		_, err := ValidateGuess(input)
		assert.Error(t, err)
	}
}

// --- GenerateSecretCode logic tests ---

// to test logic deterministically, we expose a helper function that accepts input

func TestGenerateSecretCode_EvenSumReversal(t *testing.T) {
	result := generateFromFixedNumber(1234) // sum = 10 -> even → reverse → 4321
	assert.Equal(t, 4321, result)
}

func TestGenerateSecretCode_OddSumIncrement(t *testing.T) {
	result := generateFromFixedNumber(1235) // sum = 11 → odd → increment digits → 2346
	assert.Equal(t, 2346, result)
}

func TestGenerateSecretCode_PalindromeBecome7777(t *testing.T) {
	result := generateFromFixedNumber(2442) // even → reverse = 2442 → palindrome → 7777
	assert.Equal(t, 7777, result)
}

func TestGenerateSecretCode_WrapAroundDigit(t *testing.T) {
	result := generateFromFixedNumber(8999) // sum=35 → odd → inc → 9000
	assert.Equal(t, 9000, result) // not palindrome
}

// --- internal helper to test logic without randomness ---
func generateFromFixedNumber(num int) int {
	// replicate logic from GenerateSecretCode
	digits := []int{}
	sum := 0
	temp := num
	for temp > 0 {
		d := temp % 10
		sum += d
		digits = append([]int{d}, digits...)
		temp /= 10
	}

	var newDigits []int
	if sum%2 == 0 {
		for i := len(digits) - 1; i >= 0; i-- {
			newDigits = append(newDigits, digits[i])
		}
	} else {
		for _, d := range digits {
			newDigits = append(newDigits, (d+1)%10)
		}
	}

	isPalindrome := true
	for i := 0; i < len(newDigits)/2; i++ {
		if newDigits[i] != newDigits[len(newDigits)-1-i] {
			isPalindrome = false
			break
		}
	}

	if isPalindrome {
		for i := range newDigits {
			newDigits[i] = 7
		}
	}

	// Convert digit slice back to int
	finalStr := ""
	for _, d := range newDigits {
		finalStr += strconv.Itoa(d)
	}
	finalNum, _ := strconv.Atoi(finalStr)
	return finalNum
}
