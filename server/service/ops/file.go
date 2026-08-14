package ops

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// FileEntry 远程目录条目
type FileEntry struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime string `json:"modTime"`
	Mode    string `json:"mode"`
}

// FileService 远程文件管理(SFTP): 列目录 / 读文件 / 写文件 / 删除 / 重命名。
// 复用 buildSSHClient 建立连接, 在其上开启 SFTP 子系统。
type FileService struct{}

// openSFTP 建立 SSH 连接并开启 SFTP 客户端, 返回 client(需调用方 Close) 与 sftp 客户端。
func openSFTP(ctx context.Context, assetID, credentialID uint) (*ssh.Client, *sftp.Client, error) {
	sshClient, err := buildSSHClient(ctx, assetID, credentialID)
	if err != nil {
		return nil, nil, err
	}
	sc, err := sftp.NewClient(sshClient)
	if err != nil {
		_ = sshClient.Close()
		return nil, nil, fmt.Errorf("开启 SFTP 失败: %w", err)
	}
	return sshClient, sc, nil
}

// ListDir 列出远程目录内容
func (s *FileService) ListDir(ctx context.Context, assetID, credentialID uint, dir string) ([]FileEntry, error) {
	if dir == "" {
		dir = "."
	}
	sshClient, sc, err := openSFTP(ctx, assetID, credentialID)
	if err != nil {
		return nil, err
	}
	defer sc.Close()
	defer sshClient.Close()

	infos, err := sc.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := make([]FileEntry, 0, len(infos))
	for _, fi := range infos {
		entries = append(entries, FileEntry{
			Name:    fi.Name(),
			Size:    fi.Size(),
			IsDir:   fi.IsDir(),
			ModTime: fi.ModTime().Format(time.RFC3339),
			Mode:    fi.Mode().String(),
		})
	}
	// 目录优先, 再按名称排序
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

// ReadFile 读取远程文件内容(限制大小, 避免拉超大文件)
const maxReadSize = 2 * 1024 * 1024 // 2MB

func (s *FileService) ReadFile(ctx context.Context, assetID, credentialID uint, fp string) (string, error) {
	sshClient, sc, err := openSFTP(ctx, assetID, credentialID)
	if err != nil {
		return "", err
	}
	defer sc.Close()
	defer sshClient.Close()

	r, err := sc.Open(fp)
	if err != nil {
		return "", err
	}
	defer r.Close()
	buf := bytes.NewBuffer(nil)
	if _, err := io.CopyN(buf, r, maxReadSize+1); err != nil && err != io.EOF {
		return "", err
	}
	if buf.Len() > maxReadSize {
		return buf.String()[:maxReadSize] + "\n...[截断]", nil
	}
	return buf.String(), nil
}

// WriteFile 写入/覆盖远程文件
func (s *FileService) WriteFile(ctx context.Context, assetID, credentialID uint, fp, content string) error {
	if fp == "" {
		return errors.New("文件路径不能为空")
	}
	sshClient, sc, err := openSFTP(ctx, assetID, credentialID)
	if err != nil {
		return err
	}
	defer sc.Close()
	defer sshClient.Close()

	// 确保父目录存在
	if mkErr := sc.MkdirAll(path.Dir(fp)); mkErr != nil {
		// 忽略已存在, 其它错误返回
		if !strings.Contains(mkErr.Error(), "exists") {
			// 部分服务器对已存在目录返回错误, 继续尝试写文件
		}
	}
	f, err := sc.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	return err
}

// Remove 删除远程文件或空目录
func (s *FileService) Remove(ctx context.Context, assetID, credentialID uint, fp string, isDir bool) error {
	sshClient, sc, err := openSFTP(ctx, assetID, credentialID)
	if err != nil {
		return err
	}
	defer sc.Close()
	defer sshClient.Close()
	if isDir {
		return sc.RemoveDirectory(fp)
	}
	return sc.Remove(fp)
}

// Rename 重命名 / 移动
func (s *FileService) Rename(ctx context.Context, assetID, credentialID uint, oldPath, newPath string) error {
	sshClient, sc, err := openSFTP(ctx, assetID, credentialID)
	if err != nil {
		return err
	}
	defer sc.Close()
	defer sshClient.Close()
	return sc.Rename(oldPath, newPath)
}

// Mkdir 新建目录
func (s *FileService) Mkdir(ctx context.Context, assetID, credentialID uint, dir string) error {
	sshClient, sc, err := openSFTP(ctx, assetID, credentialID)
	if err != nil {
		return err
	}
	defer sc.Close()
	defer sshClient.Close()
	return sc.Mkdir(dir)
}
