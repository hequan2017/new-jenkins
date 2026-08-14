package ops

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
)

// AssetGroupService 资产分组/环境管理。
type AssetGroupService struct{}

func (s *AssetGroupService) Create(ctx context.Context, g *ops.AssetGroup) error {
	return global.GVA_DB.WithContext(ctx).Create(g).Error
}

func (s *AssetGroupService) Update(ctx context.Context, g *ops.AssetGroup) error {
	if g.ID == 0 {
		return errors.New("分组 id 不能为空")
	}
	return global.GVA_DB.WithContext(ctx).Save(g).Error
}

func (s *AssetGroupService) Delete(ctx context.Context, id uint) error {
	// 检查是否有资产引用
	var cnt int64
	global.GVA_DB.WithContext(ctx).Model(&ops.Asset{}).Where("group_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return errors.New("该分组下仍有资产, 无法删除")
	}
	return global.GVA_DB.WithContext(ctx).Delete(&ops.AssetGroup{}, id).Error
}

func (s *AssetGroupService) GetList(ctx context.Context, info opsReq.AssetGroupSearch) (list []ops.AssetGroup, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.AssetGroup{})
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Env != "" {
		db = db.Where("env = ?", info.Env)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	limit, offset := info.LimitOffset()
	if limit == 0 {
		limit = 50
	}
	err = db.Order("sort ASC, id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return
}

// GetAll 取全部分组(供下拉选择用)
func (s *AssetGroupService) GetAll(ctx context.Context) ([]ops.AssetGroup, error) {
	var list []ops.AssetGroup
	err := global.GVA_DB.WithContext(ctx).Order("sort ASC, id DESC").Find(&list).Error
	return list, err
}
