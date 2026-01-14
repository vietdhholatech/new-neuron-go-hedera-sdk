package keylib

import (
	"encoding/json"
	"testing"
)

// Test vectors from existing SDK
const (
	testPublicKeyHex   = "02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
	testExpectedPeerID = "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
	testExpectedEVM    = "0xe364f2f1e5F4F03d1df682322500b9c68C997ec3"
)

func TestPublicKeyToPeerID(t *testing.T) {
	pubKey, err := ParsePublicKeyHex(testPublicKeyHex)
	if err != nil {
		t.Fatalf("ParsePublicKeyHex failed: %v", err)
	}

	peerID, err := pubKey.PeerID()
	if err != nil {
		t.Fatalf("PeerID derivation failed: %v", err)
	}

	if peerID.String() != testExpectedPeerID {
		t.Errorf("PeerID mismatch:\n  got:  %s\n  want: %s", peerID.String(), testExpectedPeerID)
	}
}

func TestPublicKeyToEVMAddress(t *testing.T) {
	pubKey, err := ParsePublicKeyHex(testPublicKeyHex)
	if err != nil {
		t.Fatalf("ParsePublicKeyHex failed: %v", err)
	}

	evmAddr := pubKey.EVMAddress()

	// Compare lowercase hex
	got := evmAddr.Hex()
	want := "0xe364f2f1e5f4f03d1df682322500b9c68c997ec3" // lowercase version

	if got != want {
		t.Errorf("EVMAddress mismatch:\n  got:  %s\n  want: %s", got, want)
	}

	// Also verify against the test vector constant (case-insensitive)
	expectedAddr, err := ParseEVMAddress(testExpectedEVM)
	if err != nil {
		t.Fatalf("ParseEVMAddress failed for test vector: %v", err)
	}
	if !evmAddr.Equal(expectedAddr) {
		t.Errorf("EVMAddress does not match test vector:\n  got:  %s\n  want: %s", evmAddr.Hex(), testExpectedEVM)
	}

	// Also test checksum version
	checksummed := evmAddr.ChecksumHex()
	t.Logf("Checksummed address: %s", checksummed)
}

func TestKeyGeneration(t *testing.T) {
	// Generate a new key
	privKey, err := GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey failed: %v", err)
	}

	if privKey.IsZero() {
		t.Error("Generated key is zero-value")
	}

	// Derive public key
	pubKey := privKey.PublicKey()
	if pubKey.IsZero() {
		t.Error("Derived public key is zero-value")
	}

	// Derive EVM address
	evmAddr := pubKey.EVMAddress()
	if evmAddr.IsZero() {
		t.Error("Derived EVM address is zero")
	}

	// Derive PeerID
	peerID, err := pubKey.PeerID()
	if err != nil {
		t.Fatalf("PeerID derivation failed: %v", err)
	}
	if peerID.IsZero() {
		t.Error("Derived PeerID is zero-value")
	}

	t.Logf("Generated key:\n  Private: %s...\n  Public: %s\n  EVM: %s\n  PeerID: %s",
		privKey.Hex()[:10], pubKey.Hex(), evmAddr.String(), peerID.String())
}

func TestSignAndVerify(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey failed: %v", err)
	}

	message := []byte("Hello, Neuron!")

	// Sign the message
	sig, err := privKey.SignMessage(message)
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}

	// Verify with correct public key
	pubKey := privKey.PublicKey()
	if !pubKey.Verify(message, sig) {
		t.Error("Signature verification failed for valid signature")
	}

	// Verify fails with wrong message
	if pubKey.Verify([]byte("wrong message"), sig) {
		t.Error("Signature verification should fail for wrong message")
	}

	// Verify fails with different key
	otherKey, _ := GeneratePrivateKey()
	if otherKey.PublicKey().Verify(message, sig) {
		t.Error("Signature verification should fail for wrong key")
	}
}

func TestRecoverPublicKey(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey failed: %v", err)
	}

	message := []byte("Recover me!")

	sig, err := privKey.SignMessage(message)
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}

	// Recover public key from signature
	recovered, err := RecoverPublicKey(message, sig)
	if err != nil {
		t.Fatalf("RecoverPublicKey failed: %v", err)
	}

	// Should match original public key
	if !privKey.MatchesPublicKey(recovered) {
		t.Error("Recovered public key doesn't match original")
	}
}

func TestScrambleUnscramble(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey failed: %v", err)
	}

	originalHex := privKey.Hex()
	password := "test-password-123"

	// Encrypt
	encrypted, err := privKey.Scramble(password)
	if err != nil {
		t.Fatalf("Scramble failed: %v", err)
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(encrypted)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	t.Logf("Encrypted JSON: %s", string(jsonData))

	// Deserialize from JSON
	var loaded EncryptedPrivateKey
	if err := json.Unmarshal(jsonData, &loaded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Decrypt
	restored, err := UnscramblePrivateKey(loaded, password)
	if err != nil {
		t.Fatalf("UnscramblePrivateKey failed: %v", err)
	}

	if restored.Hex() != originalHex {
		t.Error("Restored key doesn't match original")
	}

	// Wrong password should fail
	_, err = UnscramblePrivateKey(loaded, "wrong-password")
	if err == nil {
		t.Error("UnscramblePrivateKey should fail with wrong password")
	}
}

func TestParsePrivateKeyHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid 64 chars", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", false},
		{"valid with 0x", "0xa1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", false},
		{"too short", "a1b2c3", true},
		{"too long", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2a1", true},
		{"invalid hex char", "g1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", true},
		{"all zeros", "0000000000000000000000000000000000000000000000000000000000000000", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePrivateKeyHex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePrivateKeyHex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParsePublicKeyHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid compressed", testPublicKeyHex, false},
		{"valid with 0x", "0x" + testPublicKeyHex, false},
		{"wrong length", "02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99", true},
		{"invalid prefix", "05759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePublicKeyHex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePublicKeyHex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseEVMAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "0xe364f2f1e5f4f03d1df682322500b9c68c997ec3", false},
		{"valid uppercase", "0xE364F2F1E5F4F03D1DF682322500B9C68C997EC3", false},
		{"valid mixed (checksum)", "0xe364f2f1e5F4F03d1df682322500b9c68C997ec3", false},
		{"without 0x", "e364f2f1e5f4f03d1df682322500b9c68c997ec3", false},
		{"too short", "0xe364f2f1e5f4f03d1df682322500b9c68c997ec", true},
		{"too long", "0xe364f2f1e5f4f03d1df682322500b9c68c997ec33", true},
		{"invalid hex", "0xg364f2f1e5f4f03d1df682322500b9c68c997ec3", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseEVMAddress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEVMAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParsePeerID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", testExpectedPeerID, false},
		{"invalid", "invalid-peer-id", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePeerID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePeerID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKeyMatching(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	evmAddr := pubKey.EVMAddress()
	peerID, _ := pubKey.PeerID()

	// PrivateKey matches PublicKey
	if !privKey.MatchesPublicKey(pubKey) {
		t.Error("PrivateKey should match its PublicKey")
	}

	// PrivateKey matches EVMAddress
	if !privKey.MatchesEVMAddress(evmAddr) {
		t.Error("PrivateKey should match its EVMAddress")
	}

	// PublicKey matches EVMAddress
	if !pubKey.MatchesEVMAddress(evmAddr) {
		t.Error("PublicKey should match its EVMAddress")
	}

	// PublicKey matches PeerID
	if !pubKey.MatchesPeerID(peerID) {
		t.Error("PublicKey should match its PeerID")
	}

	// Different keys should not match
	otherPriv, _ := GeneratePrivateKey()
	otherPub := otherPriv.PublicKey()

	if privKey.MatchesPublicKey(otherPub) {
		t.Error("Different keys should not match")
	}
}

func TestZeroize(t *testing.T) {
	privKey, _ := GeneratePrivateKey()

	// Verify key is valid before zeroize
	if privKey.IsZero() {
		t.Fatal("Key should not be zero before zeroize")
	}

	// Zeroize
	privKey.Zeroize()

	// Key should now be zero
	if !privKey.IsZero() {
		t.Error("Key should be zero after zeroize")
	}
}

func TestMnemonicGeneration(t *testing.T) {
	// Test all valid word counts
	wordCounts := []int{12, 15, 18, 21, 24}

	for _, wc := range wordCounts {
		t.Run(string(rune('0'+wc/10))+string(rune('0'+wc%10))+" words", func(t *testing.T) {
			mnemonic, err := GenerateMnemonic(wc)
			if err != nil {
				t.Fatalf("GenerateMnemonic(%d) failed: %v", wc, err)
			}

			// Validate the generated mnemonic
			if err := ValidateMnemonic(mnemonic); err != nil {
				t.Fatalf("Generated mnemonic is invalid: %v", err)
			}

			// Derive key from mnemonic
			privKey, err := PrivateKeyFromMnemonic(mnemonic)
			if err != nil {
				t.Fatalf("PrivateKeyFromMnemonic failed: %v", err)
			}

			if privKey.IsZero() {
				t.Error("Derived key is zero-value")
			}
		})
	}

	// Test invalid word count
	_, err := GenerateMnemonic(13)
	if err == nil {
		t.Error("GenerateMnemonic(13) should fail")
	}
}

func TestCryptoSignerInterface(t *testing.T) {
	privKey, _ := GeneratePrivateKey()

	// Test Public() method
	pub := privKey.Public()
	if pub == nil {
		t.Error("Public() returned nil")
	}

	// Test Sign() method (crypto.Signer interface)
	digest := make([]byte, 32)
	for i := range digest {
		digest[i] = byte(i)
	}

	sig, err := privKey.Sign(nil, digest, nil)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	if len(sig) != 64 { // R + S, no V
		t.Errorf("Sign() returned wrong length: got %d, want 64", len(sig))
	}
}
