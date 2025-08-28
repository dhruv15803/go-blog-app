package utils

import (
	"strings"
	"unicode/utf8"
)

func IsPasswordStrong(password string) bool {
	//	strong password characteristics:
	//	1] minimum length = 6
	//	2] atleast one special character
	//	3] atleast 1 uppercase character
	//	4] atleast 1 lowercase char
	//	5] atleast 1 numerical char

	if utf8.RuneCountInString(password) < 6 {
		return false
	}

	const SPECIAL_CHARS = "!@#$%^&*()_+-="
	const NUMERICAL_CHARS = "0123456789"
	const UPPERCASE_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const LOWERCASE_CHARS = "abcdefghijklmnopqrstuvwxyz"
	hasSpecialChar := false
	hasNumericalChar := false
	hasUpperChar := false
	hasLowerChar := false

	for _, c := range password {
		if hasSpecialChar && hasNumericalChar && hasUpperChar && hasLowerChar {
			return true
		}

		if !hasSpecialChar && strings.Contains(SPECIAL_CHARS, string(c)) {
			hasSpecialChar = true
		}

		if !hasNumericalChar && strings.Contains(NUMERICAL_CHARS, string(c)) {
			hasNumericalChar = true
		}

		if !hasUpperChar && strings.Contains(UPPERCASE_CHARS, string(c)) {
			hasUpperChar = true
		}

		if !hasLowerChar && strings.Contains(LOWERCASE_CHARS, string(c)) {
			hasLowerChar = true
		}
	}

	return hasSpecialChar && hasNumericalChar && hasUpperChar && hasLowerChar
}

func IsValidEmail(email string) bool {

	if !strings.Contains(email, "@") {
		return false
	}

	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return false
	}
	firstPart, secondPart := emailParts[0], emailParts[1]
	if firstPart == "" || secondPart == "" {
		return false
	}

	if !strings.Contains(secondPart, ".") {
		return false
	}

	if len(strings.Split(secondPart, ".")) != 2 {
		return false
	}

	return true
}
