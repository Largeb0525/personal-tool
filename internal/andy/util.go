package andy

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

func base58Encode(input []byte) string {
	var result []byte
	x := new(big.Int).SetBytes(input)
	radix := big.NewInt(58)
	mod := new(big.Int)

	for x.Sign() > 0 {
		x.DivMod(x, radix, mod)
		result = append([]byte{base58Alphabet[mod.Int64()]}, result...)
	}

	// Leading zero bytes
	for _, b := range input {
		if b == 0x00 {
			result = append([]byte{base58Alphabet[0]}, result...)
		} else {
			break
		}
	}

	return string(result)
}

func base58Decode(input string) ([]byte, error) {
	result := big.NewInt(0)
	base := big.NewInt(58)

	for _, r := range input {
		index := strings.IndexRune(base58Alphabet, r)
		if index == -1 {
			return nil, errors.New("invalid base58 character")
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(index)))
	}

	decoded := result.Bytes()

	// 前導 0 處理（base58 有時會省略 leading 0x00）
	leadingZeros := 0
	for i := 0; i < len(input) && input[i] == '1'; i++ {
		leadingZeros++
	}

	return append(make([]byte, leadingZeros), decoded...), nil
}

func toBase58CheckAddress(hexAddr string) (string, error) {
	if len(hexAddr) >= 2 && hexAddr[:2] == "0x" {
		hexAddr = hexAddr[2:]
	}

	addrBytes, err := hex.DecodeString(hexAddr)
	if err != nil {
		return "", errors.New("invalid hex address")
	}

	if len(addrBytes) != 21 {
		return "", errors.New("address should be 21 bytes including '41' prefix")
	}

	// Double SHA256 checksum
	first := sha256.Sum256(addrBytes)
	second := sha256.Sum256(first[:])
	checksum := second[:4]

	full := append(addrBytes, checksum...)
	return base58Encode(full), nil
}

func TronToHexPadded32(tronAddr string) (string, error) {
	raw, err := base58Decode(tronAddr)
	if err != nil {
		return "", err
	}
	if len(raw) < 21 || raw[0] != 0x41 {
		return "", errors.New("invalid TRON address length or prefix")
	}

	addressBytes := raw[1:21] // 去掉 0x41 prefix，取 20 bytes
	hexStr := hex.EncodeToString(addressBytes)

	// 左補 0 到 64 位元 hex（32 bytes）
	padded := strings.Repeat("0", 64-len(hexStr)) + hexStr
	return "0x" + padded, nil
}

func getUSDTBalance(tokens []TokenInfo) string {
	for _, token := range tokens {
		if token.TokenAbbr == "USDT" {
			return token.Balance
		}
	}
	return ""
}

func signTransaction(txIDHex string) (string, error) {
	txIDBytes, err := hex.DecodeString(txIDHex)
	if err != nil {
		return "", fmt.Errorf("decode txID error: %w", err)
	}

	privKeyBytes, err := hex.DecodeString(TronPrivateKey)
	if err != nil {
		return "", fmt.Errorf("decode privKey error: %w", err)
	}

	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return "", fmt.Errorf("parse privKey error: %w", err)
	}

	sig, err := crypto.Sign(txIDBytes, privKey)
	if err != nil {
		return "", fmt.Errorf("sign error: %w", err)
	}

	return hex.EncodeToString(sig), nil
}
