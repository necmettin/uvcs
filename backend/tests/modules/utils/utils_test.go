package utils_test

import (
	"testing"
	"uvcs/modules/utils"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomInt(t *testing.T) {
	// Test that generated numbers are within expected range
	for i := 0; i < 100; i++ {
		num := utils.GenerateRandomInt()
		assert.GreaterOrEqual(t, num, 0)
		assert.Less(t, num, 1000000)
	}

	// Test uniqueness (with high probability)
	nums := make(map[int]bool)
	for i := 0; i < 100; i++ {
		num := utils.GenerateRandomInt()
		assert.False(t, nums[num], "Generated number should be unique")
		nums[num] = true
	}
}

func TestGenerateRandomString(t *testing.T) {
	// Test string length
	for length := 1; length <= 32; length++ {
		str := utils.GenerateRandomString(length)
		assert.Equal(t, length, len(str))
	}

	// Test character validity (alphanumeric)
	for i := 0; i < 100; i++ {
		str := utils.GenerateRandomString(16)
		for _, ch := range str {
			assert.True(t, (ch >= '0' && ch <= '9') ||
				(ch >= 'a' && ch <= 'z') ||
				(ch >= 'A' && ch <= 'Z'))
		}
	}

	// Test uniqueness (with high probability)
	strs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		str := utils.GenerateRandomString(16)
		assert.False(t, strs[str], "Generated string should be unique")
		strs[str] = true
	}
}

func TestHashPassword(t *testing.T) {
	// Test valid password hashing
	password := "mySecurePassword123"
	hash, err := utils.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Test empty password
	hash, err = utils.HashPassword("")
	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestVerifyPassword(t *testing.T) {
	password := "mySecurePassword123"
	hash, _ := utils.HashPassword(password)

	// Test valid password verification
	err := utils.VerifyPassword(hash, password)
	assert.NoError(t, err)

	// Test invalid password
	err = utils.VerifyPassword(hash, "wrongPassword")
	assert.Error(t, err)

	// Test empty password
	err = utils.VerifyPassword(hash, "")
	assert.Error(t, err)

	// Test invalid hash
	err = utils.VerifyPassword("invalid_hash", password)
	assert.Error(t, err)
}

func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		filename string
		isCode   bool
	}{
		{"main.go", true},
		{"script.py", true},
		{"index.js", true},
		{"styles.css", true},
		{"index.html", true},
		{"image.png", false},
		{"document.pdf", false},
		{"archive.zip", false},
		{"", false},
		{".gitignore", true},
		{"Dockerfile", true},
		{"README.md", true},
	}

	for _, test := range tests {
		result := utils.IsCodeFile(test.filename)
		assert.Equal(t, test.isCode, result,
			"IsCodeFile(%s) = %v, want %v",
			test.filename, result, test.isCode)
	}
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		content  []byte
		isBinary bool
	}{
		{[]byte("Hello, World!"), false},
		{[]byte{0x00, 0x01, 0x02}, true},
		{[]byte("Text with null byte\x00"), true},
		{[]byte(""), false},
		{[]byte("Plain text\nwith newlines\r\n"), false},
		{[]byte{0xFF, 0xD8, 0xFF}, true}, // JPEG magic numbers
	}

	for _, test := range tests {
		result := utils.IsBinaryFile(test.content)
		assert.Equal(t, test.isBinary, result)
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"no  extra  spaces", "no extra spaces"},
		{"tabs\tand\tspaces", "tabs and spaces"},
		{"newlines\n\nhere", "newlines here"},
		{"mixed   \t\n   whitespace", "mixed whitespace"},
		{"", ""},
		{"   leading and trailing   ", "leading and trailing"},
		{"one space", "one space"},
		{"\t\t\t", ""},
		{"\n\n\n", ""},
		{"multiple     spaces", "multiple spaces"},
	}

	for _, test := range tests {
		result := utils.NormalizeWhitespace(test.input)
		assert.Equal(t, test.expected, result,
			"NormalizeWhitespace(%q) = %q, want %q",
			test.input, result, test.expected)
	}
}
