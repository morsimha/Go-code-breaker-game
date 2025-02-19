package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGuess(t *testing.T) {
	/* Example inputs:
	valid: "007", "81", " 10  "
	invalid: "$", "-15", " "
	*/

	guess, err := ValidateGuess("20")
	assert.NoError(t, err)
	assert.Equal(t, 20, guess)

}

func TestCheckGuessCorrectness(t *testing.T) {
	isCorrect := CheckGuessCorrectness(10)
	assert.Equal(t, isCorrect, false)
}
