package keylib

import (
	"strings"
	"testing"
)

// =============================================================================
// GenerateMnemonic Tests
// =============================================================================

func TestGenerateMnemonic(t *testing.T) {
	t.Run("generates 12 word mnemonic", func(t *testing.T) {
		mnemonic, err := GenerateMnemonic(12)
		if err != nil {
			t.Fatalf("GenerateMnemonic(12) failed: %v", err)
		}
		words := strings.Fields(mnemonic)
		if len(words) != 12 {
			t.Errorf("expected 12 words, got %d", len(words))
		}
	})

	t.Run("generates 15 word mnemonic", func(t *testing.T) {
		mnemonic, err := GenerateMnemonic(15)
		if err != nil {
			t.Fatalf("GenerateMnemonic(15) failed: %v", err)
		}
		words := strings.Fields(mnemonic)
		if len(words) != 15 {
			t.Errorf("expected 15 words, got %d", len(words))
		}
	})

	t.Run("generates 18 word mnemonic", func(t *testing.T) {
		mnemonic, err := GenerateMnemonic(18)
		if err != nil {
			t.Fatalf("GenerateMnemonic(18) failed: %v", err)
		}
		words := strings.Fields(mnemonic)
		if len(words) != 18 {
			t.Errorf("expected 18 words, got %d", len(words))
		}
	})

	t.Run("generates 21 word mnemonic", func(t *testing.T) {
		mnemonic, err := GenerateMnemonic(21)
		if err != nil {
			t.Fatalf("GenerateMnemonic(21) failed: %v", err)
		}
		words := strings.Fields(mnemonic)
		if len(words) != 21 {
			t.Errorf("expected 21 words, got %d", len(words))
		}
	})

	t.Run("generates 24 word mnemonic", func(t *testing.T) {
		mnemonic, err := GenerateMnemonic(24)
		if err != nil {
			t.Fatalf("GenerateMnemonic(24) failed: %v", err)
		}
		words := strings.Fields(mnemonic)
		if len(words) != 24 {
			t.Errorf("expected 24 words, got %d", len(words))
		}
	})

	t.Run("rejects invalid word count 10", func(t *testing.T) {
		_, err := GenerateMnemonic(10)
		if err == nil {
			t.Error("GenerateMnemonic(10) should fail")
		}
	})

	t.Run("rejects invalid word count 13", func(t *testing.T) {
		_, err := GenerateMnemonic(13)
		if err == nil {
			t.Error("GenerateMnemonic(13) should fail")
		}
	})

	t.Run("rejects invalid word count 0", func(t *testing.T) {
		_, err := GenerateMnemonic(0)
		if err == nil {
			t.Error("GenerateMnemonic(0) should fail")
		}
	})

	t.Run("generated mnemonic is valid", func(t *testing.T) {
		mnemonic, _ := GenerateMnemonic(12)
		err := ValidateMnemonic(mnemonic)
		if err != nil {
			t.Errorf("generated mnemonic should be valid: %v", err)
		}
	})

	t.Run("generates unique mnemonics", func(t *testing.T) {
		m1, _ := GenerateMnemonic(12)
		m2, _ := GenerateMnemonic(12)
		if m1 == m2 {
			t.Error("generated mnemonics should be unique")
		}
	})
}

// =============================================================================
// ValidateMnemonic Tests
// =============================================================================

func TestValidateMnemonic(t *testing.T) {
	t.Run("validates correct mnemonic", func(t *testing.T) {
		mnemonic, _ := GenerateMnemonic(12)
		err := ValidateMnemonic(mnemonic)
		if err != nil {
			t.Errorf("valid mnemonic should pass validation: %v", err)
		}
	})

	t.Run("rejects invalid word count 11", func(t *testing.T) {
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
		err := ValidateMnemonic(mnemonic)
		if err == nil {
			t.Error("11 word mnemonic should fail validation")
		}
	})

	t.Run("rejects invalid word count 3", func(t *testing.T) {
		mnemonic := "abandon abandon abandon"
		err := ValidateMnemonic(mnemonic)
		if err == nil {
			t.Error("3 word mnemonic should fail validation")
		}
	})

	t.Run("rejects unknown words", func(t *testing.T) {
		// "notaword" is not in BIP39 word list
		mnemonic := "notaword abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
		err := ValidateMnemonic(mnemonic)
		if err == nil {
			t.Error("mnemonic with unknown word should fail validation")
		}
	})

	t.Run("rejects invalid checksum", func(t *testing.T) {
		// Valid words but wrong checksum - last word "ability" creates invalid checksum
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon ability"
		err := ValidateMnemonic(mnemonic)
		if err == nil {
			t.Error("mnemonic with invalid checksum should fail validation")
		}
	})

	t.Run("handles extra whitespace", func(t *testing.T) {
		validMnemonic, _ := GenerateMnemonic(12)
		// Add extra whitespace
		paddedMnemonic := "  " + validMnemonic + "  "
		err := ValidateMnemonic(paddedMnemonic)
		if err != nil {
			t.Errorf("mnemonic with extra whitespace should be valid: %v", err)
		}
	})

	t.Run("handles mixed case", func(t *testing.T) {
		validMnemonic, _ := GenerateMnemonic(12)
		// Convert to uppercase
		upperMnemonic := strings.ToUpper(validMnemonic)
		err := ValidateMnemonic(upperMnemonic)
		if err != nil {
			t.Errorf("mnemonic with uppercase should be valid: %v", err)
		}
	})
}

// =============================================================================
// PrivateKeyFromMnemonic Tests
// =============================================================================

func TestPrivateKeyFromMnemonic(t *testing.T) {
	mnemonic, _ := GenerateMnemonic(12)

	t.Run("derives key from valid mnemonic", func(t *testing.T) {
		key, err := PrivateKeyFromMnemonic(mnemonic)
		if err != nil {
			t.Fatalf("PrivateKeyFromMnemonic failed: %v", err)
		}
		if key.IsZero() {
			t.Error("derived key should not be zero")
		}
	})

	t.Run("derivation is deterministic", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonic(mnemonic)
		key2, _ := PrivateKeyFromMnemonic(mnemonic)
		if !key1.Equal(key2) {
			t.Error("same mnemonic should derive same key")
		}
	})

	t.Run("rejects invalid mnemonic", func(t *testing.T) {
		_, err := PrivateKeyFromMnemonic("invalid mnemonic phrase")
		if err == nil {
			t.Error("invalid mnemonic should fail")
		}
	})

	t.Run("rejects empty mnemonic", func(t *testing.T) {
		_, err := PrivateKeyFromMnemonic("")
		if err == nil {
			t.Error("empty mnemonic should fail")
		}
	})

	t.Run("different mnemonics produce different keys", func(t *testing.T) {
		m1, _ := GenerateMnemonic(12)
		m2, _ := GenerateMnemonic(12)
		key1, _ := PrivateKeyFromMnemonic(m1)
		key2, _ := PrivateKeyFromMnemonic(m2)
		if key1.Equal(key2) {
			t.Error("different mnemonics should produce different keys")
		}
	})
}

// =============================================================================
// PrivateKeyFromMnemonicWithPath Tests
// =============================================================================

func TestPrivateKeyFromMnemonicWithPath(t *testing.T) {
	mnemonic, _ := GenerateMnemonic(12)

	t.Run("default path produces same key", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonic(mnemonic)
		key2, _ := PrivateKeyFromMnemonicWithPath(mnemonic, DefaultDerivationPath)
		if !key1.Equal(key2) {
			t.Error("default path should produce same key as PrivateKeyFromMnemonic")
		}
	})

	t.Run("different paths produce different keys", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonicWithPath(mnemonic, "m/44'/60'/0'/0/0")
		key2, _ := PrivateKeyFromMnemonicWithPath(mnemonic, "m/44'/60'/0'/0/1")
		if key1.Equal(key2) {
			t.Error("different paths should produce different keys")
		}
	})

	t.Run("hardened vs non-hardened produces different keys", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonicWithPath(mnemonic, "m/44'/60'/0'")
		key2, _ := PrivateKeyFromMnemonicWithPath(mnemonic, "m/44/60/0")
		if key1.Equal(key2) {
			t.Error("hardened and non-hardened paths should produce different keys")
		}
	})

	t.Run("rejects invalid path format", func(t *testing.T) {
		_, err := PrivateKeyFromMnemonicWithPath(mnemonic, "invalid")
		if err == nil {
			t.Error("invalid path should fail")
		}
	})

	t.Run("rejects path without m prefix", func(t *testing.T) {
		_, err := PrivateKeyFromMnemonicWithPath(mnemonic, "44'/60'/0'/0/0")
		if err == nil {
			t.Error("path without m/ prefix should fail")
		}
	})

	t.Run("accepts path with h indicator", func(t *testing.T) {
		key1, err1 := PrivateKeyFromMnemonicWithPath(mnemonic, "m/44'/60'/0'")
		key2, err2 := PrivateKeyFromMnemonicWithPath(mnemonic, "m/44h/60h/0h")
		if err1 != nil || err2 != nil {
			t.Fatalf("both paths should work: err1=%v, err2=%v", err1, err2)
		}
		if !key1.Equal(key2) {
			t.Error("' and h hardened indicators should produce same key")
		}
	})

	t.Run("accepts empty path after m/", func(t *testing.T) {
		_, err := PrivateKeyFromMnemonicWithPath(mnemonic, "m/")
		if err != nil {
			t.Errorf("m/ path should work: %v", err)
		}
	})
}

// =============================================================================
// PrivateKeyFromMnemonicWithPassphrase Tests
// =============================================================================

func TestPrivateKeyFromMnemonicWithPassphrase(t *testing.T) {
	mnemonic, _ := GenerateMnemonic(12)

	t.Run("empty passphrase equals no passphrase", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonic(mnemonic)
		key2, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "")
		if !key1.Equal(key2) {
			t.Error("empty passphrase should produce same key as no passphrase")
		}
	})

	t.Run("passphrase changes derived key", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonic(mnemonic)
		key2, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "secret")
		if key1.Equal(key2) {
			t.Error("passphrase should produce different key")
		}
	})

	t.Run("different passphrases produce different keys", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "secret1")
		key2, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "secret2")
		if key1.Equal(key2) {
			t.Error("different passphrases should produce different keys")
		}
	})

	t.Run("passphrase is case sensitive", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "Secret")
		key2, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "secret")
		if key1.Equal(key2) {
			t.Error("passphrase should be case sensitive")
		}
	})

	t.Run("passphrase derivation is deterministic", func(t *testing.T) {
		key1, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "secret")
		key2, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "secret")
		if !key1.Equal(key2) {
			t.Error("same passphrase should produce same key")
		}
	})
}

// =============================================================================
// PrivateKeyFromMnemonicWithOptions Tests
// =============================================================================

func TestPrivateKeyFromMnemonicWithOptions(t *testing.T) {
	mnemonic, _ := GenerateMnemonic(12)

	t.Run("all options combined", func(t *testing.T) {
		key, err := PrivateKeyFromMnemonicWithOptions(mnemonic, "secret", "m/44'/60'/0'/0/1")
		if err != nil {
			t.Fatalf("PrivateKeyFromMnemonicWithOptions failed: %v", err)
		}
		if key.IsZero() {
			t.Error("derived key should not be zero")
		}
	})

	t.Run("consistent with individual functions", func(t *testing.T) {
		// Test consistency with path function
		keyPath, _ := PrivateKeyFromMnemonicWithPath(mnemonic, "m/44'/60'/0'/0/0")
		keyOpts, _ := PrivateKeyFromMnemonicWithOptions(mnemonic, "", "m/44'/60'/0'/0/0")
		if !keyPath.Equal(keyOpts) {
			t.Error("WithPath and WithOptions should produce same key for same path")
		}

		// Test consistency with passphrase function
		keyPass, _ := PrivateKeyFromMnemonicWithPassphrase(mnemonic, "secret")
		keyOpts2, _ := PrivateKeyFromMnemonicWithOptions(mnemonic, "secret", DefaultDerivationPath)
		if !keyPass.Equal(keyOpts2) {
			t.Error("WithPassphrase and WithOptions should produce same key")
		}
	})
}

// =============================================================================
// DefaultDerivationPath Tests
// =============================================================================

func TestDefaultDerivationPath(t *testing.T) {
	t.Run("default path is Ethereum standard", func(t *testing.T) {
		expected := "m/44'/60'/0'/0/0"
		if DefaultDerivationPath != expected {
			t.Errorf("DefaultDerivationPath should be %s, got %s", expected, DefaultDerivationPath)
		}
	})
}

// =============================================================================
// ValidMnemonicWordCounts Tests
// =============================================================================

func TestValidMnemonicWordCounts(t *testing.T) {
	expected := []int{12, 15, 18, 21, 24}

	t.Run("contains expected values", func(t *testing.T) {
		if len(ValidMnemonicWordCounts) != len(expected) {
			t.Errorf("expected %d valid word counts, got %d", len(expected), len(ValidMnemonicWordCounts))
		}
		for i, v := range expected {
			if ValidMnemonicWordCounts[i] != v {
				t.Errorf("expected ValidMnemonicWordCounts[%d] = %d, got %d", i, v, ValidMnemonicWordCounts[i])
			}
		}
	})
}
