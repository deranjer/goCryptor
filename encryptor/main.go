package encryptor

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/scrypt"
)

var nonceLength = 12

// EncryptFile takes in a password and a filepath and encrypts a file
func EncryptFile(password, inputFile string) error {
	// Generating salt from random reader
	salt := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return errors.New("salt generation failed: " + err.Error())
	}
	// use the scrypt library to generate a 32 bit key for the AES cipher
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return errors.New("Unable to create key from password: " + err.Error())
	}
	// read in the input file to convert to []byte
	inputBytes, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return errors.New("block error: " + err.Error())
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.New("cipher error: " + err.Error())
	}
	// creating a nonce that will be stored with the encryption
	nonce := make([]byte, nonceLength)
	// reading random data into the nonce
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return errors.New("random data read error: " + err.Error())
	}
	// get the file extension of the original file
	fileExt := filepath.Ext(inputFile)
	// convert the file ext to bytes
	fileExtBytes := []byte(fileExt)
	// Any extension over 10 char WILL ERROR
	if len(fileExtBytes) > 10 {
		return errors.New("file ext cannot be longer than 10 chars")
	}
	extBytes := make([]byte, 10)
	// Writing the ext to the byte array, will trim any excess on decrypt to write out file
	for i, byte := range fileExtBytes {
		extBytes[i] = byte
	}
	// joining the salt and extension data into a single array
	saltExt := append(salt, extBytes...)
	// adding the salt and ext data to the nonce array so all in single byte array
	metaData := append(nonce, saltExt...)
	// the metadata is prepended to the front of the encryption, to be extracted during decryption
	ciphertext := aesgcm.Seal(metaData, nonce, inputBytes, nil)
	// create a new file name, then write it with the .gcx extension
	newFileName := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
	err = ioutil.WriteFile(newFileName+fileExt+".gcx", ciphertext, 0644)
	if err != nil {
		return errors.New("Error writing file: " + err.Error())
	}
	return nil
}

// DecryptFile takes in a password and file path and decrypts that file
func DecryptFile(password, encryptedFile string, overwrite bool) error {
	// read in the input file to convert to []byte
	fileBytes, err := ioutil.ReadFile(encryptedFile)
	if err != nil {
		return errors.New("read file err: " + err.Error())
	}
	// separate the metadata from the ciphertext
	metaData, ciphertext := fileBytes[:54], fileBytes[54:]
	// extract the nonce, salt and file extension from the metadata
	nonce, salt, fileExt := metaData[:12], metaData[12:44], metaData[44:54]
	// convert the password into a key using the extracted salt
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return errors.New("Unable to create key from password: " + err.Error())
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	// perform the decryption
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}
	// strip the excess from the EXT in bytes to get a valid extension
	fileExt = bytes.Trim(fileExt, "\000")
	// remove the .gcx file ext from the encrypted file name
	newFileNameFull := strings.TrimSuffix(encryptedFile, ".gcx")
	// remove the old extension
	newFileName := strings.TrimSuffix(newFileNameFull, string(fileExt))
	// check if file exists and if we shouldn't overwrite then add decrypt to the file name
	if !overwrite {
		fullFileName := newFileName + string(fileExt)
		_, err := os.Stat(fullFileName)
		if os.IsNotExist(err) {
			err = ioutil.WriteFile(newFileName+string(fileExt), plaintext, 0644)
			if err != nil {
				return errors.New("Error writing plaintext file: " + err.Error())
			}
			return nil
		}
		err = ioutil.WriteFile(newFileName+"-decrypt"+string(fileExt), plaintext, 0644)
		if err != nil {
			return errors.New("Error writing plaintext (decrypt string) file: " + err.Error())
		}
		return nil
	}
	// if we can overwrite, then write out the decrypted file
	err = ioutil.WriteFile(newFileName+string(fileExt), plaintext, 0644)
	if err != nil {
		return errors.New("Error writing plaintext file: " + err.Error())
	}
	return nil
}
