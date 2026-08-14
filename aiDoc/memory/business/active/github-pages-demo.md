# GitHub Pages 静态演示一键部署

## 基本信息

- 提出日期：2026-08-14
- 当前状态：`active`
- 需求类型：CI/CD 与部署
- 优先级：中
- 需求文件：`aiDoc/memory/business/active/github-pages-demo.md`

## 用户原始意图摘要

补上 GitHub Pages 一键部署版本：推送到 main 后由 GitHub Actions 自动测试、构建并发布静态站点，且不需要配置任何仓库 Secret。

## 影响范围

- 后端：无代码改动（CI 中跑 `go vet` + `go test ./service/workflow/...`）
- 前端：构建命令使用 `npx vite build --mode production --base=./`（hash 路由 + 相对资源路径适配 Pages 子路径），不改 vite.config.js
- 文档：README 新增「GitHub Pages 静态演示」章节；部署方式表补充 Pages 行
- 插件 / 模块：无

## 涉及对象

- 模块：CI 工作流
- 接口：无
- 页面：无
- 配置：`.github/workflows/pages.yaml`（permissions: contents:read / pages:write / id-token:write，内置 GITHUB_TOKEN + OIDC，无 Secret）

## 已确认约束

- 不引入任何 Secret：全部使用 Actions 内置凭据
- Pages 站点无后端 API，仅界面预览；README 中明确标注该限制，完整体验引导回本地启动/部署章节
- 仓库需一次性设置 Settings → Pages → Source = GitHub Actions
- 测试失败则阻断发布（needs: test），测试范围保持轻量（vet + workflow 单测 + eslint），不跑全量 `go test ./...` 以避开既有基线问题
