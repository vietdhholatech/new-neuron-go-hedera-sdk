package keylib

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// =============================================================================
// ParseEVMAddress Tests
// =============================================================================

func TestParseEVMAddress_Comprehensive(t *testing.T) {
	t.Run("parses lowercase address", func(t *testing.T) {
		addr, err := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")
		if err != nil {
			t.Fatalf("ParseEVMAddress failed: %v", err)
		}
		if addr.IsZero() {
			t.Error("parsed address should not be zero")
		}
	})

	t.Run("parses uppercase address", func(t *testing.T) {
		addr, err := ParseEVMAddress("0xE364F2F1E5F4F03D1DF682322500B9C68C997EC3")
		if err != nil {
			t.Fatalf("ParseEVMAddress failed: %v", err)
		}
		if addr.IsZero() {
			t.Error("parsed address should not be zero")
		}
	})

	t.Run("parses checksum address", func(t *testing.T) {
		addr, err := ParseEVMAddress("0xe364f2f1e5F4F03d1df682322500b9c68C997ec3")
		if err != nil {
			t.Fatalf("ParseEVMAddress failed: %v", err)
		}
		if addr.IsZero() {
			t.Error("parsed address should not be zero")
		}
	})

	t.Run("parses without 0x prefix", func(t *testing.T) {
		addr, err := ParseEVMAddress("e364f2f1e5f4f03d1df682322500b9c68c997ec3")
		if err != nil {
			t.Fatalf("ParseEVMAddress failed: %v", err)
		}
		if addr.IsZero() {
			t.Error("parsed address should not be zero")
		}
	})

	t.Run("parses zero address", func(t *testing.T) {
		addr, err := ParseEVMAddress("0x0000000000000000000000000000000000000000")
		if err != nil {
			t.Fatalf("ParseEVMAddress failed: %v", err)
		}
		if !addr.IsZero() {
			t.Error("zero address should be zero")
		}
	})

	t.Run("rejects too short", func(t *testing.T) {
		_, err := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec")
		if err == nil {
			t.Error("too short address should fail")
		}
	})

	t.Run("rejects too long", func(t *testing.T) {
		_, err := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3a")
		if err == nil {
			t.Error("too long address should fail")
		}
	})

	t.Run("rejects invalid hex", func(t *testing.T) {
		_, err := ParseEVMAddress("0xg364f2f1e5f4f03d1df682322500b9c68c997ec3")
		if err == nil {
			t.Error("invalid hex should fail")
		}
	})

	t.Run("rejects empty", func(t *testing.T) {
		_, err := ParseEVMAddress("")
		if err == nil {
			t.Error("empty string should fail")
		}
	})
}

// =============================================================================
// EVMAddress.IsZero Tests
// =============================================================================

func TestEVMAddress_IsZero(t *testing.T) {
	t.Run("zero value is zero", func(t *testing.T) {
		var addr EVMAddress
		if !addr.IsZero() {
			t.Error("zero value should be zero")
		}
	})

	t.Run("ZeroEVMAddress is zero", func(t *testing.T) {
		if !ZeroEVMAddress.IsZero() {
			t.Error("ZeroEVMAddress should be zero")
		}
	})

	t.Run("non-zero address is not zero", func(t *testing.T) {
		addr, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")
		if addr.IsZero() {
			t.Error("non-zero address should not be zero")
		}
	})
}

// =============================================================================
// EVMAddress.Bytes Tests
// =============================================================================

func TestEVMAddress_Bytes(t *testing.T) {
	addr, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")

	t.Run("returns 20 bytes", func(t *testing.T) {
		bytes := addr.Bytes()
		if len(bytes) != 20 {
			t.Errorf("expected 20 bytes, got %d", len(bytes))
		}
	})

	t.Run("correct first byte", func(t *testing.T) {
		bytes := addr.Bytes()
		if bytes[0] != 0xe3 {
			t.Errorf("expected first byte 0xe3, got 0x%02x", bytes[0])
		}
	})

	t.Run("zero address returns zero bytes", func(t *testing.T) {
		bytes := ZeroEVMAddress.Bytes()
		for i, b := range bytes {
			if b != 0 {
				t.Errorf("zero address byte %d should be 0, got %d", i, b)
			}
		}
	})
}

// =============================================================================
// EVMAddress.Hex Tests
// =============================================================================

func TestEVMAddress_Hex(t *testing.T) {
	addr, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")

	t.Run("includes 0x prefix", func(t *testing.T) {
		hex := addr.Hex()
		if !strings.HasPrefix(hex, "0x") {
			t.Error("hex should start with 0x")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		hex := addr.Hex()
		if len(hex) != 42 {
			t.Errorf("expected 42 chars, got %d", len(hex))
		}
	})

	t.Run("lowercase", func(t *testing.T) {
		hex := addr.Hex()
		if strings.ToLower(hex) != hex {
			t.Error("hex should be lowercase")
		}
	})

	t.Run("round trip", func(t *testing.T) {
		hex := addr.Hex()
		restored, _ := ParseEVMAddress(hex)
		if !addr.Equal(restored) {
			t.Error("address should equal restored")
		}
	})
}

// =============================================================================
// EVMAddress.ChecksumHex Tests
// =============================================================================

func TestEVMAddress_ChecksumHex(t *testing.T) {
	addr, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")

	t.Run("includes 0x prefix", func(t *testing.T) {
		hex := addr.ChecksumHex()
		if !strings.HasPrefix(hex, "0x") {
			t.Error("checksum hex should start with 0x")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		hex := addr.ChecksumHex()
		if len(hex) != 42 {
			t.Errorf("expected 42 chars, got %d", len(hex))
		}
	})

	t.Run("has mixed case (EIP-55)", func(t *testing.T) {
		hex := addr.ChecksumHex()
		lower := strings.ToLower(hex)
		upper := strings.ToUpper(hex)
		// Checksum should have mixed case (unless address is all 0-9)
		if hex == lower || hex == upper {
			// Only skip if address has no letters
			hasLetter := false
			for _, c := range hex[2:] {
				if (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
					hasLetter = true
					break
				}
			}
			if hasLetter {
				t.Error("checksum hex should have mixed case")
			}
		}
	})

	t.Run("round trip preserves checksum", func(t *testing.T) {
		hex := addr.ChecksumHex()
		restored, _ := ParseEVMAddress(hex)
		if restored.ChecksumHex() != hex {
			t.Error("checksum should be preserved after round trip")
		}
	})
}

// =============================================================================
// EVMAddress.String Tests
// =============================================================================

func TestEVMAddress_String(t *testing.T) {
	addr, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")

	t.Run("String equals ChecksumHex", func(t *testing.T) {
		if addr.String() != addr.ChecksumHex() {
			t.Error("String should equal ChecksumHex")
		}
	})
}

// =============================================================================
// EVMAddress.Equal Tests
// =============================================================================

func TestEVMAddress_Equal(t *testing.T) {
	addr1, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")
	addr2, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")
	addr3, _ := ParseEVMAddress("0x1234567890123456789012345678901234567890")

	t.Run("same address equals", func(t *testing.T) {
		if !addr1.Equal(addr2) {
			t.Error("same address should equal")
		}
	})

	t.Run("different addresses not equal", func(t *testing.T) {
		if addr1.Equal(addr3) {
			t.Error("different addresses should not equal")
		}
	})

	t.Run("zero addresses equal", func(t *testing.T) {
		var z1, z2 EVMAddress
		if !z1.Equal(z2) {
			t.Error("zero addresses should equal")
		}
	})

	t.Run("case insensitive comparison", func(t *testing.T) {
		addrLower, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")
		addrUpper, _ := ParseEVMAddress("0xE364F2F1E5F4F03D1DF682322500B9C68C997EC3")
		if !addrLower.Equal(addrUpper) {
			t.Error("same address with different case should equal")
		}
	})
}

// =============================================================================
// EVMAddress Interop Tests
// =============================================================================

func TestEVMAddress_Interop(t *testing.T) {
	addr, _ := ParseEVMAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")

	t.Run("ToCommon returns correct type", func(t *testing.T) {
		commonAddr := addr.ToCommon()
		var _ common.Address = commonAddr // Type assertion
	})

	t.Run("ToCommon round trip", func(t *testing.T) {
		commonAddr := addr.ToCommon()
		restored := EVMAddressFromCommon(commonAddr)
		if !addr.Equal(restored) {
			t.Error("address should equal after ToCommon round trip")
		}
	})

	t.Run("EVMAddressFromCommon", func(t *testing.T) {
		commonAddr := common.HexToAddress("0xe364f2f1e5f4f03d1df682322500b9c68c997ec3")
		restored := EVMAddressFromCommon(commonAddr)
		if !addr.Equal(restored) {
			t.Error("EVMAddressFromCommon should produce equal address")
		}
	})
}

// =============================================================================
// ZeroEVMAddress Tests
// =============================================================================

func TestZeroEVMAddress(t *testing.T) {
	t.Run("ZeroEVMAddress is zero", func(t *testing.T) {
		if !ZeroEVMAddress.IsZero() {
			t.Error("ZeroEVMAddress should be zero")
		}
	})

	t.Run("ZeroEVMAddress hex", func(t *testing.T) {
		expected := "0x0000000000000000000000000000000000000000"
		if ZeroEVMAddress.Hex() != expected {
			t.Errorf("expected %s, got %s", expected, ZeroEVMAddress.Hex())
		}
	})
}
