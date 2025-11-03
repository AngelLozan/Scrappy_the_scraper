package main

import (
	"testing"
)

func TestNormalizeAndRemoveHomoglyphs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal ASCII Exodus",
			input:    "Exodus Wallet",
			expected: "exodus wallet",
		},
		{
			name:     "Cyrillic o in Exodus",
			input:    "Exоdus", // Cyrillic 'о' (U+043E) instead of Latin 'o'
			expected: "exodus",
		},
		{
			name:     "Cyrillic x in Exodus",
			input:    "Eхodus", // Cyrillic 'х' (U+0445) instead of Latin 'x'
			expected: "exodus",
		},
		{
			name:     "Greek O in Exodus",
			input:    "ExΟdus", // Greek 'Ο' (U+039F) instead of Latin 'O'
			expected: "exodus",
		},
		{
			name:     "Greek X in Exodus",
			input:    "EΧodus", // Greek 'Χ' (U+03A7) instead of Latin 'X'
			expected: "exodus",
		},
		{
			name:     "Multiple homoglyphs",
			input:    "Ехоdus", // Cyrillic 'Е' and 'х' and 'о'
			expected: "exodus",
		},
		{
			name:     "Unicode with combining characters",
			input:    "E\u0301xodus", // 'É' as E + combining acute accent
			expected: "exodus",
		},
		{
			name:     "All Cyrillic lookalikes",
			input:    "Ехоdυѕ", // Mix of Cyrillic and Greek
			expected: "exodus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeAndRemoveHomoglyphs(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeAndRemoveHomoglyphs(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsKeywordVariant(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		keyword  string
		expected bool
	}{
		{
			name:     "Normal match",
			text:     "Download Exodus Wallet now",
			keyword:  "exodus",
			expected: true,
		},
		{
			name:     "Cyrillic homoglyph match",
			text:     "Download Exоdus Wallet now", // Cyrillic о
			keyword:  "exodus",
			expected: true,
		},
		{
			name:     "Greek homoglyph match",
			text:     "Download EΧodus Wallet now", // Greek Χ
			keyword:  "exodus",
			expected: true,
		},
		{
			name:     "Mixed homoglyphs match",
			text:     "Get your Ехоdυѕ wallet", // Multiple Cyrillic and Greek
			keyword:  "exodus",
			expected: true,
		},
		{
			name:     "No match",
			text:     "Bitcoin wallet app",
			keyword:  "exodus",
			expected: false,
		},
		{
			name:     "Case insensitive match",
			text:     "EXODUS WALLET",
			keyword:  "exodus",
			expected: true,
		},
		{
			name:     "Crypto keyword with homoglyphs",
			text:     "сrуpto wallet", // Cyrillic с and у
			keyword:  "crypto",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsKeywordVariant(tt.text, tt.keyword)
			if result != tt.expected {
				t.Errorf("containsKeywordVariant(%q, %q) = %v, want %v", tt.text, tt.keyword, result, tt.expected)
			}
		})
	}
}
