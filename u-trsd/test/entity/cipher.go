package entity

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

// CipherTypeClass is a type that defines the supported cipher types.
type CipherTypeClass uint16

// Known cipher types.
// These types are defined at the document.
const (
	CipherTypeClassAES256CBC CipherTypeClass = iota
	CipherTypeClassAES256CFB
	CipherTypeClassAES256OFB
	CipherTypeClassAES256CTR
	CipherTypeClassAES256GCM
)

// CipherKeyLength is a type that defines the length of the encryption key.
type CipherKeyLength uint16

// padByPkcs7 adds PKCS#7 padding to the data.
// It pads the data to a multiple of the AES block size (16 bytes)
// by appending bytes with a value equal to the number of padding bytes.
// If the data length is already a multiple of the block size, a full block of padding is added.
func padByPkcs7(data []byte) []byte {
	// Calculate the number of bytes to pad.
	padSize := aes.BlockSize
	if len(data)%aes.BlockSize != 0 {
		padSize = aes.BlockSize - (len(data))%aes.BlockSize
	}

	// Create the padding slice.
	pad := bytes.Repeat([]byte{byte(padSize)}, padSize)

	// Append the padding to the original data and return it.
	return append(data, pad...)
}

// unPadByPkcs7 removes PKCS#7 padding from the data.
// It determines the padding length from the last byte of the data and removes that many bytes.
func unPadByPkcs7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		err := errors.New("data slice is empty")
		return nil, err
	}

	// The last byte of the data indicates the padding size.
	padSize := int(data[len(data)-1])

	if padSize >= len(data) {
		return nil, fmt.Errorf("invalid padSize: %d", padSize)
	}

	// Return the slice without the padding.
	return data[:len(data)-padSize], nil
}

// EncryptPacket encrypts the given buffer using AES.
//
// Parameters:
//
//	buf: The byte slice of plaintext to be encrypted.
//	key: The key to use for encryption.
//	keyCipherType: The type of cipher mode to use.
//	keyLength: The length of the encryption key (currently unused).
//
// Returns:
//
//	A byte slice of the encrypted data (IV + ciphertext) and an error.
func EncryptPacket(buf, key []byte, keyCipherType uint16, keyLength uint16) ([]byte, error) {
	// Create a new AES cipher block from the key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Pad the data to be encrypted using PKCS#7.
	paddBuf := padByPkcs7(buf)
	// Prepare a buffer to store the ciphertext (IV + ciphertext).
	cipherBuf := make([]byte, aes.BlockSize+len(paddBuf))
	// Use the beginning of the buffer as the IV (Initialization Vector).
	if len(cipherBuf) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := cipherBuf[:aes.BlockSize]
	// Generate a cryptographically secure random IV.
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate Initialization Vector: %w", err)
	}

	// Branch the process based on the specified cipher mode.
	switch CipherTypeClass(keyCipherType) {
	case CipherTypeClassAES256CBC:
		// Create a CBC mode encrypter stream.
		encryptStream := cipher.NewCBCEncrypter(block, iv)

		// Encrypt the data.
		if len(paddBuf)%aes.BlockSize != 0 {
			return nil, fmt.Errorf("plaintext is not a multiple of the block size")
		}
		encryptStream.CryptBlocks(cipherBuf[aes.BlockSize:], paddBuf)

	case CipherTypeClassAES256CFB:
		// TODO: Implement CFB mode.
		return nil, fmt.Errorf("this type AES256CFB was not implemented yet")

	case CipherTypeClassAES256OFB:
		// TODO: Implement OFB mode.
		return nil, fmt.Errorf("this type AES256OFB was not implemented yet")

	case CipherTypeClassAES256CTR:
		// TODO: Implement CTR mode.
		return nil, fmt.Errorf("this type AES256CTR was not implemented yet")

	case CipherTypeClassAES256GCM:
		// TODO: Implement GCM mode.
		return nil, fmt.Errorf("this type AES256GCM was not implemented yet")

	default:
		return nil, fmt.Errorf("unknown cipher type")
	}

	// Return the encrypted data (IV + ciphertext).
	return cipherBuf, nil
}

// DecryptPacket decrypts the given buffer using AES.
//
// Parameters:
//
//	buf: The byte slice of ciphertext (IV + ciphertext) to be decrypted.
//	key: The key to use for decryption.
//	keyCipherType: The type of cipher mode that was used.
//	keyLength: The length of the encryption key (currently unused).
//
// Returns:
//
//	A byte slice of the decrypted data (plaintext) and an error.
func DecryptPacket(buf, key []byte, keyCipherType uint16, keyLength uint16) ([]byte, error) {
	// mock
	// return buf, nil

	// Create a new AES cipher block from the key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Prepare a buffer to store the decrypted plaintext.
	decryptedText := make([]byte, len(buf[aes.BlockSize:]))

	// The IV is at the beginning of the ciphertext (first aes.BlockSize bytes).
	if len(buf) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := buf[:aes.BlockSize]

	// Branch the process based on the specified cipher mode.
	switch CipherTypeClass(keyCipherType) {
	case CipherTypeClassAES256CBC:
		// Create a CBC mode decrypter stream.
		decryptStream := cipher.NewCBCDecrypter(block, iv)

		// Decrypt the data.
		if len(buf[aes.BlockSize:])%aes.BlockSize != 0 {
			return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
		}
		decryptStream.CryptBlocks(decryptedText, buf[aes.BlockSize:])

	case CipherTypeClassAES256CFB:
		// TODO: Implement CFB mode.
		return nil, fmt.Errorf("this type AES256CFB was not implemented yet")

	case CipherTypeClassAES256OFB:
		// TODO: Implement OFB mode.
		return nil, fmt.Errorf("this type AES256OFB was not implemented yet")

	case CipherTypeClassAES256CTR:
		// TODO: Implement CTR mode.
		return nil, fmt.Errorf("this type AES256CTR was not implemented yet")

	case CipherTypeClassAES256GCM:
		// TODO: Implement GCM mode.
		return nil, fmt.Errorf("this type AES256GCM was not implemented yet")

	default:
		return nil, fmt.Errorf("unknown cipher type")
	}

	// Remove padding and return the plaintext.
	return unPadByPkcs7(decryptedText)
}
