package workflow

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// 日志流来源
const (
	LogStreamStdout = "stdout" // shell 标准输出 / HTTP 响应体
	LogStreamStderr = "stderr" // shell 标准错误 / HTTP 错误信息
	LogStreamSystem = "system" // 引擎自身的系统日志(状态变更、异常等)
)

// PipelineBuildLog 日志行(独立表, 避免单行膨胀; 支持增量推流与分页拉取)
// 按 (BuildID, StepID, Seq) 定位, Seq 在同一 Step 内自增。
// StepID 指向 PipelineBuildStep.ID(运行实例), 而非定义表。
type PipelineBuildLog struct {
	global.GVA_MODEL
	BuildID uint      `json:"buildId" form:"buildId" gorm:"index:idx_build_step_seq,priority:1;column:build_id;comment:所属构建ID"`
	StepID  uint      `json:"stepId" form:"stepId" gorm:"index:idx_build_step_seq,priority:2;column:step_id;comment:所属步骤运行ID(PipelineBuildStep.ID)"`
	Seq     int       `json:"seq" form:"seq" gorm:"index:idx_build_step_seq,priority:3;column:seq;comment:同 Step 内日志序号"`
	Stream  string    `json:"stream" form:"stream" gorm:"column:stream;comment:日志来源 stdout|stderr|system"`
	Text    string    `json:"text" form:"text" gorm:"type:text;column:text;comment:日志正文"`
	Ts      time.Time `json:"ts" form:"ts" gorm:"column:ts;comment:日志产生时间"`
}

func (PipelineBuildLog) TableName() string {
	return "wf_pipeline_build_logs"
}
