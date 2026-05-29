package csustkit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const randomPasswordChars = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"

func encryptPassword(password, salt string) (string, error) {
	if salt == "" {
		return password, nil
	}
	if len(salt) < aes.BlockSize {
		return "", fmt.Errorf("pwdEncryptSalt 长度不足")
	}

	prefix, err := randomString(64)
	if err != nil {
		return "", err
	}
	iv, err := randomString(aes.BlockSize)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(salt[:aes.BlockSize]))
	if err != nil {
		return "", err
	}

	plain := pkcs7Pad([]byte(prefix+password), aes.BlockSize)
	encrypted := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, []byte(iv)).CryptBlocks(encrypted, plain)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func randomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	out := make([]byte, length)
	for i, b := range bytes {
		out[i] = randomPasswordChars[int(b)%len(randomPasswordChars)]
	}
	return string(out), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+padding)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(padding)
	}
	return out
}
