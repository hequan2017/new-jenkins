package ops

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// AssetGroup 资产分组/环境(如 prod / staging / dev), 用于按环境批量筛选与组织。
type AssetGroup struct {
	global.GVA_MODEL
	Name string `json:"name" form:"name" gorm:"index;column:name;comment:分组名称"`
	Env  string `json:"env" form:"env" gorm:"column:env;comment:环境标识 prod|staging|dev"`
	Sort int    `json:"sort" form:"sort" gorm:"column:sort;comment:排序"`
	Desc string `json:"desc" form:"desc" gorm:"column:desc;comment:描述"`
}

func (AssetGroup) TableName() string { return "ops_asset_groups" }
