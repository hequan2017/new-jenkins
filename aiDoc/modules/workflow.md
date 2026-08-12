# workflow 模块说明

## 这个模块是做什么的

自研一套类 Jenkins 的声明式流水线(Pipeline)引擎,用于工作流与发版管理。借鉴 Jenkins 的核心概念与交互:Pipeline 由有序的 Stage 组成,每个 Stage 由有序的 Step 组成;一次"构建(Build)"按定义顺序执行,实时展示 Stage/Step 状态与日志,支持人工审批 gate。

当前范围:Step 类型为 HTTP 回调与本地 Shell 命令,在 GVA 进程内执行;支持手动、cron 和 Webhook 触发、阶段内并行、人工审批、失败继续、克隆、重跑与即时取消;状态与日志通过 SSE 实时推送。

## 它的后端入口在哪里

- 数据模型:`server/model/workflow/`(`pipeline.go`、`pipeline_stage.go`、`pipeline_step.go`、`pipeline_build.go`、`pipeline_build_stage.go`、`pipeline_build_step.go`、`pipeline_build_log.go`)
- 请求/响应结构:`server/model/workflow/{request,response}/workflow.go`
- Service:`server/service/workflow/`(`pipeline.go` 定义 CRUD、`build.go` 触发/查询/取消/审批、`engine.go` 执行核心、`executor.go` HTTP/Shell 执行器)
- API:`server/api/v1/workflow/`
- Router:`server/router/workflow/`
- 表注册:`server/initialize/gorm_biz.go` 的 `bizModel()`
- 路由注册:`server/initialize/router_biz.go` 的 `initBizRouter()`

聚合入口:`service.ServiceGroupApp.WorkflowServiceGroup`、`api.ApiGroupApp.WorkflowApiGroup`、`router.RouterGroupApp.Workflow`。

## 它的前端入口在哪里

- API 封装:`web/src/api/workflow.js`
- 页面:`web/src/view/workflow/`
  - `pipeline/index.vue` 流水线列表(WorkflowPipelineList)
  - `pipeline/edit.vue` Stage/Step 可视化编辑(WorkflowPipelineEdit)
  - `build/index.vue` 构建历史(WorkflowBuildList)
  - `build/detail.vue` 构建详情:Stage 横条 + Step 日志 + SSE + 审批(WorkflowBuildDetail)
- 菜单/权限:`server/source/system/menu.go` 与 `casbin.go` 已提供管理员/示例角色种子；页面 `component` 与 `name` 对应上述 Vue 组件

## 它依赖哪些数据或其他模块

- `global.GVA_DB` / `GVA_LOG`
- `utils/sse`(SSE 推送中枢,`sse.Default()`)
- `utils/datascope`(引擎 goroutine 用 `datascope.WithSystem(ctx)`)
- `model/common/request.PageInfo` + `LimitOffset()`(分页)
- 不依赖任何 system 业务 Service;不写数据权限过滤条件(本模块表非部门归属资源)

## 它对外暴露什么契约

除公开 Webhook 外，其余接口位于 PrivateGroup 并使用 `@Security ApiKeyAuth`：

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /workflow/createPipeline | 创建流水线(含 Stage/Step 树,级联) |
| PUT | /workflow/updatePipeline | 全量覆盖更新 |
| DELETE | /workflow/deletePipeline | 删除流水线 |
| POST | /workflow/getPipelineList | 分页列表 |
| GET | /workflow/findPipeline | 详情(含 Stage/Step 树) |
| POST | /workflow/triggerBuild | 触发构建,返回 buildId |
| POST | /workflow/cancelBuild | 取消构建 |
| POST | /workflow/approveStage | 审批 gate(通过/拒绝) |
| POST | /workflow/getBuildList | 构建历史 |
| GET | /workflow/getBuildDetail | 构建详情(build + stage + step 视图) |
| GET | /workflow/getBuildLog | 分页历史日志(按 step 维度) |
| GET | /workflow/buildStream | SSE 实时事件(状态 + 日志增量) |
| POST | /webhook/trigger/:id | 公开 Webhook 触发,使用 `X-Webhook-Secret` 鉴权 |

响应遵循统一结构 `{code,data,msg}` 与分页 `{page,pageSize,total,list}`。

## 它有哪些必须记住的限制

- 表**不带** `CreatedBy/UpdatedBy/DeletedBy/DeptId`:这些是构建流转记录,不是按部门做行级过滤的资源;`TriggerBy` 是普通 `uint` 字段,语义不同于归属列
- SSE 按用户维度投递,事件目前发给**触发者**;多实例部署需上层 Redis 扇出(见 `utils/sse/hub.go` 注释)
- `buildStream` 路由**绝不挂 TimeoutMiddleware**(与 SSE 流式响应冲突,见 hub.go:168)
- 引擎 goroutine 内必须用 `datascope.WithSystem(ctx)`,不裸用 `context.Background()`
- HTTP 步骤默认 SSRF 防护(禁止内网/环回),需 `config.allowPrivate=true` 才放行
- Shell 步骤走子进程(`sh -c` / Windows `cmd /c`)+ 超时,不做命令白名单(等同于 Jenkins 的 sh,由流水线编辑者负责)
- 日志独立成表(`wf_pipeline_build_logs`),按 `(build_id, step_id, seq)` 定位,避免主表行膨胀
- 构建触发事务会完整快照 Stage 行为与 Step 配置；定义后续修改或删除不影响已触发构建
- `(pipeline_id, build_no)` 有唯一索引，触发事务通过流水线行锁串行分配构建号
- 取消会关闭运行 context，可中断 HTTP、Shell 与审批等待；终态更新不得覆盖已落库的 `canceled`
- 审批 gate 的 Stage 跑完前置 Step 后,build 置 `running-approval`,等 `approveStage` 唤醒;最后一个 Stage 的 Approval 不阻塞
- cron 调度在服务启动时恢复注册；运行中构建与审批等待不会在进程重启后自动恢复
