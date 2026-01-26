package account

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// JSON Serialization Edge Cases - Phase 2
// These tests verify JSON marshaling/unmarshaling behavior and edge cases.
//
// NOTE: NeuronAccount currently only implements MarshalJSON. UnmarshalJSON
// is not implemented, so accounts must be created via builders.
// These tests document expected behavior and verify marshaling correctness.
// =============================================================================

// =============================================================================
// MarshalJSON Tests
// =============================================================================

func TestNeuronAccount_MarshalJSON_ParentAccount(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (NeuronAccount, error)
		checkJSON func(t *testing.T, data []byte)
	}{
		{
			name: "minimal parent account",
			setup: func() (NeuronAccount, error) {
				privKey, _ := keylib.GeneratePrivateKey()
				pubKey := privKey.PublicKey()
				did := newConcurrentMockDID(pubKey)
				return NewParentAccountBuilder(pubKey, did).Build()
			},
			checkJSON: func(t *testing.T, data []byte) {
				var m map[string]interface{}
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				// Required fields
				if _, ok := m["publicKey"]; !ok {
					t.Error("missing publicKey field")
				}
				if _, ok := m["peerId"]; !ok {
					t.Error("missing peerId field")
				}
				if _, ok := m["evmAddress"]; !ok {
					t.Error("missing evmAddress field")
				}
				if _, ok := m["accountType"]; !ok {
					t.Error("missing accountType field")
				}
				if m["accountType"] != "Parent" {
					t.Errorf("expected accountType=Parent, got %v", m["accountType"])
				}
			},
		},
		{
			name: "parent account with all optional fields",
			setup: func() (NeuronAccount, error) {
				privKey, _ := keylib.GeneratePrivateKey()
				pubKey := privKey.PublicKey()
				did := newConcurrentMockDID(pubKey)
				peerID, _ := pubKey.PeerID()
				addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()

				return NewParentAccountBuilder(pubKey, did).
					WithStdInHedera("0.0.111").
					WithStdOutHedera("0.0.222").
					WithStdErrHedera("0.0.333").
					WithReachableAddr(addr).
					Build()
			},
			checkJSON: func(t *testing.T, data []byte) {
				var m map[string]interface{}
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				// Check all fields present
				requiredFields := []string{"publicKey", "peerId", "evmAddress", "accountType", "did"}
				for _, field := range requiredFields {
					if _, ok := m[field]; !ok {
						t.Errorf("missing required field: %s", field)
					}
				}
				// Check optional fields
				optionalFields := []string{"stdIn", "stdOut", "stdErr", "reachableAddrs"}
				for _, field := range optionalFields {
					if _, ok := m[field]; !ok {
						t.Errorf("missing optional field that should be present: %s", field)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := tt.setup()
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			data, err := account.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			// Verify it's valid JSON
			if !json.Valid(data) {
				t.Fatal("MarshalJSON produced invalid JSON")
			}

			tt.checkJSON(t, data)
		})
	}
}

func TestNeuronAccount_MarshalJSON_ChildAccount(t *testing.T) {
	t.Run("child account JSON structure", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		parentPubKey := parentPriv.PublicKey()

		childPriv, _ := keylib.GeneratePrivateKey()
		childPubKey := childPriv.PublicKey()

		account, err := NewChildAccountBuilder(childPubKey, parentPubKey).
			WithStdInHedera("0.0.111").
			Build()
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}

		data, err := account.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// Child account should have parentPublicKey, not did
		if m["accountType"] != "Child" {
			t.Errorf("expected accountType=Child, got %v", m["accountType"])
		}
		if _, ok := m["parentPublicKey"]; !ok {
			t.Error("child account missing parentPublicKey field")
		}
		if _, ok := m["did"]; ok {
			t.Error("child account should not have did field")
		}
	})
}

func TestNeuronAccount_MarshalJSON_ZeroValue(t *testing.T) {
	t.Run("zero value account marshals correctly", func(t *testing.T) {
		var account NeuronAccount
		data, err := account.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Should produce valid JSON even for zero value
		if !json.Valid(data) {
			t.Fatal("invalid JSON for zero value account")
		}

		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		// Verify zero-value fields
		if m["publicKey"] != "" {
			t.Errorf("expected empty publicKey for zero value, got %v", m["publicKey"])
		}
	})
}

// =============================================================================
// JSON Field Format Tests
// =============================================================================

func TestNeuronAccount_MarshalJSON_FieldFormats(t *testing.T) {
	privKey, _ := keylib.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	did := newConcurrentMockDID(pubKey)
	peerID, _ := pubKey.PeerID()

	account, err := NewParentAccountBuilder(pubKey, did).
		WithStdInHedera("0.0.12345").
		WithReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	data, err := account.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	t.Run("publicKey is hex format", func(t *testing.T) {
		pk := m["publicKey"].(string)
		if !strings.HasPrefix(pk, "0x") {
			t.Errorf("publicKey should have 0x prefix, got: %s", pk)
		}
		// Should be 66 chars (0x + 64 hex chars for compressed key)
		if len(pk) != 68 {
			t.Errorf("expected publicKey length 68, got %d", len(pk))
		}
	})

	t.Run("peerId is valid format", func(t *testing.T) {
		pid := m["peerId"].(string)
		// PeerID should be base58btc encoded
		if len(pid) < 40 {
			t.Errorf("peerId seems too short: %s", pid)
		}
	})

	t.Run("evmAddress is hex format", func(t *testing.T) {
		addr := m["evmAddress"].(string)
		if !strings.HasPrefix(addr, "0x") {
			t.Errorf("evmAddress should have 0x prefix, got: %s", addr)
		}
		// Should be 42 chars (0x + 40 hex chars)
		if len(addr) != 42 {
			t.Errorf("expected evmAddress length 42, got %d", len(addr))
		}
	})

	t.Run("stdIn is topic format", func(t *testing.T) {
		stdIn := m["stdIn"].(string)
		if !strings.HasPrefix(stdIn, "hedera-topic:") {
			t.Errorf("stdIn should have hedera-topic: prefix, got: %s", stdIn)
		}
	})

	t.Run("reachableAddrs is array of strings", func(t *testing.T) {
		addrs, ok := m["reachableAddrs"].([]interface{})
		if !ok {
			t.Fatal("reachableAddrs should be an array")
		}
		if len(addrs) == 0 {
			t.Error("expected at least one reachable address")
		}
		for i, addr := range addrs {
			if _, ok := addr.(string); !ok {
				t.Errorf("reachableAddrs[%d] should be string", i)
			}
		}
	})
}

// =============================================================================
// JSON Edge Cases
// =============================================================================

func TestNeuronAccount_MarshalJSON_EdgeCases(t *testing.T) {
	t.Run("account with empty reachable addrs", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).Build()
		data, err := account.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var m map[string]interface{}
		json.Unmarshal(data, &m)

		// reachableAddrs should be omitted or null for empty
		if addrs, ok := m["reachableAddrs"]; ok && addrs != nil {
			if arr, ok := addrs.([]interface{}); ok && len(arr) > 0 {
				t.Error("expected no reachableAddrs for account without addresses")
			}
		}
	})

	t.Run("account with multiple reachable addrs", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)
		peerID, _ := pubKey.PeerID()

		addr1 := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()
		addr2 := "/ip4/192.168.1.2/tcp/4001/p2p/" + peerID.String()
		addr3 := "/ip6/::1/tcp/4001/p2p/" + peerID.String()

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithReachableAddr(addr1).
			WithReachableAddr(addr2).
			WithReachableAddr(addr3).
			Build()

		data, _ := account.MarshalJSON()
		var m map[string]interface{}
		json.Unmarshal(data, &m)

		addrs := m["reachableAddrs"].([]interface{})
		if len(addrs) != 3 {
			t.Errorf("expected 3 reachable addresses, got %d", len(addrs))
		}
	})

	t.Run("marshal produces consistent output", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("0.0.111").
			Build()

		// Marshal multiple times
		data1, _ := account.MarshalJSON()
		data2, _ := account.MarshalJSON()
		data3, _ := account.MarshalJSON()

		if string(data1) != string(data2) || string(data2) != string(data3) {
			t.Error("MarshalJSON should produce consistent output")
		}
	})
}

// =============================================================================
// JSON String Representation Tests
// =============================================================================

func TestNeuronAccount_String_Representations(t *testing.T) {
	t.Run("zero value string", func(t *testing.T) {
		var account NeuronAccount
		str := account.String()
		if str != "NeuronAccount{zero-value}" {
			t.Errorf("unexpected zero value string: %s", str)
		}
	})

	t.Run("parent account with DID string", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).Build()
		str := account.String()

		if !strings.Contains(str, "Parent") {
			t.Errorf("expected 'Parent' in string: %s", str)
		}
		if !strings.Contains(str, "did=") {
			t.Errorf("expected 'did=' in string: %s", str)
		}
	})

	t.Run("child account string", func(t *testing.T) {
		parentPriv, _ := keylib.GeneratePrivateKey()
		childPriv, _ := keylib.GeneratePrivateKey()

		account, _ := NewChildAccountBuilder(childPriv.PublicKey(), parentPriv.PublicKey()).Build()
		str := account.String()

		if !strings.Contains(str, "Child") {
			t.Errorf("expected 'Child' in string: %s", str)
		}
		if !strings.Contains(str, "pubKey=") {
			t.Errorf("expected 'pubKey=' in string: %s", str)
		}
	})
}

// =============================================================================
// CommAddress JSON Tests
// =============================================================================

func TestCommAddress_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		addr     CommAddress
		expected string
	}{
		{
			name:     "hedera topic address",
			addr:     func() CommAddress { a, _ := NewHederaTopicAddress("0.0.12345"); return a }(),
			expected: "hedera-topic:0.0.12345",
		},
		{
			name:     "kafka topic address",
			addr:     func() CommAddress { a, _ := NewKafkaTopicAddress("my-topic"); return a }(),
			expected: "kafka-topic:my-topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.addr.String()
			if str != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, str)
			}
		})
	}
}

// =============================================================================
// CommAddressSet JSON Tests
// =============================================================================

func TestCommAddressSet_JSONLikeOperations(t *testing.T) {
	t.Run("addresses method returns strings", func(t *testing.T) {
		addr1, _ := NewHederaTopicAddress("0.0.111")
		addr2, _ := NewHederaTopicAddress("0.0.222")
		set := NewCommAddressSet(addr1, addr2)

		addrs := set.Addresses()
		if len(addrs) != 2 {
			t.Errorf("expected 2 addresses, got %d", len(addrs))
		}
	})

	t.Run("empty set returns nil addresses", func(t *testing.T) {
		set := NewCommAddressSet()
		addrs := set.Addresses()
		if addrs != nil && len(addrs) > 0 {
			t.Error("expected nil or empty slice for empty set")
		}
	})
}

// =============================================================================
// ReachableAddrs JSON Tests
// =============================================================================

func TestReachableAddrs_StringsSerialization(t *testing.T) {
	testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

	t.Run("strings method returns valid multiaddr strings", func(t *testing.T) {
		addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID)
		addr2, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4002/p2p/" + testPeerID)

		addrs := NewReachableAddrs(addr1, addr2)
		strs := addrs.Strings()

		if len(strs) != 2 {
			t.Errorf("expected 2 strings, got %d", len(strs))
		}

		for _, s := range strs {
			if !strings.HasPrefix(s, "/ip4/") {
				t.Errorf("expected multiaddr format, got: %s", s)
			}
		}
	})

	t.Run("empty addrs returns nil strings", func(t *testing.T) {
		addrs := NewReachableAddrs()
		strs := addrs.Strings()
		if strs != nil && len(strs) > 0 {
			t.Error("expected nil or empty slice for empty addrs")
		}
	})
}

// =============================================================================
// JSON-Compatible Data Structure Tests
// =============================================================================

func TestNeuronAccount_JSONCompatibleStructure(t *testing.T) {
	// This test verifies the JSON structure matches expected external format
	t.Run("JSON structure matches spec", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)
		peerID, _ := pubKey.PeerID()

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("0.0.111").
			WithStdOutHedera("0.0.222").
			WithStdErrHedera("0.0.333").
			WithReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()).
			Build()

		data, _ := account.MarshalJSON()

		// Parse into neuronAccountJSON structure (same as internal)
		type expectedJSON struct {
			PublicKey      string   `json:"publicKey"`
			PeerID         string   `json:"peerId"`
			EVMAddress     string   `json:"evmAddress"`
			AccountType    string   `json:"accountType"`
			DID            string   `json:"did,omitempty"`
			ParentPubKey   string   `json:"parentPublicKey,omitempty"`
			StdIn          string   `json:"stdIn,omitempty"`
			StdOut         string   `json:"stdOut,omitempty"`
			StdErr         string   `json:"stdErr,omitempty"`
			ReachableAddrs []string `json:"reachableAddrs,omitempty"`
		}

		var parsed expectedJSON
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to parse into expected structure: %v", err)
		}

		// Verify field names match spec
		if parsed.PublicKey == "" {
			t.Error("publicKey field missing or empty")
		}
		if parsed.PeerID == "" {
			t.Error("peerId field missing or empty")
		}
		if parsed.EVMAddress == "" {
			t.Error("evmAddress field missing or empty")
		}
		if parsed.AccountType != "Parent" {
			t.Errorf("expected accountType=Parent, got %s", parsed.AccountType)
		}
		if parsed.StdIn == "" {
			t.Error("stdIn field missing or empty")
		}
	})
}

// =============================================================================
// Large Data Serialization Tests
// =============================================================================

func TestNeuronAccount_LargeReachableAddrs_JSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large data test in short mode")
	}

	t.Run("many reachable addresses serialize correctly", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)
		peerID, _ := pubKey.PeerID()

		builder := NewParentAccountBuilder(pubKey, did)

		// Add 100 addresses
		for i := 0; i < 100; i++ {
			addr := "/ip4/192.168.1." + string(rune('0'+i%10)) + "/tcp/4001/p2p/" + peerID.String()
			builder.WithReachableAddr(addr)
		}

		account, err := builder.Build()
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}

		data, err := account.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Verify it's valid and parseable
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
	})
}
