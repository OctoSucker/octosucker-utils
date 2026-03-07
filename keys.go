package utils

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
)

// PrivateKey 私钥接口，定义了通用的私钥操作（对上层屏蔽具体网络细节）
type PrivateKey interface {
	NetworkType() string
	Address() string
	Sign(message []byte) ([]byte, error)
	Verify(message []byte, signature []byte) (bool, error)
}

// NetworkKeyPair 网络配置：只包含 skill 需要的最小字段
// 注意：这里不依赖 a2a-x402，主工程可以通过相同字段名映射过来
type NetworkKeyPair struct {
	NetworkName string `json:"networkName"`
	PrivateKey  string `json:"privateKey"`
}

// HashMessage 使用 Ethereum personal_sign 标准计算消息哈希
func HashMessage(message string) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	return crypto.Keccak256([]byte(prefix + message))
}

// hashMessage 内部使用的哈希函数（小写开头，不导出）
func hashMessage(message string) []byte {
	return HashMessage(message)
}

// NewPrivateKeyFromNetworkKeyPair 根据网络配置创建合适的私钥实现
// 约定：
// - EVM 网络：NetworkName 以 "eip155:" 开头，PrivateKey 为 0x 前缀的十六进制
// - Solana 网络：NetworkName 以 "solana:" 开头，PrivateKey 为 base58 字符串
func NewPrivateKeyFromNetworkKeyPair(pair NetworkKeyPair) (PrivateKey, error) {
	if pair.NetworkName == "" {
		return nil, fmt.Errorf("networkName is required")
	}
	if pair.PrivateKey == "" {
		return nil, fmt.Errorf("privateKey is required")
	}

	// EVM 系列网络
	if strings.HasPrefix(pair.NetworkName, "eip155:") {
		privateKeyHex := strings.TrimPrefix(strings.TrimSpace(pair.PrivateKey), "0x")
		ecdsaKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EVM private key: %w", err)
		}
		return NewEVMPrivateKey(ecdsaKey, pair.NetworkName), nil
	}

	// Solana 系列网络
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

// EVMPrivateKey EVM 兼容网络的私钥实现
type EVMPrivateKey struct {
	key         *ecdsa.PrivateKey
	networkType string
}

// NewEVMPrivateKey 创建 EVM 私钥
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
	// Ethereum expects V to be 27 or 28
	signature[64] += 27
	return signature, nil
}

func (k *EVMPrivateKey) Verify(message []byte, signature []byte) (bool, error) {
	if len(signature) != 65 {
		return false, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(signature))
	}

	// 复制签名以避免修改原始数据
	sigBytes := make([]byte, len(signature))
	copy(sigBytes, signature)

	// 移除 27/28（Ethereum personal_sign 标准）
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

// SolanaPrivateKey Solana 网络的私钥实现
type SolanaPrivateKey struct {
	key         solana.PrivateKey
	networkType string
}

// NewSolanaPrivateKey 创建 Solana 私钥
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
	// Solana 签名：使用私钥签名消息
	signature, err := k.key.Sign(message)
	if err != nil {
		return nil, err
	}
	// 将 solana.Signature 数组转换为 []byte
	return signature[:], nil
}

func (k *SolanaPrivateKey) Verify(message []byte, signature []byte) (bool, error) {
	// Solana 签名验证：使用公钥验证签名
	publicKey := k.key.PublicKey()
	// 将 []byte 转换为 solana.Signature
	var sig solana.Signature
	if len(signature) != len(sig) {
		return false, fmt.Errorf("invalid signature length: expected %d bytes, got %d", len(sig), len(signature))
	}
	copy(sig[:], signature)
	return publicKey.Verify(message, sig), nil
}
