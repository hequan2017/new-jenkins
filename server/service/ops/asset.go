package ops

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"gorm.io/gorm"
)

// AssetService 资产(主机台账)CRUD
type AssetService struct{}

func (s *AssetService) CreateAsset(ctx context.Context, a *ops.Asset) error {
	if a.Port == 0 {
		a.Port = 22
	}
	if a.Status == "" {
		a.Status = ops.AssetStatusOffline
	}
	return global.GVA_DB.WithContext(ctx).Create(a).Error
}

func (s *AssetService) UpdateAsset(ctx context.Context, a *ops.Asset) error {
	if a.ID == 0 {
		return errors.New("资产 id 不能为空")
	}
	if a.Port == 0 {
		a.Port = 22
	}
	return global.GVA_DB.WithContext(ctx).Save(a).Error
}

func (s *AssetService) DeleteAsset(ctx context.Context, id uint) error {
	return global.GVA_DB.WithContext(ctx).Delete(&ops.Asset{}, id).Error
}

func (s *AssetService) GetAssetList(ctx context.Context, info opsReq.AssetSearch) (list []ops.Asset, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.Asset{})
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
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

// GetAssetByID 取单条资产(供跳板机 / 命令执行复用)
func (s *AssetService) GetAssetByID(ctx context.Context, id uint) (*ops.Asset, error) {
	var a ops.Asset
	if err := global.GVA_DB.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("资产不存在")
		}
		return nil, err
	}
	return &a, nil
}
