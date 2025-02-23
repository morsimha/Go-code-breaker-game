package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGuess(t *testing.T) {
	/* Example inputs:
	valid: "007123", "1181", " 1022  "
	invalid: "$", "-15", " "
	*/

	guess, err := ValidateGuess("2000")
	assert.NoError(t, err)
	assert.Equal(t, 2000, guess)

}

func TestCheckGuessCorrectness(t *testing.T) {
	secretCode := GenerateSecretCode()
	assert.Equal(t, secretCode, 1111)
}
