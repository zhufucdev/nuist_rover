package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// i-NUIST login data is in AES 128 ECB with PKCS7 padding

const ENCRYPTION_KEY = "axaQiQpsdFAacccs"

func GenerateEncryptionKey(username string) string {
	data := ENCRYPTION_KEY + username
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func Encrypt(key string, data string) (string, error) {
	plaintext := pkcs7Pad([]byte(data), aes.BlockSize)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintext))

	// ECB: encrypt each block separately
	if len(plaintext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("plaintext is not a multiple of block size")
	}
	for i := 0; i < len(plaintext); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], plaintext[i:i+aes.BlockSize])
	}
	return hex.EncodeToString(ciphertext), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
