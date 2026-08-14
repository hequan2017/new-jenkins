// Package crypto 提供 AES-GCM 对称加密能力, 用于运维凭据(SSH 密码 / 私钥)的加密存储。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// padKey 把任意长度密钥规整为 AES 合法长度(16/24/32 字节)。
// 不足则右侧补 0, 超长则截断到 32 字节, 默认走 AES-256。
func padKey(key string) []byte {
	b := []byte(key)
	switch {
	case len(b) >= 32:
		return b[:32]
	case len(b) >= 24:
		return b[:24]
	case len(b) >= 16:
		return b[:16]
	default:
		padded := make([]byte, 16)
		copy(padded, b)
		return padded
	}
}

func newGCM() (cipher.AEAD, error) {
	aesKey := ""
	if global.GVA_CONFIG.Ops.AESKey != "" {
		aesKey = global.GVA_CONFIG.Ops.AESKey
	}
	block, err := aes.NewCipher(padKey(aesKey))
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}
	return cipher.NewGCM(block)
}

// Encrypt 加密明文, 返回 base64 编码的密文(nonce + ciphertext 拼接后编码)。
func Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成 nonce 失败: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt 解密 Encrypt 产出的 base64 密文。
func Decrypt(cipherText string) (string, error) {
	if cipherText == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("密文 base64 解码失败: %w", err)
	}
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("密文长度不足")
	}
	nonce, body := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}
	return string(plain), nil
}
