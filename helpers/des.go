package helpers

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
)

// des加密
func DesEncrypt(origData, key []byte) (string, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	encodeString := base64.StdEncoding.EncodeToString(crypted)
	return encodeString, nil
}

// des解密
func DesDecrypt(base64Crypted string, key []byte) (string, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	crypted, err := base64.StdEncoding.DecodeString(base64Crypted)
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return string(origData[:]), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
