package utils

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
)

type PrivateKey interface {
	NetworkType() string
	Address() string
	Sign(message []byte) ([]byte, error)
	Verify(message []byte, signature []byte) (bool, error)
}

type NetworkKeyPair struct {
	NetworkName string `json:"networkName"`
	PrivateKey  string `json:"privateKey"`
}

func HashMessage(message string) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	return crypto.Keccak256([]byte(prefix + message))
}

func hashMessage(message string) []byte {
	return HashMessage(message)
}

func NewPrivateKeyFromNetworkKeyPair(pair NetworkKeyPair) (PrivateKey, error) {
	if pair.NetworkName == "" {
		return nil, fmt.Errorf("networkName is required")
	}
	if pair.PrivateKey == "" {
		return nil, fmt.Errorf("privateKey is required")
	}

	if strings.HasPrefix(pair.NetworkName, "eip155:") {
		privateKeyHex := strings.TrimPrefix(strings.TrimSpace(pair.PrivateKey), "0x")
		ecdsaKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EVM private key: %w", err)
		}
		return NewEVMPrivateKey(ecdsaKey, pair.NetworkName), nil
	}

	if strings.HasPrefix(pair.NetworkName, "solana:") {
		privateKeyStr := strings.TrimSpace(pair.PrivateKey)
		solanaKey, err := solana.PrivateKeyFromBase58(privateKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Solana private key: %w", err)
		}
		return NewSolanaPrivateKey(solanaKey, pair.NetworkName), nil
	}

	return nil, fmt.Errorf("unsupported network type: %s", pair.NetworkName)
}

type EVMPrivateKey struct {
	key         *ecdsa.PrivateKey
	networkType string
}

func NewEVMPrivateKey(key *ecdsa.PrivateKey, networkType string) *EVMPrivateKey {
	return &EVMPrivateKey{
		key:         key,
		networkType: networkType,
	}
}

func (k *EVMPrivateKey) NetworkType() string {
	return k.networkType
}

func (k *EVMPrivateKey) Address() string {
	publicKey := k.key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return ""
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
}

func (k *EVMPrivateKey) Sign(message []byte) ([]byte, error) {
	hash := hashMessage(string(message))
	signature, err := crypto.Sign(hash, k.key)
	if err != nil {
		return nil, err
	}
	signature[64] += 27
	return signature, nil
}

func (k *EVMPrivateKey) Verify(message []byte, signature []byte) (bool, error) {
	if len(signature) != 65 {
		return false, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(signature))
	}

	sigBytes := make([]byte, len(signature))
	copy(sigBytes, signature)

	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	hash := hashMessage(string(message))
	pubKey, err := crypto.SigToPub(hash[:], sigBytes)
	if err != nil {
		return false, err
	}

	recoveredAddress := crypto.PubkeyToAddress(*pubKey)
	expectedAddress := k.Address()
	return recoveredAddress.Hex() == expectedAddress, nil
}

type SolanaPrivateKey struct {
	key         solana.PrivateKey
	networkType string
}

func NewSolanaPrivateKey(key solana.PrivateKey, networkType string) *SolanaPrivateKey {
	return &SolanaPrivateKey{
		key:         key,
		networkType: networkType,
	}
}

func (k *SolanaPrivateKey) NetworkType() string {
	return k.networkType
}

func (k *SolanaPrivateKey) Address() string {
	return k.key.PublicKey().String()
}

func (k *SolanaPrivateKey) Sign(message []byte) ([]byte, error) {
	signature, err := k.key.Sign(message)
	if err != nil {
		return nil, err
	}
	return signature[:], nil
}

func (k *SolanaPrivateKey) Verify(message []byte, signature []byte) (bool, error) {
	publicKey := k.key.PublicKey()
	var sig solana.Signature
	if len(signature) != len(sig) {
		return false, fmt.Errorf("invalid signature length: expected %d bytes, got %d", len(sig), len(signature))
	}
	copy(sig[:], signature)
	return publicKey.Verify(message, sig), nil
}
