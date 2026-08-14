package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// 资产状态
const (
	AssetStatusOnline      = "online"      // 在线
	AssetStatusOffline     = "offline"     // 离线
	AssetStatusMaintenance = "maintenance" // 维护中
)

// Asset 资产(主机台账)
type Asset struct {
	global.GVA_MODEL
	Name   string         `json:"name" form:"name" gorm:"index;column:name;comment:资产名称"`
	Host   string         `json:"host" form:"host" gorm:"column:host;comment:主机地址(IP/域名)"`
	Port   int            `json:"port" form:"port" gorm:"column:port;comment:SSH端口;default:22"`
	OS     string         `json:"os" form:"os" gorm:"column:os;comment:操作系统"`
	Tags   datatypes.JSON `json:"tags" form:"tags" gorm:"column:tags;comment:标签(JSON 数组)" swaggertype:"array,string"`
	Status string         `json:"status" form:"status" gorm:"column:status;comment:状态 online|offline|maintenance;default:offline"`
	Remark string         `json:"remark" form:"remark" gorm:"column:remark;comment:备注"`
}

func (Asset) TableName() string { return "ops_assets" }

// Credential 凭据库(SSH 用户名 / 密码 / 私钥)
// Secret 字段使用 AES-GCM 加密存储, 列表 / 详情接口永不返回明文。
type Credential struct {
	global.GVA_MODEL
	Name     string `json:"name" form:"name" gorm:"index;column:name;comment:凭据名称"`
	Type     string `json:"type" form:"type" gorm:"column:type;comment:凭据类型 password|key;default:password"`
	Username string `json:"username" form:"username" gorm:"column:username;comment:登录用户名"`
	// Secret 密文: password 时为加密后的密码, key 时为加密后的 PEM 私钥(可附带 passphrase)
	// 输出时一律置空, 仅内部解密使用, 见 request 中的明文字段
	Secret string `json:"-" form:"-" gorm:"column:secret;comment:加密凭据"`
	// Passphrase 私钥口令密文(仅 type=key 时使用)
	Passphrase string `json:"-" form:"-" gorm:"column:passphrase;comment:加密私钥口令"`
	Remark     string `json:"remark" form:"remark" gorm:"column:remark;comment:备注"`
}

func (Credential) TableName() string { return "ops_credentials" }
