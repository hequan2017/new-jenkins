package ops

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// 备份任务状态
const (
	BackupStatusRunning  = "running"
	BackupStatusSuccess  = "success"
	BackupStatusFailed   = "failed"
	BackupStatusCanceled = "canceled"
)

// BackupTask 备份任务: 定期 SSH+SFTP 拉取目标文件并归档到本地, 记录历史支持下载恢复。
type BackupTask struct {
	global.GVA_MODEL
	Name         string `json:"name" form:"name" gorm:"index;column:name;comment:任务名称"`
	AssetID      uint   `json:"assetId" form:"assetId" gorm:"column:asset_id;comment:目标资产ID"`
	CredentialID uint   `json:"credentialId" form:"credentialId" gorm:"column:credential_id;comment:凭据ID"`
	RemotePath   string `json:"remotePath" form:"remotePath" gorm:"column:remote_path;comment:远程文件/目录路径"`
	Spec         string `json:"spec" form:"spec" gorm:"column:spec;comment:cron表达式"`
	WithSeconds  bool   `json:"withSeconds" form:"withSeconds" gorm:"column:with_seconds;comment:cron是否含秒位"`
	Enabled      bool   `json:"enabled" form:"enabled" gorm:"column:enabled;comment:是否启用;default:false"`
	KeepCount    int    `json:"keepCount" form:"keepCount" gorm:"column:keep_count;comment:保留历史份数;default:7"`
	Remark       string `json:"remark" form:"remark" gorm:"column:remark;comment:备注"`
}

func (BackupTask) TableName() string { return "ops_backup_tasks" }

// BackupRecord 备份执行记录(每次一份)
type BackupRecord struct {
	global.GVA_MODEL
	TaskID    uint   `json:"taskId" form:"taskId" gorm:"index;column:task_id;comment:任务ID"`
	Status    string `json:"status" form:"status" gorm:"index;column:status;comment:状态 running|success|failed"`
	LocalPath string `json:"localPath" form:"localPath" gorm:"column:local_path;comment:本地归档路径"`
	Size      int64  `json:"size" form:"size" gorm:"column:size;comment:归档大小(字节)"`
	Detail    string `json:"detail" form:"detail" gorm:"column:detail;type:text;comment:详情/错误信息"`
}

func (BackupRecord) TableName() string { return "ops_backup_records" }
