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

	t.Run("version is 1", func(t *testing.T) {
		if encrypted.Version != 1 {
			t.Errorf("expected version 1, got %d", encrypted.Version)
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

	t.Run("wrong version fails", func(t *testing.T) {
		bad := validEncrypted
		bad.Version = 2
		_, err := UnscramblePrivateKey(bad, "password")
		if err == nil {
			t.Error("wrong version should fail")
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
