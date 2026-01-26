# Hedera Interaction from Neuron Account - Old SDK Analysis

This document describes how the old SDK (`neuron-go-hedera-sdk/`) interacts with Hedera from a Neuron Account and publishes heartbeats. This analysis is intended to inform the implementation in the new SDK.

---

## Overview

The Neuron SDK uses Hedera Consensus Service (HCS) topics for communication and a smart contract for peer registration/discovery. The flow is:

1. **Load private key and location** from environment
2. **Derive EVM address** from public key
3. **Query smart contract** to get registered HCS topics (stdin, stdout, stderr)
4. **Send heartbeat messages** periodically to the stdout topic

---

## Architecture Components

### 1. Hedera Client Setup

**File**: [neuron-go-hedera-sdk/hedera/main.go](neuron-go-hedera-sdk/hedera/main.go)

```go
func GetHederaClientUsingEnv() *hedera.Client {
    c, _ := hedera.ClientForName(hedera.NetworkNameTestnet.String())
    op, _ := hedera.AccountIDFromString(os.Getenv("hedera_id"))
    
    pkString := os.Getenv("private_key")
    
    if len(pkString) == 64 {
        pk, _ := hedera.PrivateKeyFromStringECDSA(pkString)
        c.SetOperator(op, pk)
    } else {
        pk, _ := hedera.PrivateKeyFromStringEd25519(pkString)
        c.SetOperator(op, pk)
    }
    return c
}
```

**Environment Variables Required**:
- `hedera_id` - Hedera account ID (e.g., "0.0.12345")
- `private_key` - Hex-encoded private key (64 chars for ECDSA, longer for Ed25519)

### 2. Smart Contract RPC Client

**File**: [neuron-go-hedera-sdk/hedera/main.go](neuron-go-hedera-sdk/hedera/main.go)

```go
func GetHRpcClient() *hederacontract.HederacontractCaller {
    scAddress := os.Getenv("smart_contract_address")  // or --smart-contract-address flag
    client, _ := ethclient.Dial(os.Getenv("eth_rpc_url"))
    contractCaller, _ := hederacontract.NewHederacontractCaller(
        common.HexToAddress(scAddress),
        client,
    )
    return contractCaller
}
```

**Environment Variables Required**:
- `smart_contract_address` - Ethereum-style address of the rendezvous contract (0x...)
- `eth_rpc_url` - Hedera JSON-RPC endpoint (e.g., https://testnet.hashio.io/api)

### 3. PeerInfo from Smart Contract

**File**: [neuron-go-hedera-sdk/hedera/rpc.go](neuron-go-hedera-sdk/hedera/rpc.go)

```go
type PeerInfo struct {
    Available   bool
    PeerID      string
    StdOutTopic uint64  // HCS topic ID for peer's output
    StdInTopic  uint64  // HCS topic ID for peer's input
    StdErrTopic uint64  // HCS topic ID for peer's errors
}

func GetPeerInfo(hederaAccEvmAddress string) (PeerInfo, error) {
    contractCaller := GetHRpcClient()
    peerInfo, err := contractCaller.HederaAddressToPeer(
        &bind.CallOpts{},
        common.HexToAddress(hederaAccEvmAddress),
    )
    return peerInfo, err
}
```

---

## Heartbeat Flow

### Step 1: Ensure Topics and Notify Contract

**File**: [neuron-go-hedera-sdk/hedera/rpc.go](neuron-go-hedera-sdk/hedera/rpc.go)

```go
func EnsureTopicsAndNotifyContract(p2pHost host.Host) (hedera.TopicID, hedera.TopicID, hedera.TopicID, error) {
    // 1. Extract public key from p2p host
    hostPubKey, _ := p2pHost.ID().ExtractPublicKey()
    hostPubKeyByte, _ := hostPubKey.Raw()
    hostPubKeyStr := common.Bytes2Hex(hostPubKeyByte)

    // 2. Convert to EVM address
    toEthAddress := keylib.ConverHederaPublicKeyToEthereunAddress(hostPubKeyStr)

    // 3. Query smart contract for registered topics
    peerInfo, err := GetPeerInfo(toEthAddress)
    if err != nil {
        log.Panic("Cannot talk to smart contract:", err)
    }

    // 4. Return topics from contract
    if peerInfo.StdInTopic != 0 && peerInfo.StdOutTopic != 0 {
        stdOutTopicID, _ := hedera.TopicIDFromString(fmt.Sprintf("0.0.%d", peerInfo.StdOutTopic))
        stdInTopicID, _ := hedera.TopicIDFromString(fmt.Sprintf("0.0.%d", peerInfo.StdInTopic))
        stdErrTopicID, _ := hedera.TopicIDFromString(fmt.Sprintf("0.0.%d", peerInfo.StdErrTopic))
        return stdOutTopicID, stdInTopicID, stdErrTopicID, nil
    }
    
    // Self-registration branch (disabled in production)
    // Creates new topics and calls putPeerAvailableSelf() on contract
}
```

### Step 2: Send Heartbeat Messages

**File**: [neuron-go-hedera-sdk/neuron-sdk.go](neuron-go-hedera-sdk/neuron-sdk.go)

```go
func hederaAnnounceAndHeartBeat(ctx context.Context, p2pHost host.Host) (hedera.TopicID, hedera.TopicID, hedera.TopicID, error) {
    // 1. Get topics from contract
    stdOutTopic, stdInTopic, stdErrTopic, err := hedera_helper.EnsureTopicsAndNotifyContract(p2pHost)

    // 2. Start heartbeat goroutine
    go func() {
        for {
            peers := p2pHost.Network().Peers()
            myPublicAddresses := hostsPublicAddressesSorted(p2pHost)
            
            if len(myPublicAddresses) > 0 {
                // Build heartbeat message
                heartbeatMessage := types.NeuronHeartBeatMsg{
                    MessageType:        "NeuronHeartBeat",
                    Location:           commonlib.MyLocation,
                    NatDeviceType:      "",
                    NatReachability:    whoami.NatReachability,
                    BuyerOrSeller:      *flags.BuyerOrSellerFlag,
                    Version:            Version,
                    ConnectedPeersAbrv: getAbbreviatedPeerPublicKeys(p2pHost),
                }
                
                // Marshal to JSON
                heartbeatJSON, _ := json.Marshal(heartbeatMessage)
                
                // Send to HCS topic
                err = hedera_helper.SendToTopic(stdOutTopic, string(heartbeatJSON))
            }
            
            time.Sleep(time.Second * 40)  // Every 40 seconds
        }
    }()
    
    return stdOutTopic, stdInTopic, stdErrTopic, nil
}
```

### Step 3: Send Message to HCS Topic

**File**: [neuron-go-hedera-sdk/hedera/main.go](neuron-go-hedera-sdk/hedera/main.go)

```go
func SendToTopic(topicID hedera.TopicID, content string) error {
    client := GetHederaClientUsingEnv()
    defer client.Close()

    _, err := hedera.NewTopicMessageSubmitTransaction().
        SetMessage([]byte(content)).
        SetTopicID(topicID).
        Execute(client)

    return err
}
```

---

## Message Types

### NeuronHeartBeatMsg

**File**: [neuron-go-hedera-sdk/types/messages.go](neuron-go-hedera-sdk/types/messages.go)

```go
type NeuronHeartBeatMsg struct {
    MessageType        string                 `json:"messageType"`    // "NeuronHeartBeat"
    Location           EnvironmentVarLocation `json:"location"`
    NatDeviceType      string                 `json:"natDeviceType"`
    NatReachability    bool                   `json:"natReachability"`
    Version            string                 `json:"version"`
    BuyerOrSeller      string                 `json:"buyerOrSeller"`
    ConnectedPeersAbrv []string               `json:"connectedPeers"` // Abbreviated peer IDs
}

type EnvironmentVarLocation struct {
    Latitude  float64 `json:"lat"`
    Longitude float64 `json:"lon"`
    Altitude  float64 `json:"alt"`
    GPSFix    string  `json:"gpsfix"`
}
```

---

## Key Conversion Functions

**File**: [neuron-go-hedera-sdk/keylib/convert.go](neuron-go-hedera-sdk/keylib/convert.go)

### Public Key → PeerID
```go
func ConvertHederaPublicKeyToPeerID(hederaPublicKey string) (string, error) {
    hpub, _ := hedera.PublicKeyFromString(hederaPublicKey)
    libp2pPubKey, _ := libp2pcrypto.UnmarshalSecp256k1PublicKey(hpub.BytesRaw())
    peerID, _ := peer.IDFromPublicKey(libp2pPubKey)
    return peerID.String(), nil
}
```

### Public Key → EVM Address
```go
func ConverHederaPublicKeyToEthereunAddress(hederaPublicKey string) string {
    publicKeyBytes, _ := hex.DecodeString(hederaPublicKey)
    decompressedPubkey, _ := ethcrypto.DecompressPubkey(publicKeyBytes)
    ethaddress := ethcrypto.PubkeyToAddress(*decompressedPubkey)
    return strings.ToLower(ethaddress.String()[2:])  // Without 0x prefix
}
```

---

## Smart Contract Functions

**File**: [neuron-go-hedera-sdk/hederacontract/Rendezvous.go](neuron-go-hedera-sdk/hederacontract/Rendezvous.go)

### Reading (via RPC)
- `HederaAddressToPeer(address)` → Returns `PeerInfo` (topics, availability, peerID)
- `GetPeerArraySize()` → Returns count of registered peers
- `PeerList(index)` → Returns peer address at index

### Writing (via Hedera SDK)
- `PutPeerAvailableSelf(stdOutTopic, stdInTopic, stdErrTopic, peerID, serviceIDs, prices)` → Registers peer

---

## Environment Variables Summary

| Variable | Description | Example |
|----------|-------------|---------|
| `hedera_id` | Hedera account ID | `0.0.12345` |
| `private_key` | Hex-encoded private key (ECDSA 64 chars) | `4c0883a69102...` |
| `smart_contract_address` | Rendezvous contract EVM address | `0x1234...abcd` |
| `eth_rpc_url` | Hedera JSON-RPC URL | `https://testnet.hashio.io/api` |
| `location` | JSON location data | `{"lat":37.7,"lon":-122.4,"alt":0}` |

---

## Global State (commonlib)

**File**: [neuron-go-hedera-sdk/common-lib/environment.go](neuron-go-hedera-sdk/common-lib/environment.go)

```go
var (
    MyEnvFile    string
    MyProtocol   protocol.ID
    MyStdIn      hedera.TopicID    // Peer's input topic
    MyStdOut     hedera.TopicID    // Peer's output topic
    MyStdErr     hedera.TopicID    // Peer's error topic
    MyPublicKey  hedera.PublicKey
    MyPrivateKey crypto.PrivateKey
    MyLocation   types.EnvironmentVarLocation
)
```

---

## Key Implementation Notes

1. **ECDSA Only**: The new SDK should enforce ECDSA secp256k1 keys only (Ed25519 rejected)
2. **Topic Discovery**: Topics are pre-registered in the smart contract, not created on-the-fly
3. **Heartbeat Interval**: 40 seconds between heartbeats
4. **Address Derivation**: Use Keccak256 hash of uncompressed public key, take last 20 bytes
5. **Retry Logic**: GetPeerInfo has exponential backoff (25 retries max)
6. **Client Management**: Each operation creates and closes its own Hedera client

---

## Suggested Implementation for New SDK

1. Create a `HederaService` that manages:
   - Client creation with proper key handling
   - Topic message submission
   - Smart contract queries via JSON-RPC

2. Create a `HeartbeatService` that:
   - Periodically sends NeuronHeartBeatMsg to stdOutTopic
   - Uses the NeuronPrivateKey from keylib for signing

3. Create a `NeuronAccount` type that encapsulates:
   - Private key (using new keylib.NeuronPrivateKey)
   - Hedera Account ID
   - Registered HCS topics (stdin, stdout, stderr)
   - Location information

4. Use the existing keylib for:
   - Key parsing and validation
   - EVM address derivation (already implemented)
   - PeerID derivation (already implemented)
