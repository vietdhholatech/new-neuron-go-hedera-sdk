package account

import (
	"errors"
	"sync"
	"testing"

	"github.com/aspect-build/neuron-go-hedera-sdk/keylib"
)

// =============================================================================
// Concurrent Access Tests - Phase 1
// These tests verify thread safety using the race detector: go test -race
//
// DESIGN NOTE: AccountBuilder is NOT thread-safe by design. Each goroutine
// should create and use its own builder instance. The immutable NeuronAccount
// returned by Build() IS safe for concurrent reads.
// =============================================================================

// concurrentMockDID is a thread-safe mock DID for concurrent tests
type concurrentMockDID struct {
	did      string
	matchKey keylib.NeuronPublicKey
	mu       sync.RWMutex
}

func (m *concurrentMockDID) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.did
}

func (m *concurrentMockDID) Method() string     { return "mock" }
func (m *concurrentMockDID) Identifier() string { return "concurrent" }
func (m *concurrentMockDID) Validate() error    { return nil }

func (m *concurrentMockDID) Equal(other NeuronDID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.did == other.String()
}

func (m *concurrentMockDID) PublicKey() (keylib.NeuronPublicKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.matchKey, nil
}

func (m *concurrentMockDID) MatchesKey(pubKey keylib.NeuronPublicKey) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.matchKey.Equal(pubKey)
}

func newConcurrentMockDID(pubKey keylib.NeuronPublicKey) *concurrentMockDID {
	return &concurrentMockDID{
		did:      "did:mock:concurrent",
		matchKey: pubKey,
	}
}

// =============================================================================
// Builder Parallel Usage Tests (Each Goroutine Gets Own Builder)
// =============================================================================

func TestBuilder_ParallelUsage(t *testing.T) {
	t.Run("parallel builders each produce valid account", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan NeuronAccount, 100)
		errs := make(chan error, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Each goroutine creates its own builder - this is the correct pattern
				privKey, _ := keylib.GeneratePrivateKey()
				pubKey := privKey.PublicKey()
				did := newConcurrentMockDID(pubKey)

				account, err := NewParentAccountBuilder(pubKey, did).
					WithStdInHedera("0.0.111").
					Build()

				if err != nil {
					errs <- err
				} else {
					results <- account
				}
			}()
		}

		wg.Wait()
		close(results)
		close(errs)

		// All should succeed
		errCount := 0
		for err := range errs {
			t.Errorf("unexpected error: %v", err)
			errCount++
		}

		successCount := 0
		for acc := range results {
			if acc.IsZero() {
				t.Error("got zero account")
			}
			successCount++
		}

		if successCount == 0 {
			t.Error("expected successful builds")
		}
	})

	t.Run("parallel child account builders", func(t *testing.T) {
		// Generate shared parent
		parentPriv, _ := keylib.GeneratePrivateKey()
		parentPubKey := parentPriv.PublicKey()

		var wg sync.WaitGroup
		results := make(chan NeuronAccount, 50)

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				childPriv, _ := keylib.GeneratePrivateKey()
				childPubKey := childPriv.PublicKey()

				account, err := NewChildAccountBuilder(childPubKey, parentPubKey).
					WithStdInHedera("0.0.444").
					Build()

				if err == nil {
					results <- account
				}
			}()
		}

		wg.Wait()
		close(results)

		// All children should reference same parent
		for acc := range results {
			if !acc.ParentPublicKey().Equal(parentPubKey) {
				t.Error("child should reference correct parent")
			}
		}
	})
}

// =============================================================================
// Validator Parallel Usage Tests
// =============================================================================

func TestAccountValidator_ParallelUsage(t *testing.T) {
	t.Run("parallel validators each produce correct result", func(t *testing.T) {
		var wg sync.WaitGroup
		validCount := make(chan bool, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Each goroutine creates its own validator
				privKey, _ := keylib.GeneratePrivateKey()
				pubKey := privKey.PublicKey()
				did := newConcurrentMockDID(pubKey)

				v := NewAccountValidator().
					ValidatePublicKey(pubKey).
					ValidateDID(did).
					ValidateAccountType(AccountTypeParent)

				validCount <- v.IsValid()
			}()
		}

		wg.Wait()
		close(validCount)

		// All should be valid
		for valid := range validCount {
			if !valid {
				t.Error("expected all validators to pass")
			}
		}
	})

	t.Run("parallel validation of same immutable account", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		// Build once
		account, _ := NewParentAccountBuilder(pubKey, did).Build()

		var wg sync.WaitGroup
		results := make([]error, 100)

		// Validate concurrently (reading immutable data)
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				results[idx] = account.Validate()
			}(i)
		}

		wg.Wait()

		// All should return same result
		for i, err := range results {
			if (err == nil) != (results[0] == nil) {
				t.Errorf("inconsistent validation at index %d", i)
			}
		}
	})
}

func TestValidationResult_ParallelCreation(t *testing.T) {
	t.Run("parallel ValidationResult creation", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan *ValidationResult, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				// Each goroutine creates its own result
				result := NewValidationResult()
				result.AddError(errors.New("error"))
				results <- result
			}(i)
		}

		wg.Wait()
		close(results)

		// All should have exactly one error
		for result := range results {
			if len(result.Errors()) != 1 {
				t.Errorf("expected 1 error, got %d", len(result.Errors()))
			}
		}
	})
}

// =============================================================================
// Collection Concurrent Access Tests
// =============================================================================

func TestReachableAddrs_ConcurrentIteration(t *testing.T) {
	t.Run("concurrent Addrs calls", func(t *testing.T) {
		testPubKeyHex := "0x02759b048e7ccf6ba68f9658105a4a139b5f9f5dfd451857c600cc28f33a1a99ae"
		testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"

		addr1, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID)
		addr2, _ := ParseReachableAddr("/ip4/192.168.1.2/tcp/4001/p2p/" + testPeerID)

		_ = testPubKeyHex // Used for context

		addrs := NewReachableAddrs(addr1, addr2)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(3)

			go func() {
				defer wg.Done()
				_ = addrs.Addrs()
			}()

			go func() {
				defer wg.Done()
				_ = addrs.Strings()
			}()

			go func() {
				defer wg.Done()
				_ = addrs.First()
			}()
		}

		wg.Wait()
		// No panics, race detector passes
	})

	t.Run("concurrent DirectAddrs and RelayAddrs", func(t *testing.T) {
		testPeerID := "16Uiu2HAm3Lkn9NRieuh3UUTWMNthSDumQL9ctTBKxQqdCC79WUSq"
		relayPeerID := "12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN"

		directAddr, _ := ParseReachableAddr("/ip4/192.168.1.1/tcp/4001/p2p/" + testPeerID)
		relayAddr, _ := ParseReachableAddr("/ip4/192.168.1.100/tcp/4001/p2p/" + relayPeerID + "/p2p-circuit/p2p/" + testPeerID)

		addrs := NewReachableAddrs(directAddr, relayAddr)

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(2)

			go func() {
				defer wg.Done()
				_ = addrs.DirectAddrs()
			}()

			go func() {
				defer wg.Done()
				_ = addrs.RelayAddrs()
			}()
		}

		wg.Wait()
		// No panics
	})
}

func TestCommAddressSet_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent Addresses and First", func(t *testing.T) {
		addr1, _ := NewHederaTopicAddress("0.0.111")
		addr2, _ := NewHederaTopicAddress("0.0.222")
		addr3, _ := NewHederaTopicAddress("0.0.333")

		set := NewCommAddressSet(addr1, addr2, addr3)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(2)

			go func() {
				defer wg.Done()
				_ = set.Addresses()
			}()

			go func() {
				defer wg.Done()
				_ = set.First()
			}()
		}

		wg.Wait()
		// No panics
	})
}

// =============================================================================
// Account Concurrent Access Tests
// =============================================================================

func TestNeuronAccount_ConcurrentAccessors(t *testing.T) {
	t.Run("concurrent accessor calls", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		peerID, _ := pubKey.PeerID()
		addr := "/ip4/192.168.1.1/tcp/4001/p2p/" + peerID.String()

		account, err := NewParentAccountBuilder(pubKey, did).
			WithHederaTopics("0.0.111", "0.0.222", "0.0.333").
			WithReachableAddr(addr).
			Build()
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(8)

			go func() {
				defer wg.Done()
				_ = account.PublicKey()
			}()

			go func() {
				defer wg.Done()
				_ = account.PeerID()
			}()

			go func() {
				defer wg.Done()
				_ = account.EVMAddress()
			}()

			go func() {
				defer wg.Done()
				_ = account.AccountType()
			}()

			go func() {
				defer wg.Done()
				_ = account.DID()
			}()

			go func() {
				defer wg.Done()
				_ = account.StdIn()
			}()

			go func() {
				defer wg.Done()
				_ = account.ReachableAddrs()
			}()

			go func() {
				defer wg.Done()
				_ = account.String()
			}()
		}

		wg.Wait()
		// No panics, race detector passes
	})

	t.Run("concurrent Validate calls", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).Build()

		var wg sync.WaitGroup
		errors := make([]error, 50)

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				errors[idx] = account.Validate()
			}(i)
		}

		wg.Wait()

		// All should return same result
		for i, err := range errors {
			if (err == nil) != (errors[0] == nil) {
				t.Errorf("inconsistent Validate at index %d", i)
			}
		}
	})

	t.Run("concurrent MarshalJSON calls", func(t *testing.T) {
		privKey, _ := keylib.GeneratePrivateKey()
		pubKey := privKey.PublicKey()
		did := newConcurrentMockDID(pubKey)

		account, _ := NewParentAccountBuilder(pubKey, did).
			WithStdInHedera("0.0.111").
			Build()

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = account.MarshalJSON()
			}()
		}

		wg.Wait()
		// No panics
	})
}

// =============================================================================
// Stress Tests
// =============================================================================

func TestBuilder_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	t.Run("high concurrency builder usage", func(t *testing.T) {
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				privKey, _ := keylib.GeneratePrivateKey()
				pubKey := privKey.PublicKey()
				did := newConcurrentMockDID(pubKey)

				b := NewParentAccountBuilder(pubKey, did)

				// Multiple operations
				b.WithStdInHedera("0.0.111")
				b.WithStdOutHedera("0.0.222")
				b.WithStdErrHedera("0.0.333")

				_, _ = b.Build()
			}()
		}

		wg.Wait()
		// No panics
	})
}
