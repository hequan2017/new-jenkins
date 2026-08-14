package config

import "time"

// Ops 运维模块配置(资产管理 / 跳板机 / 工单发版)
type Ops struct {
	// AESKey 凭据(SSH 密码 / 私钥)对称加密密钥, 32 字节对应 AES-256。
	// 生产环境务必改成随机强密钥并妥善保管, 丢失则历史密文无法解密。
	AESKey string `mapstructure:"aes-key" json:"aesKey" yaml:"aes-key"`
	// SSHTimeout 跳板机 SSH 连接 / 命令执行超时
	SSHTimeout time.Duration `mapstructure:"ssh-timeout" json:"sshTimeout" yaml:"ssh-timeout"`
}
