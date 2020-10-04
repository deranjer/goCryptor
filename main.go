package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/scrypt"
)

var nonceLength = 12

func encryptFile(password, inputFile string) {
	// Generating salt from random reader
	salt := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		log.Fatal("salt generation failed: ", err)
	}
	// use the scrypt library to generate a 32 bit key for the AES cipher
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		fmt.Println("Unable to create key from password: ", err)
	}
	// read in the input file to convert to []byte
	inputBytes, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Println("unable to read in file: ", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	// creating a nonce that will be stored with the encryption
	nonce := make([]byte, nonceLength)
	// reading random data into the nonce
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		panic(err.Error())
	}
	// get the file extension of the original file
	fileExt := filepath.Ext(inputFile)
	// convert the file ext to bytes
	fileExtBytes := []byte(fileExt)
	// Any extension over 10 char WILL ERROR
	if len(fileExtBytes) > 10 {
		log.Fatal("file ext cannot be longer than 10 chars")
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
		fmt.Println("Error writing file; ", err)
	}
}

func decryptFile(password, encryptedFile string) {
	// read in the input file to convert to []byte
	fileBytes, err := ioutil.ReadFile(encryptedFile)
	if err != nil {
		fmt.Println("unable to read in file: ", err)
	}
	// separate the metadata from the ciphertext
	metaData, ciphertext := fileBytes[:54], fileBytes[54:]
	// extract the nonce, salt and file extension from the metadata
	nonce, salt, fileExt := metaData[:12], metaData[12:44], metaData[44:54]
	// convert the password into a key using the extracted salt
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		fmt.Println("Unable to create key from password: ", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	// perform the decryption
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	// strip the excess from the EXT in bytes to get a valid extension
	fileExt = bytes.Trim(fileExt, "\000")
	// remove the .gcx file ext from the encrypted file name
	newFileName := strings.TrimSuffix(encryptedFile, ".gcx")
	// write out the decrypted file
	err = ioutil.WriteFile(newFileName+"decrypt"+string(fileExt), plaintext, 0644)
	if err != nil {
		fmt.Println("Error writing file; ", err)
	}
}

func main() {
	//encryptFile("Password1", "test.pdf")
	decryptFile("Password1", "backup.gcx")
}
