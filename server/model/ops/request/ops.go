package request

import (
	common "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// AssetSearch 资产分页查询
type AssetSearch struct {
	common.PageInfo
	Name   string `json:"name" form:"name"`
	Status string `json:"status" form:"status"`
}

// CredentialSearch 凭据分页查询
type CredentialSearch struct {
	common.PageInfo
	Name string `json:"name" form:"name"`
	Type string `json:"type" form:"type"`
}

// CredentialInput 凭据创建 / 更新入参(Secret / Passphrase 为明文, 仅写入时使用)
type CredentialInput struct {
	Name       string `json:"name" form:"name"`
	Type       string `json:"type" form:"type"`
	Username   string `json:"username" form:"username"`
	Secret     string `json:"secret" form:"secret"`         // 明文密码或 PEM 私钥; 更新时留空表示不改
	Passphrase string `json:"passphrase" form:"passphrase"` // 私钥口令明文; 更新时留空表示不改
	Remark     string `json:"remark" form:"remark"`
}

// TestConnectionReq 跳板机连接测试 / 执行命令入参
type TestConnectionReq struct {
	AssetID   uint   `json:"assetId" form:"assetId"`
	CredentialID uint `json:"credentialId" form:"credentialId"`
}

// ExecCommandReq 跳板机执行单条命令
type ExecCommandReq struct {
	AssetID   uint   `json:"assetId" form:"assetId"`
	CredentialID uint `json:"credentialId" form:"credentialId"`
	Command   string `json:"command" form:"command"`
}

// TerminalOpenReq 打开交互终端(查询参数, WS 握手时使用)
type TerminalOpenReq struct {
	AssetID   uint `json:"assetId" form:"assetId"`
	CredentialID uint `json:"credentialId" form:"credentialId"`
	Cols      int  `json:"cols" form:"cols"`
	Rows      int  `json:"rows" form:"rows"`
}

// TicketSearch 工单分页查询
type TicketSearch struct {
	common.PageInfo
	Status     string `json:"status" form:"status"`
	ApplicantID uint  `json:"applicantId" form:"applicantId"`
	PipelineID uint   `json:"pipelineId" form:"pipelineId"`
}

// TicketInput 工单创建入参
type TicketInput struct {
	Title       string                   `json:"title" form:"title"`
	PipelineID  uint                     `json:"pipelineId" form:"pipelineId"`
	Params      []map[string]interface{} `json:"params" form:"params"`
	ApplyReason string                   `json:"applyReason" form:"applyReason"`
	Remark      string                   `json:"remark" form:"remark"`
}

// ApproveTicketReq 工单审批
type ApproveTicketReq struct {
	ID      uint   `json:"id" form:"id"`
	Approve bool   `json:"approve" form:"approve"` // true=通过并触发流水线, false=拒绝
	Comment string `json:"comment" form:"comment"`
}
