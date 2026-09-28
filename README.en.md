[简体中文](README.md) | [English](README.en.md)

# New Jenkins

A full-stack DevOps management system deeply customized on top of [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin): a self-developed Jenkins-like declarative pipeline engine plus an all-in-one operations center.

[![CI](https://github.com/hequan2017/new-jenkins/actions/workflows/ci.yaml/badge.svg)](https://github.com/hequan2017/new-jenkins/actions/workflows/ci.yaml)
[![Pages](https://github.com/hequan2017/new-jenkins/actions/workflows/pages.yaml/badge.svg)](https://github.com/hequan2017/new-jenkins/actions/workflows/pages.yaml)
[![License: BSL 1.1](https://img.shields.io/badge/License-BSL%201.1-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](server/go.mod)
[![Node.js](https://img.shields.io/badge/Node.js-20-339933?logo=node.js&logoColor=white)](web/Dockerfile)
[![GitHub issues](https://img.shields.io/github/issues/hequan2017/new-jenkins)](https://github.com/hequan2017/new-jenkins/issues)

> [!IMPORTANT]
> This project is under active development. The pipeline executor and SSH capabilities run on the host where the server process runs; there is no remote agent, workspace isolation, or credential vault yet. Before deploying, please read [Security Boundaries and Current Limitations](#%EF%B8%8F-security-boundaries-and-current-limitations) and [License](#-license).

## Introduction

While keeping GVA's (gin-vue-admin) original RBAC, code generation, and plugin infrastructure, New Jenkins ships with two self-developed modules:

- **Declarative pipeline engine**: Jenkins-like `Pipeline -> Stage -> Step` orchestration with HTTP / Shell steps, manual approvals, a parameter system, real-time SSE logs, and cron / Webhook triggers — no external Jenkins service required.
- **Operations center**: asset / group / credential management, a bastion host (command execution + web terminal), SFTP file management, ticket-based releases (approvals automatically trigger pipelines), inspection monitoring, backup & restore, an alert center, a schedule center, and operation audit.

Who it is for: small and mid-sized teams and individual ops engineers / developers who want a lightweight all-in-one platform for CI/CD plus host operations, without maintaining a standalone Jenkins cluster.

## ✨ Features

### Declarative Pipeline Engine

- **Three-layer model**: `Pipeline -> Stage -> Step`, semantically aligned with Jenkins Job / Stage / Step
- **Two step types**:
  - `http`: HTTP callback; 2xx (or a configured status code) counts as success, with built-in SSRF protection (private/loopback addresses are rejected by default and only allowed with an explicit `allowPrivate`)
  - `shell`: runs shell commands in-process; exit code 0 counts as success, with timeout control and custom environment variables
- **Orchestration**: stages run sequentially; each stage can enable `Parallel` steps, be marked as `Approval` (manual approval gate), or set `ContinueOnError` (keep running later stages after a failure)
- **Parameter system**: parameters declared via `ParamSchema` (string / number / bool); required fields and types are validated at trigger time, with `${param.xxx}` / `$param.xxx` variable substitution
- **Definition/runtime separation**: stage/step configs are snapshotted when a build triggers, so later edits or deletions never affect running builds

### Build Management and Real-time Logs

- State machine `pending -> running -> running-approval -> success | failed | canceled`, with atomically assigned build numbers
- **Instant cancellation** (interrupts HTTP / Shell / approval waiting) and **re-run** reusing historical parameters
- **SSE real-time push** of `build:status` / `stage:status` / `step:status` / `step:log` event streams, with automatic catch-up after reconnection; build failures/cancellations alert administrators directly
- Logs are stored categorized as `stdout` / `stderr` / `system` and can be fetched page by page

### Trigger Methods

| Method | Description |
| --- | --- |
| Manual | Triggered from the UI with a parameter collection form |
| Schedule | Cron expressions (including `@daily`-style descriptors); schedules are restored from the database after a service restart |
| Webhook | Public endpoint `POST /webhook/trigger/{id}` authenticated via the `X-Webhook-Secret` header; the request body maps automatically to build parameters |
| Ticket | The operations center's "ticket release" automatically triggers the bound pipeline once approved |

Webhook trigger example:

```bash
curl -X POST "http://127.0.0.1:8888/webhook/trigger/1" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: <secret>" \
  -d '{"version":"1.2.3","dryRun":false}'
```

### Operations Center (12 pages)

| Module | Capabilities |
| --- | --- |
| Asset management | Asset online status, groups (prod/staging/dev), SSH credentials (password/key, stored encrypted) |
| Bastion host | Run commands on selected assets, web terminal (SSE streaming output), SFTP directory browsing / upload / rename / delete |
| Ticket release | Request → approve → automatically trigger a pipeline build, with status write-back and approval comments on record |
| Inspection | Periodically run check commands over SSH; keyword hits raise alerts automatically |
| Backup & restore | Periodically pull remote files to a local archive via SFTP, with retention control and download for restore |
| Alert center | Alerts from inspection / tickets / backups handled in one place (resolve / ignore) |
| Schedule center | Unified schedule view across pipeline crons, inspections, and backups |
| Operation audit | Logins, command executions, terminal sessions, file operations, tickets, and inspections all recorded and searchable on multiple dimensions |

### Platform Capabilities (inherited from GVA)

- RBAC (Casbin), row-level data permission, frontend/backend plugin mechanism, code generation, form designer, Swagger docs

## 🛠 Tech Stack

| Layer | Technologies |
| --- | --- |
| Frontend | Vue 3, Vite, Pinia, Element Plus, UnoCSS, ECharts, xterm (Node.js 20) |
| Backend | Go 1.24, Gin, GORM, Casbin, Viper, Zap, JWT, robfig/cron, golang.org/x/crypto/ssh |
| Storage | SQLite by default (works out of the box), also MySQL / PostgreSQL / SQL Server / Oracle; Redis optional |
| Deployment | Docker, docker-compose, Kubernetes, GitHub Pages (static frontend demo) |

## 🚀 Quick Start

Requirements: Go 1.24, Node.js 20, npm; SQLite by default, no extra services needed.

```bash
git clone https://github.com/hequan2017/new-jenkins.git
cd new-jenkins

# Start the backend (listens on http://127.0.0.1:8888 by default; SQLite writes to server/data/gva.db)
cd server
go mod download
go run .

# In another terminal, start the frontend (http://127.0.0.1:8080; /api and SSE are proxied by Vite to 8888)
cd web
npm install
npm run dev
```

Default administrator account: `admin` / `123456` (authority 888). For any externally reachable deployment, change the default password first and enable captcha, lockout, and password-strength policies under "System Settings → Security Configuration".

To get started with pipelines: log in → "Workflow Platform → Pipeline Management" → create a pipeline with parameters, stages, and HTTP/Shell steps → trigger manually or configure cron / Webhook → check stage status, step logs, and approval actions in "Build History".

Common build commands:

```bash
make build-local        # Build frontend and backend locally; artifacts go to build/
make doc                # Generate Swagger docs
make image              # Build the all-in-one image
cd server && go test ./service/workflow && go vet ./...   # Backend checks
cd web && npm run lint && npm run build                   # Frontend checks
```

Key configuration files: `server/config.yaml` (local backend config), `server/config.docker.yaml` (container environment example), `web/.env.development` / `web/.env.production` (frontend environment variables). Never commit real passwords, JWT secrets, or webhook secrets to Git; inject them through a secret manager in production.

Deployment assets: `deploy/docker-compose/docker-compose.yaml` (single-host example), `deploy/kubernetes/` (cluster manifests), `server/Dockerfile` and `web/Dockerfile` (standalone images), `deploy/docker/Dockerfile` (all-in-one image), and `.github/workflows/pages.yaml` (pushing to main automatically publishes a static frontend demo to GitHub Pages with no repo secrets required; the Pages site has no backend API and only previews the UI). The compose file contains example plaintext credentials — replace them before deploying.

## 📁 Directory Structure

```
├── server/                  Go + Gin backend
│   ├── api/v1/workflow/     Pipeline/build/approval/webhook APIs
│   ├── api/v1/ops/          Operations center APIs
│   ├── service/workflow/    Orchestration engine (engine.go), executor (executor.go), scheduler (schedule.go)
│   ├── service/ops/         SSH/SFTP/tickets/inspection/backup/alerts/audit
│   └── model/               workflow (wf_ prefix) and ops (ops_ prefix) data models
├── web/                     Vue 3 + Vite frontend
│   └── src/view/            workflow/ and ops/ (12 operations pages)
├── deploy/                  Docker, docker-compose, Kubernetes deployment assets
├── docs/screenshots/        UI screenshots
└── .github/                 CI / Pages workflows and community configs
```

## 📸 Screenshots

| Login | Dashboard |
| --- | --- |
| ![Login](docs/screenshots/login.png) | ![Dashboard](docs/screenshots/dashboard.png) |

**Workflow Platform**

| Pipeline management | Pipeline editor |
| --- | --- |
| ![Pipeline management](docs/screenshots/workflow-pipeline.png) | ![Pipeline editor](docs/screenshots/workflow-pipeline-edit.png) |

| Build history | Build detail (stages / steps / logs) |
| --- | --- |
| ![Build history](docs/screenshots/workflow-build.png) | ![Build detail](docs/screenshots/workflow-build-detail.png) |

**Operations Center**

| Operations dashboard | Asset management |
| --- | --- |
| ![Operations dashboard](docs/screenshots/ops-dashboard.png) | ![Asset management](docs/screenshots/ops-asset.png) |

| Ticket release | Alert center |
| --- | --- |
| ![Ticket release](docs/screenshots/ops-ticket.png) | ![Alert center](docs/screenshots/ops-alert.png) |

| Bastion host | SFTP file management |
| --- | --- |
| ![Bastion host](docs/screenshots/ops-bastion.png) | ![File management](docs/screenshots/ops-file.png) |

| Inspection | Backup & restore |
| --- | --- |
| ![Inspection](docs/screenshots/ops-inspect.png) | ![Backup & restore](docs/screenshots/ops-backup.png) |

| Schedule center | Operation audit |
| --- | --- |
| ![Schedule center](docs/screenshots/ops-schedule.png) | ![Operation audit](docs/screenshots/ops-audit.png) |

| Asset groups | Credential management |
| --- | --- |
| ![Asset groups](docs/screenshots/ops-group.png) | ![Credential management](docs/screenshots/ops-credential.png) |

## ⚠️ Security Boundaries and Current Limitations

- Shell commands run with the same privileges as the server process. The engine provides no container sandbox or command whitelist, so pipeline editing permissions should only be granted to trusted users.
- HTTP steps reject loopback, link-local, and private-network addresses by default; internal targets are only reachable with an explicit `allowPrivate=true`.
- SSH / SFTP capabilities act directly on target assets. Credentials are stored encrypted, but that is not a credential vault; restrict role access to the operations center menus.
- The executor runs inside the single GVA process on the local machine: no remote agent, workspace isolation, or artifact repository. The SSE hub is in-process, so multi-instance deployments would split real-time events — keep the backend to a single instance.
- After a process restart, cron registrations are restored, but builds that were `running` / `running-approval` before the restart are not taken over automatically.

Roadmap (discussion welcome via Issues): remote agents and heterogeneous execution nodes, workspace/artifact isolation, a credential vault, a cross-instance SSE event bus, interrupted-build recovery, alert notification channels (email / DingTalk / Feishu), and more.

## 🔗 Related Projects

- [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin): the base project this repository is customized from; original attribution and license statements are preserved.
- [Issues](https://github.com/hequan2017/new-jenkins/issues): bug reports and feature requests; use the [Bug Report](https://github.com/hequan2017/new-jenkins/issues/new?template=bug_report.yaml) template for bugs and the [Feature Request](https://github.com/hequan2017/new-jenkins/issues/new?template=feature_request.yaml) template for suggestions. Do not disclose security issues in public issues; report them privately to the repository owner.
- `server/docs/`: generated backend Swagger files; `aiDoc/` and `AGENTS.md`: AI collaboration rules and docs.

Contributions are welcome: for major changes, open an issue first; fork and branch from the latest `main`, keep changes focused, and add tests and docs. Commit messages should follow `type(scope): description`.

## 📄 License

The licensed work in this repository is released under the [Business Source License 1.1](LICENSE): personal use, evaluation and development use that satisfy its terms, plus non-commercial teaching, research, or academic use, are allowed; any use not expressly allowed is Production Use and requires a valid commercial license. Each version reaches its Change Date three years after first public release and converts to Apache License 2.0. See the [LICENSE](LICENSE) file for the authoritative terms.

This project is customized from gin-vue-admin, and its original attribution, license statements, and brand rights are preserved; this license does not grant rights to its trademarks, service marks, trade names, or logos.
