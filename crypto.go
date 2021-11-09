// Cryptographic functions
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"math/big"
	"os"
	"unsafe"

	crand "crypto/rand"
)

const KEY_SIZE = 32
const SALT_SIZE = 64
const KEY_N_ITER = 4096
const HMAC_SHA512_SIZE = 64
const MAGIC_HEADER = 0xcafebabe

// Generate random bytes of the given length
func generateRandomBytes(size int) (error, []byte) {
	var data []byte

	data = make([]byte, size)

	_, err := crand.Read(data)

	if err != nil {
		fmt.Printf("Error generating random data - \"%s\"\n", err.Error())
		return err, data
	}

	return nil, data
}

// Generate a key from the given passphrase and (optional) salt
// If 2nd argument is nil, salt will be generated.
func generateKey(passPhrase string, oldSalt *[]byte) (error, []byte, []byte) {

	var salt []byte
	var key []byte
	var err error

	if oldSalt == nil {
		err, salt = generateRandomBytes(SALT_SIZE)
	} else {
		salt = *oldSalt
	}

	if err != nil {
		return err, key, salt
	}
	if len(salt) == 0 {
		return errors.New("invalid salt"), key, salt
	}

	key = pbkdf2.Key([]byte(passPhrase), salt, KEY_N_ITER, KEY_SIZE, sha512.New)
	return nil, key, salt
}

// Return if file is encrypted by looking at the magic header
func isFileEncrypted(encDbPath string) (error, bool) {

	var magicBytes string
	var header []byte
	var err error
	var fh *os.File

	fh, err = os.Open(encDbPath)
	if err != nil {
		return fmt.Errorf("Error - Can't read database -\"%s\"\n", err.Error()), false
	}

	defer fh.Close()

	// Read the header
	magicBytes = fmt.Sprintf("%x", MAGIC_HEADER)
	header = make([]byte, unsafe.Sizeof(MAGIC_HEADER))

	_, err = io.ReadFull(fh, header[:])
	if err != nil {
		return fmt.Errorf("Error - Can't read file header -\"%s\"\n", err.Error()), false
	}

	if string(header) != magicBytes {
		return fmt.Errorf("Not an encrypted database - invalid magic number"), false
	}

	return nil, true
}

// Encrypt the database path using AES
func encryptFile(dbPath string, password string) error {

	var err error
	var key []byte
	var salt []byte
	var nonce []byte
	var plainText []byte
	var cipherText []byte
	var magicBytes []byte
	var encText []byte
	var encDbPath string
	var hmacHash []byte

	plainText, err = os.ReadFile(dbPath)
	if err != nil {
		fmt.Printf("Error - Can't read database -\"%s\"\n", err)
		return err
	}

	err, key, salt = generateKey(password, nil)

	if err != nil {
		fmt.Printf("Error - Key derivation failed -\"%s\"\n", err)
		return err
	}

	//	fmt.Printf("\nsalt: %x\n", salt)
	//	fmt.Printf("key: %x\n", key)
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("Error - Cipher block creation failed - \"%s\"\n", err)
		return err
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		fmt.Printf("Error - AES GCM creation failed - \"%s\"\n", err)
		return err
	}

	err, nonce = generateRandomBytes(aesGCM.NonceSize())

	if err != nil {
		fmt.Printf("Error - Init vector generation failed -\"%s\"\n", err)
		return err
	}

	//	fmt.Printf("nonce: %x\n", nonce)
	magicBytes = []byte(fmt.Sprintf("%x", MAGIC_HEADER))
	cipherText = aesGCM.Seal(nonce, nonce, plainText, nil)

	// Calculate hmac signature and write it
	hCipher := hmac.New(sha512.New, key)
	hCipher.Write(cipherText)

	hmacHash = hCipher.Sum(nil)

	encText = append(magicBytes, salt...)
	encText = append(encText, hmacHash...)
	encText = append(encText, cipherText...)

	encDbPath = dbPath + ".varuh"

	err = os.WriteFile(encDbPath, encText, 0644)
	if err == nil {
		err = os.WriteFile(dbPath, encText, 0644)
		if err == nil {
			// Remove backup
			os.Remove(encDbPath)
		} else {
			fmt.Printf("Error writing encrypted database - \"%s\"\n", err.Error())
		}
	}
	//	fmt.Printf("%x\n", cipherText)

	return err
}

// Decrypt an already encrypted database file using given password
func decryptFile(encDbPath string, password string) error {

	var encText []byte
	var cipherText []byte
	var plainText []byte
	var key []byte
	var salt []byte
	var nonce []byte
	var hmacHash []byte
	var hmacSig []byte
	var origFile string

	var err error

	encText, err = os.ReadFile(encDbPath)
	if err != nil {
		fmt.Printf("Error - Can't read database -\"%s\"\n", err)
		return err
	}

	encText = encText[unsafe.Sizeof(MAGIC_HEADER):]
	// Read the old salt
	salt, encText = encText[:SALT_SIZE], encText[SALT_SIZE:]
	// Read the hmac hash checksum
	hmacHash, encText = encText[:HMAC_SHA512_SIZE], encText[HMAC_SHA512_SIZE:]

	err, key, _ = generateKey(password, &salt)

	if err != nil {
		fmt.Printf("Error - Key derivation failed -\"%s\"\n", err)
		return err
	}

	// verify the hmac
	// Calculate hmac signature and write it
	hCipher := hmac.New(sha512.New, key)
	hCipher.Write(encText)

	hmacSig = hCipher.Sum(nil)

	// Compare
	if !hmac.Equal(hmacSig, hmacHash) {
		fmt.Println("Invalid password or tampered data. Aborted")
		return errors.New("signature check failed")
	}
	//	fmt.Printf("\nsalt: %x\n", salt)
	//	fmt.Printf("key: %x\n", key)

	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("Error - Cipher block creation failed - \"%s\"\n", err)
		return err
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		fmt.Printf("Error - AES GCM creation failed - \"%s\"\n", err)
		return err
	}

	nonceSize := aesGCM.NonceSize()

	nonce, cipherText = encText[:nonceSize], encText[nonceSize:]
	//	fmt.Printf("nonce: %x\n", nonce)
	plainText, err = aesGCM.Open(nil, nonce, cipherText, nil)

	if err != nil {
		fmt.Printf("Error - Decryption failed - \"%s\"\n", err)
		return err
	}

	err = rewriteBaseFile(encDbPath, plainText)

	if err != nil {
		fmt.Printf("Error writing decrypted data to %s - \"%s\"\n", origFile, err.Error())
	}

	//	fmt.Printf("%s\n", string(plainText))
	return err
}

// Generate a random password - for adding listings
func generateRandomPassword(length int) (error, string) {

	var data []byte
	const source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789#!+$@~"

	data = make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(source))))
		if err != nil {
			return err, ""
		}

		data[i] = source[num.Int64()]
	}

	return nil, string(data)
}
