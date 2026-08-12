package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

// BuildDetail 构建详情聚合: build 本体 + 阶段运行视图 + 步骤运行视图
// 前端详情页用它在顶部渲染 Stage 横条, 在下方渲染选中 Step 的日志。
type BuildDetail struct {
	Build  workflow.PipelineBuild       `json:"build"`
	Stages []workflow.PipelineBuildStage `json:"stages"`
	Steps  []workflow.PipelineBuildStep  `json:"steps"`
}
