# 自研类 Jenkins 声明式流水线引擎

## 基本信息

- 提出日期：2026-08-12
- 当前状态：`active`
- 需求类型：新模块
- 优先级：高
- 需求文件：`aiDoc/memory/business/active/workflow-pipeline-engine.md`

## 用户原始意图摘要

在 GVA 项目内自研一套类 Jenkins 的工作流/发版管理平台:声明式 Pipeline→Stage→Step 流水线引擎,借鉴 Jenkins 的核心概念与交互(Stage 视图、构建历史、参数化触发),不对接真实 Jenkins,GVA 进程内执行,构建状态与日志通过 SSE 实时推送。

## 影响范围

- 后端：新增 `server/model/workflow/`、`server/service/workflow/`、`server/api/v1/workflow/`、`server/router/workflow/`;修改三层 `enter.go`、`initialize/gorm_biz.go`、`initialize/router_biz.go`
- 前端：新增 `web/src/api/workflow.js`、`web/src/view/workflow/{pipeline,build}/*.vue`
- 文档：新增 `aiDoc/modules/workflow.md`
- 插件 / 模块：新增业务模块 `workflow`,挂载于 `bizModel()` / `initBizRouter()`,不污染 system 模块

## 涉及对象

- 模块：`workflow`
- 接口：`/workflow/createPipeline|updatePipeline|deletePipeline|getPipelineList|findPipeline|triggerBuild|cancelBuild|approveStage|getBuildList|getBuildDetail|getBuildLog|buildStream`
- 页面：`WorkflowPipelineList`、`WorkflowPipelineEdit`、`WorkflowBuildList`、`WorkflowBuildDetail`
- 配置：无独立配置;菜单/权限通过后台菜单管理手动建(本期不写 source seed)

## 已确认约束

- 只借鉴 Jenkins 概念与交互,不对接真实 Jenkins API
- Step 类型:HTTP 回调 + Shell(本地命令),GVA 进程内执行
- 状态/日志:SSE 实时推送(复用 `utils/sse` Hub),历史日志走分页接口
- 本期产出颗粒度:后端完整可编译 + 前端核心页
- HTTP 步骤默认 SSRF 防护(禁止内网/环回),可由 `allowPrivate` 显式放行
- Shell 步骤执行走子进程隔离 + 超时,不做命令白名单(视同 Jenkins sh)
- 模块表不带数据权限公共字段(流转记录,非部门归属资源)
- 引擎 goroutine 内用 `datascope.WithSystem(ctx)`,不裸 `context.Background()`

## 当前进展

- 后端 model 层 7 张表 + request/response 结构 ✓
- Service 层:PipelineService(CRUD + 级联)、BuildService(触发/查询/取消/审批)✓
- Engine + httpExecutor + shellExecutor + SSE 发布 ✓
- API + Router + Swagger 注释(全部 PrivateGroup + ApiKeyAuth + @Success 落具体类型)✓
- Initialize 注册(gorm_biz AutoMigrate、router_biz、三层 enter.go)✓
- `go build ./...` + `go vet` 通过;Engine 状态机单测 3 条(成功/失败/审批 gate)✓
- 前端 api/workflow.js + 四个核心页(pipeline 列表/编辑、build 列表/详情)✓
- build 详情页:Stage 横条 + 选中 Step 日志 + SSE 订阅 + 审批操作 ✓

## 后续待办

- 菜单/权限 seed:如需开箱即用,在 `server/source/system/menu.go` 补 workflow 菜单 seed(本期走后台手动建)
- schedule/webhook 触发:当前 triggerType 字段预留,调度器与 webhook 入口未实现
- Step 并行:当前 Stage 内 Step 串行,后续支持 parallel
- 远程 Agent 架构:StepExecutor 已可插拔,后续可扩展远程 Agent
- SSE 多实例:Hub 为进程内实现,多实例部署需上层 Redis 扇出(见 `utils/sse/hub.go` 注释)

## 更新规则

- 本文件只承载**自研流水线引擎**这一个功能点
- 出现新功能点(如远程 Agent、Webhook 入口、调度器)时,新建独立文件(前缀 `workflow-`),不要追加到本文件
- 同模块相关功能以反向链接互相指向
