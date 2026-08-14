package ops

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"golang.org/x/crypto/ssh"
)

// sshTimeout 取配置的 SSH 超时, 兜底 10s。
func sshTimeout() time.Duration {
	if global.GVA_CONFIG.Ops.SSHTimeout > 0 {
		return global.GVA_CONFIG.Ops.SSHTimeout
	}
	return 10 * time.Second
}

// buildSSHClient 依据资产 + 凭据构造已认证的 SSH 客户端。
// 凭据密文在本方法内解密, 解出的明文不离开本函数。
func buildSSHClient(ctx context.Context, assetID, credentialID uint) (*ssh.Client, error) {
	asset, err := (&AssetService{}).GetAssetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}
	cred, err := (&CredentialService{}).GetCredentialByID(ctx, credentialID)
	if err != nil {
		return nil, err
	}
	secret, passphrase, err := (&CredentialService{}).DecryptSecret(cred)
	if err != nil {
		return nil, fmt.Errorf("凭据解密失败: %w", err)
	}

	var authMethod ssh.AuthMethod
	switch cred.Type {
	case "key":
		var signer ssh.Signer
		var sErr error
		if passphrase != "" {
			signer, sErr = ssh.ParsePrivateKeyWithPassphrase([]byte(secret), []byte(passphrase))
		} else {
			signer, sErr = ssh.ParsePrivateKey([]byte(secret))
		}
		if sErr != nil {
			return nil, fmt.Errorf("私钥解析失败: %w", sErr)
		}
		authMethod = ssh.PublicKeys(signer)
	default: // password
		authMethod = ssh.Password(secret)
	}

	config := &ssh.ClientConfig{
		User:            cred.Username,
		Auth:            []ssh.AuthMethod{authMethod},
		Timeout:         sshTimeout(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 跳板机演示场景: 跳过主机指纹校验
	}
	addr := strings.TrimSpace(asset.Host)
	if asset.Port == 0 {
		asset.Port = 22
	}
	target := fmt.Sprintf("%s:%d", addr, asset.Port)
	// 用超时控制拨号; context 取消由后续握手/命令环节承接
	conn, err := net.DialTimeout("tcp", target, sshTimeout())
	if err != nil {
		return nil, fmt.Errorf("连接主机失败: %w", err)
	}
	// 把 context 取消传导到底层连接
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	c, chans, reqs, err := ssh.NewClientConn(conn, target, config)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("SSH 握手失败: %w", err)
	}
	return ssh.NewClient(c, chans, reqs), nil
}

// SSHService 跳板机 SSH 能力: 连接测试 + 单条命令执行。
type SSHService struct{}

// TestConnection 测试资产 + 凭据能否建立 SSH 连接。
func (s *SSHService) TestConnection(ctx context.Context, assetID, credentialID uint) error {
	client, err := buildSSHClient(ctx, assetID, credentialID)
	if err != nil {
		return err
	}
	defer client.Close()
	// 建一个 session 探活, 确认握手后通道可用
	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("建立会话失败: %w", err)
	}
	defer sess.Close()
	return nil
}

// ExecCommand 在目标主机执行单条命令, 返回合并后的 stdout+stderr 输出。
func (s *SSHService) ExecCommand(ctx context.Context, assetID, credentialID uint, command string) (string, error) {
	if strings.TrimSpace(command) == "" {
		return "", errors.New("命令不能为空")
	}
	client, err := buildSSHClient(ctx, assetID, credentialID)
	if err != nil {
		return "", err
	}
	defer client.Close()
	sess, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("建立会话失败: %w", err)
	}
	defer sess.Close()

	var buf bytes.Buffer
	sess.Stdout = &buf
	sess.Stderr = &buf
	done := make(chan error, 1)
	go func() { done <- sess.Run(command) }()
	select {
	case <-ctx.Done():
		_ = sess.Signal(ssh.SIGKILL)
		return buf.String(), ctx.Err()
	case err := <-done:
		return buf.String(), err
	}
}
