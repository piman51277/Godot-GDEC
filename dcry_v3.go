package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func generateKey(password string) []byte {
	h := md5.New()
	io.WriteString(h, password)
	hash := h.Sum(nil)
	str := hex.EncodeToString(hash)
	return []byte(str)
}

func decryptFile(inputFile, outputFile, password string) {
	file, err := os.Open(inputFile)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outFile.Close()

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(outFile)

	header := make([]byte, 4)
	reader.Read(header)

	if string(header) != "GDEC" {
		fmt.Println("Incorrect file format")
		return
	}

	fileModeBytes := make([]byte, 4)
	reader.Read(fileModeBytes)

	if binary.LittleEndian.Uint32(fileModeBytes) != 1 {
		fmt.Println("Incorrect file format. Is this from Godot v3 or before?")
		return
	}

	fileHash := make([]byte, 16)
	reader.Read(fileHash)

	fmt.Print("Plaintext MD5 Hash: ")
	for _, v := range fileHash {
		fmt.Printf("%02x", v)
	}
	fmt.Println("")

	dataSizeBytes := make([]byte, 8)
	reader.Read(dataSizeBytes)

	dataSize := binary.LittleEndian.Uint64(dataSizeBytes)
	fmt.Printf("Data size: %d\n", dataSize)

	//Round to next multiple of 16
	blockCount := (dataSize + 15) >> 4
	fmt.Printf("Block Count: %d\n", blockCount)

	padding := blockCount*16 - dataSize

	key := generateKey(password)
	cipher, _ := aes.NewCipher(key)
	hash := md5.New()

	fmt.Print("Using key: ")
	for _, v := range key {
		fmt.Printf("%02x", v)
	}
	fmt.Println("")

	blockBytes := make([]byte, 16)
	dcryBytes := make([]byte, 16)
	for i := uint64(0); i < blockCount-1; i++ {
		reader.Read(blockBytes)
		cipher.Decrypt(dcryBytes, blockBytes)
		writer.Write(dcryBytes)
		hash.Write(dcryBytes)
	}

	//decrypt the last block
	reader.Read(blockBytes)
	cipher.Decrypt(dcryBytes, blockBytes)
	writer.Write(dcryBytes[:16-padding])
	hash.Write(dcryBytes[:16-padding])

	writer.Flush()

	finalHash := hash.Sum(nil)

	fmt.Print("Decrypted MD5 Hash: ")
	for _, v := range finalHash {
		fmt.Printf("%02x", v)
	}
	fmt.Println("")

	if bytes.Equal(fileHash, finalHash) {
		fmt.Println("Decryption succesful!")
	} else {
		fmt.Println("Decryption failed! Is your password correct?")
	}
}

func encryptFile(inputFile, outputFile, password string) {
	file, err := os.Open(inputFile)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outFile.Close()

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(outFile)

	//header
	writer.WriteString("GDEC")

	//file mode
	fileModeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(fileModeBytes, 1)
	writer.Write(fileModeBytes)

	//hash the input data
	h := md5.New()
	io.Copy(h, reader)
	hashBytes := h.Sum(nil)
	writer.Write(hashBytes)

	fmt.Print("Plaintext MD5 Hash: ")
	for _, v := range hashBytes {
		fmt.Printf("%02x", v)
	}
	fmt.Println("")

	//reset the input buffer
	file.Seek(0, 0)

	//file size
	fstat, _ := file.Stat()
	dataSize := uint64(fstat.Size())
	dataSizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataSizeBytes, dataSize)
	writer.Write(dataSizeBytes)
	fmt.Printf("Data size: %d\n", dataSize)

	//Round to next multiple of 16
	blockCount := (dataSize + 15) >> 4
	fmt.Printf("Block Count: %d\n", blockCount)

	//encrypt
	key := generateKey(password)
	cipher, _ := aes.NewCipher(key)
	blockBytes := make([]byte, 16)
	dcryBytes := make([]byte, 16)
	for i := uint64(0); i < blockCount-1; i++ {
		reader.Read(blockBytes)
		cipher.Encrypt(dcryBytes, blockBytes)
		writer.Write(dcryBytes)
	}
	//encrypt the last block
	blockBytes = make([]byte, 16)
	reader.Read(blockBytes)
	cipher.Encrypt(dcryBytes, blockBytes)
	writer.Write(dcryBytes)
	writer.Flush()

	fmt.Println("Encryption finished!")
}

func main() {
	encryptFile("samples/plaintext.txt", "samples/ciphertext.save", "piman51277")
	decryptFile("samples/ciphertext.save", "samples/plaintext-new.txt", "piman51277")
}
