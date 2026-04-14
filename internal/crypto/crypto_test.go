package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chiragguptadtu/trigger/internal/crypto"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := "super secret value"
	ciphertext, err := crypto.Encrypt(key, plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, []byte(plaintext), ciphertext)

	decrypted, err := crypto.Decrypt(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncrypt_ProducesUniqueOutput(t *testing.T) {
	key := make([]byte, 32)
	plaintext := "same input"

	c1, err := crypto.Encrypt(key, plaintext)
	require.NoError(t, err)
	c2, err := crypto.Encrypt(key, plaintext)
	require.NoError(t, err)

	// Random nonce means two encryptions of the same value differ.
	assert.NotEqual(t, c1, c2)
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := make([]byte, 32)
	wrongKey := make([]byte, 32)
	wrongKey[0] = 0xFF

	ciphertext, err := crypto.Encrypt(key, "secret")
	require.NoError(t, err)

	_, err = crypto.Decrypt(wrongKey, ciphertext)
	require.Error(t, err)
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	ciphertext, err := crypto.Encrypt(key, "secret")
	require.NoError(t, err)

	ciphertext[len(ciphertext)-1] ^= 0xFF // flip last byte
	_, err = crypto.Decrypt(key, ciphertext)
	require.Error(t, err)
}

func TestEncrypt_InvalidKeySize(t *testing.T) {
	_, err := crypto.Encrypt([]byte("tooshort"), "value")
	require.Error(t, err)
}

func TestKeyFromHex(t *testing.T) {
	// 32 bytes = 64 hex chars
	hex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	key, err := crypto.KeyFromHex(hex)
	require.NoError(t, err)
	assert.Len(t, key, 32)

	_, err = crypto.KeyFromHex("tooshort")
	require.Error(t, err)

	_, err = crypto.KeyFromHex("gg" + hex[2:]) // invalid hex
	require.Error(t, err)
}
