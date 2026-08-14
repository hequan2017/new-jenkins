package ops

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	OpsRouter
}

// opsApi 直接取 ops 模块的 API 聚合(含 AssetApi/CredentialApi/BastionApi/TicketApi)。
// 通过嵌入字段提升, 可直接调用 opsApi.CreateAsset 等。
var opsApi = api.ApiGroupApp.OpsApiGroup
