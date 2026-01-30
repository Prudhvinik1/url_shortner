package main

import "testing"

func TestGenerateShortCode_LengthAndCharset(t *testing.T) {
	code, err := generateShortCode()
	if err != nil {
		t.Fatalf("generateShortCode: %v", err)
	}
	if len(code) != codeLength {
		t.Fatalf("expected length %d, got %d", codeLength, len(code))
	}
	for _, r := range code {
		found := false
		for _, allowed := range charset {
			if r == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("unexpected character in code: %q", r)
		}
	}
}


