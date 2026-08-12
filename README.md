# gin-vue-admin（自研流水线版）

基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin) 深度定制的全栈管理系统。除保留 GVA 原有的 RBAC 权限、代码生成、插件化等基础设施外，核心亮点是内置了一套**自研类 Jenkins 声明式流水线引擎**，以轻量可插拔的方式提供 CI/CD / 任务编排能力，不依赖外部 Jenkins 服务。

## 核心特性

### 声明式流水线引擎

- **三层模型**：`流水线(Pipeline) -> 阶段(Stage) -> 步骤(Step)`，语义对齐 Jenkins 的 Job / Stage / Step
- **两种步骤类型**：
  - `http`：HTTP 回调，2xx 或指定状态码视为成功，内置 SSRF 防护（默认禁止内网/环回地址，可显式 `allowPrivate` 放行），响应体截断 64KB 写入日志
  - `shell`：进程内执行 Shell 命令，退出码 0 视为成功，支持超时控制与自定义环境变量，Windows 下自动回退 `sh` / `cmd`
- **编排能力**：
  - 阶段按序执行，步骤默认串行；阶段可开启 `Parallel` 实现步骤并行执行
  - 阶段可标记 `Approval`（人工审批 gate），跑完后构建进入 `running-approval` 等待人工批准/拒绝
  - 阶段可标记 `ContinueOnError`，步骤失败后继续执行后续阶段
- **参数体系**：流水线可声明 `ParamSchema`（string / number / bool），触发时按 Schema 校验必填与类型、回填默认值；执行器支持 `${param.xxx}` 与 `$param.xxx` 变量替换
- **定义与运行分离**：Stage/Step 定义在触发构建时快照到运行实例，后续修改定义不影响历史构建

### 构建管理

- **状态机**：`pending -> running -> running-approval -> success | failed | canceled`
- 构建序号（buildNo）同流水线自增，历史记录分页可查
- 支持**取消**（running / running-approval / pending 状态）与**重跑**（复用历史参数触发新构建）
- 构建详情按阶段/步骤展示运行视图，日志独立落表支持分页拉取

### 实时日志与告警

- **SSE 实时推送**：`build:status` / `stage:status` / `step:status` / `step:log` 事件流，构建过程实时刷新；断线重连后从数据库拉取补全
- **失败告警**：构建失败/取消时通过 SSE 定向推送给管理员角色（authority_id=888）用户
- 日志按 `stdout` / `stderr` / `system` 三种流分类，每条日志带时间戳与步骤内自增序号

### 触发方式

| 方式 | 说明 |
| --- | --- |
| 手动 | 页面触发，带参数收集表单 |
| 定时 | cron 表达式（支持 `@daily` 等描述符，可含秒位），基于 `robfig/cron` 注册；服务重启后自动从数据库恢复全部调度，启停联动 `enabled` 状态 |
| Webhook | 公开触发入口 `POST /webhook/trigger/{id}`，以 `X-Webhook-Secret` 头做密钥鉴权，请求体键值对自动映射为构建参数 |

### 流水线管理

- 增删改查（级联保存 Stage/Step 树）、启用/停用、**克隆**（深拷贝定义树，克隆后默认手动触发避免重复定时）
- webhook 类型自动生成 32 字节随机密钥，更新时不覆盖已有密钥

### 平台能力（继承自 GVA）

- RBAC 权限控制（Casbin）、行级数据权限（GORM 全局回调）
- 前后端插件机制，插件可独立打包分发
- 代码生成、表单设计器、Swagger 文档、统一响应结构

## 技术栈

| 端 | 技术 |
| --- | --- |
| 前端 | Vue 3、Vite、Pinia、Element Plus、UnoCSS、Vue Router、Axios、ECharts、VueUse、SSE |
| 后端 | Go、Gin、GORM、Casbin、Viper、Zap、Redis、JWT、robfig/cron |
| 存储 | 默认 SQLite（开箱即用），支持 MySQL / PostgreSQL / SQL Server / Oracle |
| 部署 | Docker、docker-compose、Kubernetes |

## 目录结构

```
├── server/                     Go + Gin 后端
│   ├── api/v1/workflow/        API 层（触发/审批/Webhook 等）
│   ├── model/workflow/         数据模型与 request/response 结构
│   ├── service/workflow/       Service 层（引擎 / 执行器 / 调度 / 构建 / 流水线）
│   │   ├── engine.go           编排核心：Stage→Step 状态机 + SSE + 审批 gate
│   │   ├── executor.go         可插拔执行器（http / shell）+ 变量替换 + SSRF 防护
│   │   ├── schedule.go         定时调度（robfig/cron，重启恢复）
│   │   ├── build.go            构建触发 / 取消 / 重跑 / 审批
│   │   └── pipeline.go         流水线定义 CRUD / 克隆 / 校验
│   ├── router/workflow/        路由注册（含公开 Webhook 路由）
│   ├── initialize/             启动装配（含 LoadWorkflowSchedules 恢复调度）
│   └── ...                     GVA 既有基础设施
├── web/                        Vue 3 + Vite 前端
│   ├── src/view/workflow/      流水线列表/编辑、构建列表/详情页面
│   └── src/api/workflow.js     前端 API 封装
├── deploy/                     部署资产（Docker、docker-compose、Kubernetes）
├── aiDoc/                      结构化 AI 协作文档层（规则、示例、记忆）
└── docs/                       项目文档与设计记录
```

## 数据模型（workflow 模块，`wf_` 前缀）

| 表 | 说明 |
| --- | --- |
| `wf_pipelines` | 流水线定义（名称、触发方式、cron、webhook 密钥、参数 Schema、启停） |
| `wf_pipeline_stages` | 阶段定义（顺序、是否审批、是否并行、是否容错继续） |
| `wf_pipeline_steps` | 步骤定义（类型、JSON 配置、顺序） |
| `wf_pipeline_builds` | 构建实例（状态机、入参、触发方式/人、时间） |
| `wf_pipeline_build_stages` | 构建阶段运行视图（名称/顺序快照、状态、时间） |
| `wf_pipeline_build_steps` | 构建步骤运行视图（类型/顺序快照、退出码、时间） |
| `wf_pipeline_build_logs` | 日志行（build+step+seq 定位，流分类，分页/推流） |

## 接口一览

### 流水线管理（需登录）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/workflow/createPipeline` | 创建流水线 |
| PUT | `/workflow/updatePipeline` | 更新流水线（全量覆盖 Stage/Step 树） |
| DELETE | `/workflow/deletePipeline` | 删除流水线（构建历史保留） |
| POST | `/workflow/togglePipeline` | 启用 / 停用 |
| POST | `/workflow/clonePipeline` | 克隆流水线 |
| POST | `/workflow/getPipelineList` | 流水线分页列表 |
| GET | `/workflow/findPipeline` | 流水线详情（含 Stage/Step 树） |

### 构建管理（需登录）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/workflow/triggerBuild` | 触发构建（带参数） |
| POST | `/workflow/cancelBuild` | 取消构建 |
| POST | `/workflow/retryBuild` | 重跑历史构建 |
| POST | `/workflow/approveStage` | 审批 gate（批准 / 拒绝） |
| POST | `/workflow/getBuildList` | 构建历史分页列表 |
| GET | `/workflow/getBuildDetail` | 构建详情（阶段/步骤视图） |
| GET | `/workflow/getBuildLog` | 构建日志分页拉取 |
| GET | `/workflow/buildStream` | SSE 实时事件流 |

### Webhook（公开，免登录）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/webhook/trigger/{id}` | 触发流水线，需 `X-Webhook-Secret` 头 |

## 快速开始

### 本地开发

**后端**（`server/` 目录）：

```bash
go mod tidy
go run main.go
```

- 默认监听 `:8888`，数据库使用 SQLite（`server/config.yaml` → `system.db-type: sqlite`，数据落在 `data/gva.db`）
- 首次启动自动建表并初始化默认数据（账号见初始化日志）
- 定时流水线在启动时自动恢复注册，无需额外配置

**前端**（`web/` 目录）：

```bash
npm install
npm run dev
```

默认开发地址 `http://localhost:8080`，Vite 已代理 API 与 SSE 到后端。

### 构建与部署

```bash
# 本地打包前后端（产物进 build/）
make build-local

# 生成 Swagger 文档
make doc

# Docker 镜像构建
make build-image-web / build-image-server / image / images
```

部署编排见 `deploy/docker-compose/docker-compose.yaml`（前端 + 后端二合一），Kubernetes 清单见 `deploy/kubernetes/`，镜像 Dockerfile 见 `deploy/docker/Dockerfile`。

## 文档

- `AGENTS.md`：AI 协作规则唯一真源
- `aiDoc/`：结构化 AI 文档层（技术栈画像、开发流程、分层规则、示例等）
- `docs/`：项目文档与设计记录

## 许可证

本项目遵循 [LICENSE](LICENSE) 许可。
