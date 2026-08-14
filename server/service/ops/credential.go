package ops

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/crypto"
	"gorm.io/gorm"
)

// CredentialService 凭据库 CRUD, Secret / Passphrase 加密存储。
type CredentialService struct{}

func (s *CredentialService) CreateCredential(ctx context.Context, in opsReq.CredentialInput) (*ops.Credential, error) {
	c := ops.Credential{
		Name:     in.Name,
		Type:     in.Type,
		Username: in.Username,
		Remark:   in.Remark,
	}
	if c.Type == "" {
		c.Type = "password"
	}
	if in.Secret != "" {
		enc, err := crypto.Encrypt(in.Secret)
		if err != nil {
			return nil, err
		}
		c.Secret = enc
	}
	if in.Passphrase != "" {
		enc, err := crypto.Encrypt(in.Passphrase)
		if err != nil {
			return nil, err
		}
		c.Passphrase = enc
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateCredential 更新凭据; Secret / Passphrase 留空表示不改, 避免覆盖历史密文。
func (s *CredentialService) UpdateCredential(ctx context.Context, id uint, in opsReq.CredentialInput) error {
	if id == 0 {
		return errors.New("凭据 id 不能为空")
	}
	updates := map[string]interface{}{
		"name":     in.Name,
		"type":     in.Type,
		"username": in.Username,
		"remark":   in.Remark,
	}
	if in.Type == "" {
		updates["type"] = "password"
	}
	if in.Secret != "" {
		enc, err := crypto.Encrypt(in.Secret)
		if err != nil {
			return err
		}
		updates["secret"] = enc
	}
	if in.Passphrase != "" {
		enc, err := crypto.Encrypt(in.Passphrase)
		if err != nil {
			return err
		}
		updates["passphrase"] = enc
	}
	return global.GVA_DB.WithContext(ctx).Model(&ops.Credential{}).Where("id = ?", id).Updates(updates).Error
}

func (s *CredentialService) DeleteCredential(ctx context.Context, id uint) error {
	return global.GVA_DB.WithContext(ctx).Delete(&ops.Credential{}, id).Error
}

func (s *CredentialService) GetCredentialList(ctx context.Context, info opsReq.CredentialSearch) (list []ops.Credential, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.Credential{})
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Type != "" {
		db = db.Where("type = ?", info.Type)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	limit, offset := info.LimitOffset()
	if limit == 0 {
		limit = 10
	}
	err = db.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return
}

// GetCredentialByID 取单条凭据(含密文, 供跳板机内部解密使用, 不对外暴露)。
func (s *CredentialService) GetCredentialByID(ctx context.Context, id uint) (*ops.Credential, error) {
	var c ops.Credential
	if err := global.GVA_DB.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("凭据不存在")
		}
		return nil, err
	}
	return &c, nil
}

// DecryptSecret 解密凭据密钥内容(密码 / 私钥 / 口令), 仅供跳板机 SSH 连接内部使用。
func (s *CredentialService) DecryptSecret(c *ops.Credential) (secret, passphrase string, err error) {
	secret, err = crypto.Decrypt(c.Secret)
	if err != nil {
		return
	}
	passphrase, err = crypto.Decrypt(c.Passphrase)
	return
}
