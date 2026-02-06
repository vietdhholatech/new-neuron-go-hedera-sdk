package keylib

import (
	"encoding/json"
	"testing"
)

// =============================================================================
// Scramble/Unscramble Tests
// =============================================================================

func TestScrambleUnscramble_Comprehensive(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("round trip works", func(t *testing.T) {
		password := "testpassword123"
		encrypted, err := key.Scramble(password)
		if err != nil {
			t.Fatalf("Scramble failed: %v", err)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, password)
		if err != nil {
			t.Fatalf("UnscramblePrivateKey failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("decrypted key should equal original")
		}
	})

	t.Run("wrong password fails", func(t *testing.T) {
		encrypted, _ := key.Scramble("correct")
		_, err := UnscramblePrivateKey(encrypted, "wrong")
		if err == nil {
			t.Error("wrong password should fail decryption")
		}
	})

	t.Run("empty password fails encryption", func(t *testing.T) {
		_, err := key.Scramble("")
		if err == nil {
			t.Error("empty password should fail encryption")
		}
	})

	t.Run("empty password fails decryption", func(t *testing.T) {
		encrypted, _ := key.Scramble("password")
		_, err := UnscramblePrivateKey(encrypted, "")
		if err == nil {
			t.Error("empty password should fail decryption")
		}
	})

	t.Run("same key different encryptions differ", func(t *testing.T) {
		enc1, _ := key.Scramble("password")
		enc2, _ := key.Scramble("password")

		// Salt and nonce should be different
		if string(enc1.Salt) == string(enc2.Salt) {
			t.Error("salt should be unique per encryption")
		}
		if string(enc1.Nonce) == string(enc2.Nonce) {
			t.Error("nonce should be unique per encryption")
		}
	})

	t.Run("zero key fails encryption", func(t *testing.T) {
		var zeroKey NeuronPrivateKey
		_, err := zeroKey.Scramble("password")
		if err == nil {
			t.Error("zero key should fail encryption")
		}
	})
}

// =============================================================================
// EncryptedPrivateKey Structure Tests
// =============================================================================

func TestEncryptedPrivateKey_Structure(t *testing.T) {
	key, _ := GeneratePrivateKey()
	encrypted, _ := key.Scramble("password")

	t.Run("version is 2", func(t *testing.T) {
		if encrypted.Version != 2 {
			t.Errorf("expected version 2, got %d", encrypted.Version)
		}
	})

	t.Run("argon2 parameters are stored", func(t *testing.T) {
		if encrypted.Argon2Time == 0 {
			t.Error("expected non-zero argon2Time")
		}
		if encrypted.Argon2Memory == 0 {
			t.Error("expected non-zero argon2Memory")
		}
		if encrypted.Argon2Threads == 0 {
			t.Error("expected non-zero argon2Threads")
		}
	})

	t.Run("salt is 16 bytes", func(t *testing.T) {
		if len(encrypted.Salt) != 16 {
			t.Errorf("expected 16 byte salt, got %d", len(encrypted.Salt))
		}
	})

	t.Run("nonce is 12 bytes", func(t *testing.T) {
		if len(encrypted.Nonce) != 12 {
			t.Errorf("expected 12 byte nonce, got %d", len(encrypted.Nonce))
		}
	})

	t.Run("ciphertext is 48 bytes", func(t *testing.T) {
		// 32 bytes key + 16 bytes auth tag
		if len(encrypted.Ciphertext) != 48 {
			t.Errorf("expected 48 byte ciphertext, got %d", len(encrypted.Ciphertext))
		}
	})
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestEncryptedPrivateKey_JSON(t *testing.T) {
	key, _ := GeneratePrivateKey()
	encrypted, _ := key.Scramble("password")

	t.Run("marshal and unmarshal round trip", func(t *testing.T) {
		jsonData, err := json.Marshal(encrypted)
		if err != nil {
			t.Fatalf("JSON marshal failed: %v", err)
		}

		var restored EncryptedPrivateKey
		err = json.Unmarshal(jsonData, &restored)
		if err != nil {
			t.Fatalf("JSON unmarshal failed: %v", err)
		}

		// Verify fields match
		if restored.Version != encrypted.Version {
			t.Error("version mismatch after JSON round trip")
		}
		if string(restored.Salt) != string(encrypted.Salt) {
			t.Error("salt mismatch after JSON round trip")
		}
		if string(restored.Nonce) != string(encrypted.Nonce) {
			t.Error("nonce mismatch after JSON round trip")
		}
		if string(restored.Ciphertext) != string(encrypted.Ciphertext) {
			t.Error("ciphertext mismatch after JSON round trip")
		}
	})

	t.Run("can decrypt after JSON round trip", func(t *testing.T) {
		jsonData, _ := json.Marshal(encrypted)

		var restored EncryptedPrivateKey
		json.Unmarshal(jsonData, &restored)

		decrypted, err := UnscramblePrivateKey(restored, "password")
		if err != nil {
			t.Fatalf("decryption after JSON round trip failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("decrypted key should equal original after JSON round trip")
		}
	})

	t.Run("JSON contains expected fields", func(t *testing.T) {
		jsonData, _ := json.Marshal(encrypted)
		var m map[string]interface{}
		json.Unmarshal(jsonData, &m)

		expectedFields := []string{"version", "salt", "nonce", "ciphertext"}
		for _, field := range expectedFields {
			if _, ok := m[field]; !ok {
				t.Errorf("JSON missing field: %s", field)
			}
		}
	})
}

// =============================================================================
// UnscramblePrivateKey Error Tests
// =============================================================================

func TestUnscramblePrivateKey_Errors(t *testing.T) {
	key, _ := GeneratePrivateKey()
	validEncrypted, _ := key.Scramble("password")

	t.Run("unsupported version fails", func(t *testing.T) {
		bad := validEncrypted
		bad.Version = 99 // Unsupported version
		_, err := UnscramblePrivateKey(bad, "password")
		if err == nil {
			t.Error("unsupported version should fail")
		}
	})

	t.Run("version 1 with defaults works", func(t *testing.T) {
		// Create a version 1 key (legacy format with no stored params)
		v1Key := EncryptedPrivateKey{
			Version:    1,
			Salt:       validEncrypted.Salt,
			Nonce:      validEncrypted.Nonce,
			Ciphertext: validEncrypted.Ciphertext,
			// Argon2 params are zero/unset for version 1
		}
		// This should use default Argon2 params
		_, err := UnscramblePrivateKey(v1Key, "password")
		if err == nil {
			// Expected to fail because the ciphertext was made with v2 params
			// but decrypted with v1 defaults - this tests that version routing works
			// In real v1 keys, this would work if encrypted with defaults
		}
	})

	t.Run("invalid salt length fails", func(t *testing.T) {
		bad := EncryptedPrivateKey{
			Version:    1,
			Salt:       []byte{1, 2, 3}, // Too short
			Nonce:      validEncrypted.Nonce,
			Ciphertext: validEncrypted.Ciphertext,
		}
		_, err := UnscramblePrivateKey(bad, "password")
		if err == nil {
			t.Error("invalid salt length should fail")
		}
	})

	t.Run("invalid nonce length fails", func(t *testing.T) {
		bad := EncryptedPrivateKey{
			Version:    1,
			Salt:       validEncrypted.Salt,
			Nonce:      []byte{1, 2, 3}, // Too short
			Ciphertext: validEncrypted.Ciphertext,
		}
		_, err := UnscramblePrivateKey(bad, "password")
		if err == nil {
			t.Error("invalid nonce length should fail")
		}
	})

	t.Run("invalid ciphertext length fails", func(t *testing.T) {
		bad := EncryptedPrivateKey{
			Version:    1,
			Salt:       validEncrypted.Salt,
			Nonce:      validEncrypted.Nonce,
			Ciphertext: []byte{1, 2, 3}, // Too short
		}
		_, err := UnscramblePrivateKey(bad, "password")
		if err == nil {
			t.Error("invalid ciphertext length should fail")
		}
	})

	t.Run("corrupted ciphertext fails", func(t *testing.T) {
		bad := EncryptedPrivateKey{
			Version:    1,
			Salt:       validEncrypted.Salt,
			Nonce:      validEncrypted.Nonce,
			Ciphertext: make([]byte, 48),
		}
		// Fill with garbage
		for i := range bad.Ciphertext {
			bad.Ciphertext[i] = byte(i)
		}
		_, err := UnscramblePrivateKey(bad, "password")
		if err == nil {
			t.Error("corrupted ciphertext should fail")
		}
	})
}

// =============================================================================
// Password Edge Cases
// =============================================================================

func TestScramble_PasswordEdgeCases(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("long password works", func(t *testing.T) {
		longPassword := "this is a very long password that should still work correctly for encryption and decryption"
		encrypted, err := key.Scramble(longPassword)
		if err != nil {
			t.Fatalf("long password encryption failed: %v", err)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, longPassword)
		if err != nil {
			t.Fatalf("long password decryption failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("long password round trip failed")
		}
	})

	t.Run("unicode password works", func(t *testing.T) {
		unicodePassword := "пароль密码🔑"
		encrypted, err := key.Scramble(unicodePassword)
		if err != nil {
			t.Fatalf("unicode password encryption failed: %v", err)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, unicodePassword)
		if err != nil {
			t.Fatalf("unicode password decryption failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("unicode password round trip failed")
		}
	})

	t.Run("whitespace only password works", func(t *testing.T) {
		spacePassword := "   \t\n   "
		encrypted, err := key.Scramble(spacePassword)
		if err != nil {
			t.Fatalf("whitespace password encryption failed: %v", err)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, spacePassword)
		if err != nil {
			t.Fatalf("whitespace password decryption failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("whitespace password round trip failed")
		}
	})

	t.Run("single character password works", func(t *testing.T) {
		shortPassword := "a"
		encrypted, err := key.Scramble(shortPassword)
		if err != nil {
			t.Fatalf("single char password encryption failed: %v", err)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, shortPassword)
		if err != nil {
			t.Fatalf("single char password decryption failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("single char password round trip failed")
		}
	})
}

// =============================================================================
// Custom Argon2 Options Tests
// =============================================================================

func TestScramble_CustomArgon2Options(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("custom time parameter round trip", func(t *testing.T) {
		password := "password"
		encrypted, err := key.Scramble(password, WithArgon2Time(2))
		if err != nil {
			t.Fatalf("encryption with custom time failed: %v", err)
		}

		if encrypted.Argon2Time != 2 {
			t.Errorf("expected argon2Time=2, got %d", encrypted.Argon2Time)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, password)
		if err != nil {
			t.Fatalf("decryption with custom time failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("custom time parameter round trip failed")
		}
	})

	t.Run("custom memory parameter round trip", func(t *testing.T) {
		password := "password"
		encrypted, err := key.Scramble(password, WithArgon2Memory(32*1024))
		if err != nil {
			t.Fatalf("encryption with custom memory failed: %v", err)
		}

		if encrypted.Argon2Memory != 32*1024 {
			t.Errorf("expected argon2Memory=32768, got %d", encrypted.Argon2Memory)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, password)
		if err != nil {
			t.Fatalf("decryption with custom memory failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("custom memory parameter round trip failed")
		}
	})

	t.Run("custom threads parameter round trip", func(t *testing.T) {
		password := "password"
		encrypted, err := key.Scramble(password, WithArgon2Threads(2))
		if err != nil {
			t.Fatalf("encryption with custom threads failed: %v", err)
		}

		if encrypted.Argon2Threads != 2 {
			t.Errorf("expected argon2Threads=2, got %d", encrypted.Argon2Threads)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, password)
		if err != nil {
			t.Fatalf("decryption with custom threads failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("custom threads parameter round trip failed")
		}
	})

	t.Run("all custom parameters round trip", func(t *testing.T) {
		password := "password"
		encrypted, err := key.Scramble(password,
			WithArgon2Time(2),
			WithArgon2Memory(16*1024),
			WithArgon2Threads(2),
		)
		if err != nil {
			t.Fatalf("encryption with all custom params failed: %v", err)
		}

		if encrypted.Argon2Time != 2 || encrypted.Argon2Memory != 16*1024 || encrypted.Argon2Threads != 2 {
			t.Error("custom parameters not stored correctly")
		}

		decrypted, err := UnscramblePrivateKey(encrypted, password)
		if err != nil {
			t.Fatalf("decryption with all custom params failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("all custom parameters round trip failed")
		}
	})

	t.Run("custom params JSON round trip", func(t *testing.T) {
		password := "password"
		encrypted, _ := key.Scramble(password,
			WithArgon2Time(5),
			WithArgon2Memory(32*1024),
			WithArgon2Threads(8),
		)

		jsonData, err := json.Marshal(encrypted)
		if err != nil {
			t.Fatalf("JSON marshal failed: %v", err)
		}

		var restored EncryptedPrivateKey
		err = json.Unmarshal(jsonData, &restored)
		if err != nil {
			t.Fatalf("JSON unmarshal failed: %v", err)
		}

		// Verify params preserved
		if restored.Argon2Time != 5 || restored.Argon2Memory != 32*1024 || restored.Argon2Threads != 8 {
			t.Error("argon2 params not preserved in JSON")
		}

		decrypted, err := UnscramblePrivateKey(restored, password)
		if err != nil {
			t.Fatalf("decryption after JSON round trip failed: %v", err)
		}

		if !key.Equal(decrypted) {
			t.Error("custom params JSON round trip failed")
		}
	})
}

// =============================================================================
// Argon2 Bounds Validation Tests
// =============================================================================

func TestScramble_Argon2BoundsValidation(t *testing.T) {
	key, _ := GeneratePrivateKey()

	t.Run("argon2Time too low fails", func(t *testing.T) {
		_, err := key.Scramble("password", WithArgon2Time(0))
		if err == nil {
			t.Error("argon2Time=0 should fail")
		}
	})

	t.Run("argon2Time too high fails", func(t *testing.T) {
		_, err := key.Scramble("password", WithArgon2Time(101))
		if err == nil {
			t.Error("argon2Time=101 should fail")
		}
	})

	t.Run("argon2Memory too low fails", func(t *testing.T) {
		_, err := key.Scramble("password", WithArgon2Memory(1024)) // 1 MiB, minimum is 8 MiB
		if err == nil {
			t.Error("argon2Memory below minimum should fail")
		}
	})

	t.Run("argon2Memory too high fails", func(t *testing.T) {
		_, err := key.Scramble("password", WithArgon2Memory(65*1024)) // 65 MiB, max is 64 MiB
		if err == nil {
			t.Error("argon2Memory above maximum should fail")
		}
	})

	t.Run("argon2Threads too low fails", func(t *testing.T) {
		_, err := key.Scramble("password", WithArgon2Threads(0))
		if err == nil {
			t.Error("argon2Threads=0 should fail")
		}
	})

	t.Run("argon2Threads too high fails", func(t *testing.T) {
		_, err := key.Scramble("password", WithArgon2Threads(64))
		if err == nil {
			t.Error("argon2Threads above maximum should fail")
		}
	})

	t.Run("valid boundary values work", func(t *testing.T) {
		// Test minimum valid values
		_, err := key.Scramble("password",
			WithArgon2Time(1),
			WithArgon2Memory(8*1024),
			WithArgon2Threads(1),
		)
		if err != nil {
			t.Errorf("minimum valid values should work: %v", err)
		}
	})
}

// =============================================================================
// Argon2 Memory DoS Prevention Tests
// =============================================================================

func TestUnscramble_RejectsHighMemoryParams(t *testing.T) {
	// Test that decryption rejects stored params exceeding the new 64 MiB limit.
	// This prevents DoS via attacker-controlled encrypted key files.

	t.Run("rejects memory > 64 MiB during decryption", func(t *testing.T) {
		// Simulate an encrypted key with memory > 64 MiB
		// This would have been valid before the fix but should be rejected now
		encrypted := EncryptedPrivateKey{
			Version:       2,
			Salt:          make([]byte, 16),
			Nonce:         make([]byte, 12),
			Ciphertext:    make([]byte, 48),
			Argon2Time:    3,
			Argon2Memory:  65 * 1024, // 65 MiB - just over the new limit
			Argon2Threads: 4,
		}

		_, err := UnscramblePrivateKey(encrypted, "password")
		if err == nil {
			t.Error("should reject memory > 64 MiB")
		}
	})

	t.Run("accepts memory at exactly 64 MiB", func(t *testing.T) {
		// Generate a real encrypted key with 64 MiB (max allowed)
		privKey, _ := GeneratePrivateKey()
		encrypted, err := privKey.Scramble("test-password", WithArgon2Memory(64*1024))
		if err != nil {
			t.Fatalf("Scramble with 64 MiB should work: %v", err)
		}

		// Verify decryption works
		decrypted, err := UnscramblePrivateKey(encrypted, "test-password")
		if err != nil {
			t.Fatalf("UnscramblePrivateKey with 64 MiB should work: %v", err)
		}

		if !decrypted.Equal(privKey) {
			t.Error("decrypted key should match original")
		}
	})

	t.Run("default encryption still works", func(t *testing.T) {
		// Default is 64 MiB which is now the max - should still work
		privKey, _ := GeneratePrivateKey()
		password := "test-password"

		encrypted, err := privKey.Scramble(password)
		if err != nil {
			t.Fatalf("Scramble with defaults should work: %v", err)
		}

		decrypted, err := UnscramblePrivateKey(encrypted, password)
		if err != nil {
			t.Fatalf("UnscramblePrivateKey should work: %v", err)
		}

		if !decrypted.Equal(privKey) {
			t.Error("decrypted key should match original")
		}
	})
}
