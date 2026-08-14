package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	opssvc "github.com/flipped-aurora/gin-vue-admin/server/service/ops"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/ws"
	"github.com/gin-gonic/gin"
)

type BastionApi struct{}

// TestConnection
// @Tags      Ops
// @Summary   测试资产+凭据的 SSH 连接
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.TestConnectionReq                  true  "资产ID与凭据ID"
// @Success   200   {object}  response.Response{msg=string}             "连接成功"
// @Router    /ops/testConnection [post]
func (a *BastionApi) TestConnection(c *gin.Context) {
	var req opsReq.TestConnectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := sshService.TestConnection(c.Request.Context(), req.AssetID, req.CredentialID); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Warn("跳板机连接测试失败")
		response.FailWithMessage("连接失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("连接成功", c)
}

// ExecCommand
// @Tags      Ops
// @Summary   在目标主机执行单条命令并返回输出
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.ExecCommandReq                     true  "资产/凭据/命令"
// @Success   200   {object}  response.Response{data=object,msg=string} "执行完成, 返回输出"
// @Router    /ops/execCommand [post]
func (a *BastionApi) ExecCommand(c *gin.Context) {
	var req opsReq.ExecCommandReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	out, err := sshService.ExecCommand(c.Request.Context(), req.AssetID, req.CredentialID, req.Command)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Warn("跳板机命令执行失败")
		response.OkWithDetailed(gin.H{"output": out, "error": err.Error()}, "执行完成(含错误)", c)
		return
	}
	response.OkWithDetailed(gin.H{"output": out}, "执行完成", c)
}

// TerminalStream 交互式 SSH 终端(WebSocket)
// 升级为 WS 后, 建立 SSH PTY 会话并桥接: 前端输入 -> PTY stdin, PTY stdout -> 前端。
// 该路由挂在 PrivateGroup(已完成 JWT/Casbin 鉴权), 绝不能套 TimeoutMiddleware。
// @Tags      Ops
// @Summary   交互式 SSH 终端(WebSocket)
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     assetId       query  int  true  "资产ID"
// @Param     credentialId  query  int  true  "凭据ID"
// @Param     cols          query  int  false "初始列数"
// @Param     rows          query  int  false "初始行数"
// @Success   200  {object}  response.Response{msg=string}  "WS 流"
// @Router    /ops/terminal [get]
func (a *BastionApi) TerminalStream(c *gin.Context) {
	conn, err := ws.Upgrade(c)
	if err != nil {
		// Upgrade 已自行写入 http 响应, 这里不再重复输出
		return
	}
	defer conn.Close()

	var req opsReq.TerminalOpenReq
	req.AssetID = parseUint(c.Query("assetId"))
	req.CredentialID = parseUint(c.Query("credentialId"))
	req.Cols = parseInt(c.Query("cols"))
	req.Rows = parseInt(c.Query("rows"))
	if req.AssetID == 0 || req.CredentialID == 0 {
		writeTermMsg(conn, opssvc.TermMsgError, "缺少 assetId / credentialId")
		return
	}

	tc := &terminalConnAdapter{conn: conn}
	_ = terminalService.Serve(c.Request.Context(), req.AssetID, req.CredentialID, req.Cols, req.Rows, tc)
}

// terminalConnAdapter 把 ws.Conn 适配成 service 的 TerminalConn 接口
type terminalConnAdapter struct {
	conn *ws.Conn
}

func (t *terminalConnAdapter) Read() (opssvc.TerminalMessage, error) {
	raw, err := t.conn.ReadText()
	if err != nil {
		return opssvc.TerminalMessage{}, err
	}
	return opssvc.DecodeTerminalMsg(raw)
}

func (t *terminalConnAdapter) Write(m opssvc.TerminalMessage) error {
	s, err := opssvc.EncodeTerminalMsg(m)
	if err != nil {
		return err
	}
	return t.conn.WriteText(s)
}

func (t *terminalConnAdapter) Close() error { return t.conn.Close() }
func (t *terminalConnAdapter) Done() <-chan struct{} { return t.conn.Done() }

func writeTermMsg(conn *ws.Conn, typ, data string) {
	m, _ := opssvc.EncodeTerminalMsg(opssvc.TerminalMessage{Type: typ, Data: data})
	_ = conn.WriteText(m)
}
