# Old Neuron SDK Structure Analysis

This document provides a comprehensive analysis of the original `neuron-go-hedera-sdk` codebase for reference during development of the new SDK.

## Overview

The old SDK is a decentralized P2P SDK for buyer/seller interactions, using:
- **Hedera Hashgraph** for rendezvous and communication topics (HCS)
- **libp2p** for P2P connectivity and data streaming
- **Smart contracts** for peer registration and discovery

Module: `github.com/NeuronInnovations/neuron-go-hedera-sdk`

---

## Package Structure

### 1. `neuron-sdk.go` - Main Entry Point

**Key Functions:**
- `LaunchSDK()` - Main entry point that initializes P2P host, Hedera topics, and launches buyer/seller protocols
- `SetupKeysAndLocation()` - Initializes keys from environment or configurator
- `LoadPrivateKey()` - Decodes secp256k1 private key from hex
- `hederaAnnounceAndHeartBeat()` - Ensures topics exist and sends periodic heartbeats

**Key Types:**
- `Seller` - Exported type from stream-buyer-vs-seller package

**Global Variables:**
- `fixPrivKey_g` - libp2p crypto private key
- `Version` - SDK version string

**Operational Modes:**
1. **Relay Mode** - Acts as relay for NAT traversal (WIP)
2. **Peer Mode** - Direct P2P participation as buyer or seller

---

### 2. `keylib/` Package

**Files:** `convert.go`, `exchange.go`

**Key Functions:**
- `ConvertHederaPublicKeyToPeerID(hederaPublicKey string) (string, error)` - Converts Hedera public key to libp2p peer ID
- `ConverHederaPublicKeyToEthereunAddress(hederaPublicKey string) string` - Converts to Ethereum address (0x stripped, lowercase)
- `IsValidEthereumAddress(address string) bool` - Validates Ethereum address format
- `EncryptForOtherside(message []byte, myPrivateKeyETH, otherPubliKeyHedera string) ([]byte, error)` - ECDH encryption with AES-CFB
- `DecryptFromOtherside(message []byte, myPrivateKeyEth, otherPubliKeyHedera string) ([]byte, error)` - Corresponding decryption

**Crypto Details:**
- Uses secp256k1 curve (S256)
- AES encryption with CFB mode
- Fixed IV derived from hardcoded `runningHashLong`

---

### 3. `types/` Package

**Files:** `messages.go`, `buffer.go`, `connection.go`

**Message Types:**
- `NeuronHeartBeatMsg` - Periodic status broadcasts
- `NeuronServiceRequestMsg` - Buyer requests service from seller
- `NeuronScheduleSignRequestMsg` - Seller invoice to buyer
- `NeuronPunchMeRequestMsg` - Hole punching initiation
- `NeuronPunchMeAcknowledgmentMsg` - Hole punching acknowledgment
- `NeuronPeerErrorMsg` - Peer-to-peer error messages
- `NeuronSelfErrorMsg` - Self-reported errors

**Connection States:**
```go
const (
    ConnectionLost, ConnectionLostFlushError, ConnectionLostWriteError
    CanNotConnectUnknownReason, CanNotConnectStreamError
    Connected, Connecting, Reconnecting
    HolePunchingScheduled, HolePunchingInProgress, HolePunchingCompleted
)
```

**Rendezvous States:**
```go
const (
    NotInitiated, SendOK, SendFail
    ReceivedOK, ReceivedFail
    WeDoNotKnowPeer, HoldYourHorses
)
```

**Error Types:**
```go
const (
    TooEarlyDialError, DialError, FlushError, WriteError
    StreamError, DisconnectedError, NoKnownAddressError
    VersionError, BalanceError, IpDecryptionError
    HeartBeatError, ServiceError, BadMessageError, ExplorerReachError
)
```

**Buffer Types:**
- `NodeBufferInfo` - Runtime info for remote peers (connection state, attempts, timestamps)
- `TopicPostalEnvelope` - Message wrapper with destination topic
- `PeerStatusInfo` - Detailed peer connection status

---

### 4. `hedera/` Package

**Files:** `main.go`, `http.go`, `rpc.go`

**Core Functions:**
- `GetHederaClientUsingEnv()` - Creates Hedera client from environment
- `GetHRpcClient()` - Creates RPC client for smart contract
- `CreateTopic()` - Creates new HCS topic
- `SendToTopic()` - Sends message to HCS topic
- `BuyerPrepareServiceRequest()` - Prepares service request with shared account
- `SellerSendScheduledTransferRequest()` - Creates scheduled transfer for payment
- `BuyerCounterSignSchedule()` - Signs schedule to release payment
- `ListenToTopicAndCallBack()` - Subscribes to HCS topic
- `PeerSendErrorMessage()` - Sends error to peer's stdin topic
- `SendSelfErrorMessage()` - Logs error to own stderr topic
- `EnsureTopicsAndNotifyContract()` - Verifies/creates topics and registers with smart contract

**Mirror API Functions:**
- `GetAccountInfoFromMirror()` - Gets account info from mirror node
- `GetAllDevicesFromExplorer()` - Gets devices from Neuron explorer
- `GetLastMessageFromTopic()` - Gets recent HCS messages
- `GetDeviceParent()` - Finds parent account of device

**Smart Contract Functions:**
- `GetPeerInfo()` - Gets peer info from smart contract
- `GetPeerArraySize()` - Gets number of registered peers
- `GetAllPeers()` - Lists all registered peers

---

### 5. `hederacontract/` Package

**File:** `Rendezvous.go` (auto-generated Go bindings)

**Contract ABI Functions:**
- `hederaAddressToPeer(address) -> (available, peerID, stdOutTopic, stdInTopic, stdErrTopic)`
- `peerList(index) -> address`
- `getPeerArraySize() -> uint256`
- `putPeerAvailableSelf(stdOut, stdIn, stdErr, peerID, serviceIDs, prices)` - Self-registration
- `allowedServices(index) -> string`

---

### 6. `common-lib/` Package

**Files:** `buffers.go`, `flags.go`, `environment.go`, `connection.go`

**NodeBuffers Management:**
- `NewNodeBuffers()` - Creates buffer manager
- `AddBuffer2/3()` - Adds peer buffer with state
- `GetBuffer()` - Gets peer buffer info
- `RemoveBuffer()` - Removes peer buffer
- `UpdateBufferLibP2PState()` - Updates connection state
- `UpdateBufferRendezvousState()` - Updates rendezvous state
- `IncrementReconnectAttempts()` - Exponential backoff tracking
- `ShowDetailedPeerStatus()` - Gets all peer statuses

**Command-Line Flags:**
- `--mode` (peer/relay)
- `--port`
- `--buyer-or-seller`
- `--force-location`
- `--my-public-ip`, `--my-public-port`
- `--list-of-sellers-source` (explorer/env)
- `--radius` (km filtering)
- `--use-local-address`
- `--smart-contract-address`
- `--enable-upnp`
- `--clear-cache`
- `--envFile`

**Global Variables:**
- `MyEnvFile`, `MyProtocol`, `MyStdIn`, `MyStdOut`, `MyStdErr`
- `MyPublicKey`, `MyPrivateKey`, `MyLocation`

**Connection Functions:**
- `InitialConnect()` - Establishes P2P connection
- `IsRequestTooEarly()` - Exponential backoff check
- `WriteAndFlushBuffer()` - Writes to peer stream
- `HolePunchConnectIfNotConnected()` - NAT traversal connection

---

### 7. `dapp-protocols/stream-buyer-vs-seller/` Package

**Files:** `buyer-case.go`, `seller-case.go`

**Buyer Functions:**
- `HandleBuyerCase()` - Main buyer protocol handler
- `processSeller()` - Processes individual seller
- `prepareServiceRequestMsg()` - Creates service request
- `handlePunchMeRequest()` - Handles hole punching from seller
- `performBuyerHolePunching()` - Executes hole punch
- `ReplaceSellers()` - Updates seller list dynamically
- `ShowCurrentPeerStatus()` - Gets current seller map

**Seller Functions:**
- `HandleSellerCase()` - Main seller protocol handler
- `preparePunchMeRequest()` - Creates hole punch request
- `handlePunchMeAcknowledgment()` - Handles buyer ack
- `performSellerHolePunching()` - Executes hole punch

**Seller Discovery:**
1. From environment (`list_of_sellers`)
2. From explorer API (with radius filtering)

---

### 8. `validator-lib/` Package

**File:** `validator.go`

**Functions:**
- `IsRequestPermitted() bool` - Always returns true (placeholder)

---

### 9. `whoami/` Package

**File:** `nat-info.go`

**Key Functions:**
- `GetNatInfoAndUpdateGlobals()` - Determines NAT type via STUN
- `mappingTests()` - RFC5780 NAT mapping behavior tests
- `filteringTests()` - RFC5780 NAT filtering behavior tests

**NAT Types Detected:**
- `no-nat` - Endpoint independent (no NAT)
- `cone` - Endpoint independent mapping
- `port-restricted-cone` - Address dependent
- `symmetric` - Address and port dependent

**Global Variables:**
- `NatDeviceType`, `NatIPAddress`, `NatPort`, `NatReachability`

---

### 10. `upnp/` Package

**File:** `port-forward.go`

**Functions:**
- `SendSSDPRequest()` - Discovers UPnP devices
- `AddPortMapping()` - Adds port mapping rule
- `DeletePortMapping()` - Removes port mapping
- `ListPortMappings()` - Lists current mappings

---

## Communication Flow

### Account Structure
```
Parent Account (holds funds)
    └── Device Account (child)
        ├── stdIn Topic (receives messages)
        ├── stdOut Topic (broadcasts heartbeats)
        └── stdErr Topic (error logging)
```

### Key Derivation
```
secp256k1 Private Key
    └── Public Key / Alias Key
        ├── IPFS Peer ID (libp2p)
        ├── EVM Address (Ethereum/Hedera)
        └── Hedera Account ID
```

### Buyer-Seller Protocol
1. Buyer discovers sellers (explorer or env)
2. Buyer sends `serviceRequest` to seller's stdIn (via HCS)
3. Seller decrypts buyer's IP, attempts connection
4. If direct fails, initiates hole punching protocol
5. Seller creates stream, sends data
6. Seller sends `scheduleSignRequest` (invoice)
7. Buyer counter-signs to release payment

### Hole Punching Protocol
1. Seller sends `punchMeRequest` with encrypted IP
2. Buyer sends `punchMeAcknowledgment`
3. Both schedule simultaneous connection at timestamp + delay
4. Use libp2p's `WithSimultaneousConnect` and `WithForceDirectDial`

---

## Environment Variables

```
private_key=<secp256k1 hex>
hedera_evm_id=<device EVM address>
hedera_id=<device account ID, e.g., 0.0.1234>
location={"lat":50.1,"lon":1.8898,"alt":0.0}
list_of_sellers=<comma-separated public keys>
eth_rpc_url=https://testnet.hashio.io/api
mirror_api_url=https://testnet.mirrornode.hedera.com/api/v1
neuron_explorer_url=https://explorer.neuron.world/api/v1/device/wip-all
smart_contract_address=0x...
```

---

## Dependencies

Key external packages:
- `github.com/hashgraph/hedera-sdk-go/v2` - Hedera SDK
- `github.com/libp2p/go-libp2p` - P2P networking
- `github.com/ethereum/go-ethereum` - Ethereum crypto
- `github.com/multiformats/go-multiaddr` - Multi-address handling
- `github.com/pion/stun` - STUN protocol
- `github.com/umahmood/haversine` - Distance calculations
- `github.com/joho/godotenv` - Environment file loading
