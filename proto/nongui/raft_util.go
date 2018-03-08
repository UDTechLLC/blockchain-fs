package nongui

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type FileRaftValue struct {
	Filename  string
	TimeStamp time.Time
	ShaKey    string
	CpkIndex  string
}

func EcdsaSignWithCSK(walletInfo *WalletCreateInfo, basename64 string) (string, error) {
	// TODO: check walletInfo and Keys
	if walletInfo == nil {
		return "", fmt.Errorf("Wallet Info is nil. We can't get Keys.")
	}
	if len(walletInfo.PrivKey) != 64 {
		return "", fmt.Errorf("Private Key is wrong!")
	}
	if len(walletInfo.PubKey) != 128 {
		return "", fmt.Errorf("Public Key is wrong!")
	}
	ECDSAKeyD := walletInfo.PrivKey
	ECDSAKeyX := walletInfo.PubKey[:64]
	ECDSAKeyY := walletInfo.PubKey[64:]

	keyD := new(big.Int)
	keyX := new(big.Int)
	keyY := new(big.Int)
	keyD.SetString(ECDSAKeyD, 16)
	keyX.SetString(ECDSAKeyX, 16)
	keyY.SetString(ECDSAKeyY, 16)

	//fmt.Println("basename64:", basename64)
	claims := &jwt.MapClaims{
		"basename64": basename64,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     keyX,
		Y:     keyY,
	}

	privateKey := ecdsa.PrivateKey{D: keyD, PublicKey: publicKey}
	signed64, err := token.SignedString(&privateKey)
	if err != nil {
		return "", err
	}
	//fmt.Println("signed64:", signed64)

	return signed64, nil
}

func EcdsaParseVerifyWithCPK(walletInfo *WalletCreateInfo, signed64 string) (string, error) {
	// TODO: check walletInfo and Keys
	if walletInfo == nil {
		return "", fmt.Errorf("Wallet Info is nil. We can't get Keys")
	}
	if len(walletInfo.PubKey) != 128 {
		return "", fmt.Errorf("Public Key is wrong!")
	}
	ECDSAKeyX := walletInfo.PubKey[:64]
	ECDSAKeyY := walletInfo.PubKey[64:]

	keyX := new(big.Int)
	keyY := new(big.Int)
	keyX.SetString(ECDSAKeyX, 16)
	keyY.SetString(ECDSAKeyY, 16)

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     keyX,
		Y:     keyY,
	}

	token, err := jwt.Parse(signed64, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return &publicKey, nil
	})
	// TODO: err != nil
	if err != nil {
		return "", err
	}

	var basename64 string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		basename64 = claims["basename64"].(string)
	} else {
		// TODO: not ok?
	}
	return basename64, nil
}

func GetZeroIndex(walletInfo *WalletCreateInfo, raftApi *RaftApi) ([]byte, int64, error) {
	// get last CPKIndex = CPK + 0000000000000000 (8 bytes)
	// prepare key for Get
	cpkIndex0 := []byte(walletInfo.PubKey)
	index0 := make([]byte, 8)
	binary.LittleEndian.PutUint64(index0, uint64(0))
	index0 = []byte(hex.EncodeToString(index0))
	//fmt.Printf("index0: %s\n", string(index0))

	cpkIndex0 = append(cpkIndex0, index0...)

	// Get last CPIIndex
	// TODO: try to use wallet.CpkZeroIndex instead of this
	cpkIndexLast, err := raftApi.GetKey(string(cpkIndex0))
	if err != nil {
		fmt.Printf("Try to get last CPKIndex was failed with error: %s\n", err.Error())
		return nil, -1, err
	}
	//fmt.Printf("cpkIndexLast: %s\n", cpkIndexLast)

	// casting string to int64
	var cpkIndexLastInt64 int64
	if cpkIndexLast == "" {
		// if last CPKIndex is not existing, just set it to 0
		cpkIndexLastInt64 = int64(0)
	} else {
		cpkIndexLastDecode, err := hex.DecodeString(cpkIndexLast)
		if err != nil {
			fmt.Printf("Try to decode last CPKIndex was failed with error: %s\n", err.Error())
			return nil, -1, err
		}
		cpkIndexLastInt64 = int64(binary.LittleEndian.Uint64(cpkIndexLastDecode))
	}
	//fmt.Printf("cpkIndexLastInt64: %d\n", cpkIndexLastInt64)

	return cpkIndex0, cpkIndexLastInt64, nil
}

func GetFileIndex(index int64, walletInfo *WalletCreateInfo, raftApi *RaftApi) (fileRaft *FileRaftValue, err error) {
	cpkIndex := []byte(walletInfo.PubKey)

	cpkIndexNew := make([]byte, 8)
	binary.LittleEndian.PutUint64(cpkIndexNew, uint64(index))
	cpkIndexNew = []byte(hex.EncodeToString(cpkIndexNew))
	//fmt.Printf("cpkIndexNew: %s\n", string(cpkIndexNew))

	cpkIndex = append(cpkIndex, cpkIndexNew...)
	//fmt.Printf("cpkIndex: %s\n", string(cpkIndex))

	shaKeyString, err := raftApi.GetKey(string(cpkIndex))
	if err != nil {
		// TODO:
		fmt.Println("Error when getting SHA256:", err)
		return nil, err
	}
	// CPKIndex is absent
	if shaKeyString == "" {
		//fmt.Printf("Skip this index: %d\n", index)
		return nil, fmt.Errorf("Skip this index!")
	}
	//fmt.Printf("shaKeyString: %s\n", shaKeyString)

	value, err := raftApi.GetKey(shaKeyString)
	if err != nil {
		// TODO:
		fmt.Println("Error when getting FileInfo:", err)
		return nil, err
	}
	//fmt.Printf("value: %s\n", value)

	// CPK
	cpkTest := string(value[0:128])
	if cpkTest != walletInfo.PubKey {
		// TODO:
		fmt.Println("CPK was not matched!")
		return nil, fmt.Errorf("CPK was not matched!")
	}

	// Info (Base64(File.Basename) + Timestamp)
	info := value[128:]
	infoLen := len(info)

	// TODO: Base64 signed with CSK - Parse & Verify with CPK
	signed64 := info[0 : infoLen-16]
	//fmt.Printf("signed64: %s\n", signed64)
	basename64, err := EcdsaParseVerifyWithCPK(walletInfo, signed64)
	if err != nil {
		// if we got error then we don't add this file to list
		return nil, err
	}
	//fmt.Printf("basename64: %s\n", basename64)

	fileBasename, err := base64.RawURLEncoding.DecodeString(string(basename64))
	if err != nil {
		// if we got error then we don't add this file to list
		return nil, err
	}
	//fmt.Println("Filename:", string(fileBasename))

	timeStamp := info[infoLen-16:]
	timeStampDecode, err := hex.DecodeString(timeStamp)
	if err != nil {
		// if we got error then we don't add this file to list
		return nil, err
	}
	timeStampInt64 := int64(binary.LittleEndian.Uint64(timeStampDecode))
	timeStampTime := time.Unix(timeStampInt64, 0)
	//fmt.Println("Timestamp: ", timeStampTime.Format(time.RFC1123))

	fileRaft = &FileRaftValue{
		Filename:  string(fileBasename),
		TimeStamp: timeStampTime,
		ShaKey:    shaKeyString,
		CpkIndex:  string(cpkIndex),
	}

	return fileRaft, nil
}

func SaveFileToRaft(file string, walletInfo *WalletCreateInfo, raftApi *RaftApi) error {
	basename := filepath.Base(file)
	//fmt.Printf("basename: %s\n", basename)
	fi, err := os.Stat(file)
	if err != nil || fi == nil {
		fmt.Printf("os.Stat error: %s\n", err.Error())
		return err
	}

	// Key = SHA256 [ Base64(File.Basename) + File.Size + Timestamp ]
	key := []byte{}

	// Base64(File.Basename)
	basename64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(basename)))
	base64.RawURLEncoding.Encode(basename64, []byte(basename))
	//fmt.Printf("basename64: %s\n", string(basename64))
	key = append(key, basename64...)

	// File.Size, int64 to []byte
	fileSize := make([]byte, 8)
	binary.LittleEndian.PutUint64(fileSize, uint64(fi.Size()))
	fileSize = []byte(hex.EncodeToString(fileSize))
	//fmt.Printf("fi.Size(): %d\n", fi.Size())
	//fmt.Printf("fileSize: %s\n", string(fileSize))
	key = append(key, fileSize...)

	// Timestamp
	timeStamp := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeStamp, uint64(time.Now().Unix()))
	timeStamp = []byte(hex.EncodeToString(timeStamp))
	//fmt.Printf("timeStamp: %s\n", string(timeStamp))
	key = append(key, timeStamp...)

	//fmt.Printf("key: %s\n", string(key))

	shaKey := sha256.Sum256(key)
	////shaKey2 := sha256.Sum256([]byte(string(key)))
	//fmt.Printf("shaKey: %x\n", shaKey[:])
	shaKeyString := hex.EncodeToString(shaKey[:])

	// Value = CPK + Base64(File.Basename) + Timestamp
	value := []byte(walletInfo.PubKey) // CPK

	// TODO: Base64 signed with CSK
	//fmt.Printf("basename64: %s\n", string(basename64))
	signed64, err := EcdsaSignWithCSK(walletInfo, string(basename64))
	if err != nil {
		// if we got error then we don't save this file to Raft
		return err
	}
	//fmt.Printf("signed64: %s\n", signed64)
	//basename64test, err := EcdsaParseVerifyWithCPK(walletInfo, signed64)
	//if err != nil {
	//	// if we got error then we don't save this file to Raft
	//	return err
	//}
	//fmt.Printf("basename64test: %s\n", basename64test)

	value = append(value, signed64...)  // Base64
	value = append(value, timeStamp...) // Timestamp
	//fmt.Printf("value: %s\n", string(value))

	////

	// CPK Index
	cpkIndex := []byte(walletInfo.PubKey)

	// CHECKIT:
	cpkIndex0, cpkIndexLastInt64, err := GetZeroIndex(walletInfo, raftApi)
	if err != nil {
		return err
	}

	// create new CPKIndex
	cpkIndexNew := make([]byte, 8)
	binary.LittleEndian.PutUint64(cpkIndexNew, uint64(cpkIndexLastInt64+1))
	cpkIndexNew = []byte(hex.EncodeToString(cpkIndexNew))
	//fmt.Printf("cpkIndexNew: %s\n", string(cpkIndexNew))

	cpkIndex = append(cpkIndex, cpkIndexNew...)
	//fmt.Printf("cpkIndex: %s\n", string(cpkIndex))

	// Main Key/Value Store
	raftApi.SetKey(shaKeyString, string(value))

	// CPKIndex Key/Value Store
	raftApi.SetKey(string(cpkIndex), shaKeyString)
	raftApi.SetKey(string(cpkIndex0), string(cpkIndexNew))

	// TODO: save last cpkIndex to wallet
	walletInfo.CpkZeroIndex = string(cpkIndexNew)

	return nil
}

func RemoveFileFromRaft(fileRaft *FileRaftValue, walletInfo *WalletCreateInfo, raftApi *RaftApi) {
	fmt.Println("shaKey:", fileRaft.ShaKey)
	fmt.Println("cpkIndex:", fileRaft.CpkIndex)

	raftApi.DeleteKey(fileRaft.ShaKey)
	raftApi.DeleteKey(fileRaft.CpkIndex)

	// TODO: we can fix cpkIndex0 for last cpkIndex?
	lastIndex := walletInfo.PubKey + walletInfo.CpkZeroIndex
	if fileRaft.CpkIndex == lastIndex {
		// we can update CpkZeroIndex here!
		fmt.Println("We can update CpkZeroIndex here")
	}

	// TODO: or we can rebuild index, yeah!
	// TODO: or we can save file count with cpkIndex!?
	// so cpkIndex will have 2 values: last cpkIndex and file count
}
