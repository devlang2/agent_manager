package libs

import (
	//	"bytes"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"strconv"

	//	"github.com/davecgh/go-spew/spew"
	//	"encoding/binary"
	//	"fmt"
	//	"math"
)

func Encrypt(key, plaintext []byte) ([]byte, error) {
	plaintext = pad(plaintext)
	if len(plaintext)%aes.BlockSize != 0 {
		return nil, errors.New("plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte(""), err
	}

	if len(ciphertext) < aes.BlockSize {
		return []byte(""), errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	data := make([]byte, len(ciphertext)-aes.BlockSize)
	copy(data, ciphertext[aes.BlockSize:])

	if len(ciphertext[aes.BlockSize:])%aes.BlockSize != 0 {
		return []byte(""), errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, ciphertext[aes.BlockSize:])
	return unpad(data)
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)              // 144
	unpadding := int(src[length-1]) // 14

	if unpadding > length {
		return nil, errors.New("Unpad error. This could happen when incorrect encryption key is used")
	}
	return src[:(length - unpadding)], nil
}

func ByteToFloat64(b []byte) float64 {
	f, _ := strconv.ParseFloat(string(b), 64)
	return f
}

func ByteToInt64(b []byte) int64 {
	str := string(b)
	i, _ := strconv.ParseInt(str, 10, 64)
	return i
}
