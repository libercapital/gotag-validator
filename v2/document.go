package gotag_validator

import (
	"bytes"
	"regexp"
	"strconv"
	"unicode"
)

//Font: https://github.com/paemuri/brdoc/blob/master/cpfcnpj.go

// Regexp pattern for CPF and CNPJ.
var (
	CPFRegexp  = regexp.MustCompile(`^\d{3}\.?\d{3}\.?\d{3}-?\d{2}$`)
	CNPJRegexp = regexp.MustCompile(`(?i)^[A-Z0-9]{2}\.?[A-Z0-9]{3}\.?[A-Z0-9]{3}/?[A-Z0-9]{4}-?\d{2}$`)
)

// IsCPF verifies if the given string is a valid CPF document.
func isCPF(doc string) bool {

	const size = 9

	return isCPFOrCNPJ(doc, CPFRegexp, size)
}

// IsCNPJ verifies if the given string is a valid CNPJ document.
func isCNPJ(doc string) bool {

	const size = 12

	if !CNPJRegexp.MatchString(doc) {
		return false
	}

	cleanSeparators(&doc)

	if allEq(doc) {
		return false
	}

	dig1 := calculateDigit(doc[:size], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	dig2 := calculateDigit(doc[:size+1], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})

	return dig1 == int(doc[size]-'0') && dig2 == int(doc[size+1]-'0')
}

// isCPFOrCNPJ generates the digits for a given CPF and compares it with the original digits.
func isCPFOrCNPJ(doc string, pattern *regexp.Regexp, size int) bool {

	if !pattern.MatchString(doc) {
		return false
	}

	cleanNonDigits(&doc)

	// Invalidates documents with all digits equal.
	if allEq(doc) {
		return false
	}

	d := doc[:size]
	digit := calculateDigit(d, []int{10, 9, 8, 7, 6, 5, 4, 3, 2})

	d = d + strconv.Itoa(digit)
	digit2 := calculateDigit(d, []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2})

	return doc == d+strconv.Itoa(digit2)
}

// cleanNonDigits removes every rune that is not a digit.
func cleanNonDigits(doc *string) {

	buf := bytes.NewBufferString("")
	for _, r := range *doc {
		if unicode.IsDigit(r) {
			buf.WriteRune(r)
		}
	}

	*doc = buf.String()
}

// cleanSeparators removes formatting separators (., -, /) while preserving alphanumeric characters.
// Used for CNPJ which may contain uppercase letters in the new alphanumeric format.
func cleanSeparators(doc *string) {

	buf := bytes.NewBufferString("")
	for _, r := range *doc {
		if r != '.' && r != '-' && r != '/' {
			buf.WriteRune(unicode.ToUpper(r))
		}
	}

	*doc = buf.String()
}

// allEq checks if every rune in a given string is equal.
func allEq(doc string) bool {

	base := doc[0]
	for i := 1; i < len(doc); i++ {
		if base != doc[i] {
			return false
		}
	}

	return true
}

// calculateDigit calculates the next digit for the given document using explicit weights.
func calculateDigit(doc string, weights []int) int {

	sum := 0
	for i := 0; i < len(doc); i++ {
		sum += charValue(doc[i]) * weights[i]
	}

	rest := sum % 11
	if rest < 2 {
		return 0
	}
	return 11 - rest
}

// charValue converts a byte to its numeric value for check digit calculation.
// Digits 0-9 map to values 0-9; uppercase letters A-Z map to 17-42 (ASCII value minus 48).
// This follows the Receita Federal IN 2229/2024 specification for alphanumeric CNPJ.
func charValue(c byte) int {
	return int(c) - 48
}
