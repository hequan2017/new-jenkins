# 首页与登录页项目化定制

## 基本信息

- 提出日期：2026-08-14
- 当前状态：`active`
- 需求类型：前端定制
- 优先级：高
- 需求文件：`aiDoc/memory/business/active/home-login-customization.md`

## 用户原始意图摘要

修改登录页与首页使其贴合本项目定位：登录页去掉验证码展示、清理误导性的「前往初始化」入口；首页改为聚合流水线与运维数据的平台大盘；新开发的功能菜单（工作流平台、运维中心）置顶显示。

## 影响范围

- 后端：`server/source/system/menu.go` 一级菜单 Sort 重排（workflow=1、ops=2，其余顺延）；本地库 `sys_base_menus.sort` 同步更新
- 前端：`web/src/view/login/index.vue`（平台副标题 + needInit 才显示初始化按钮）、`web/src/view/dashboard/`（统计卡/最近构建/最新告警/平台入口四个新组件 + quickLinks 换项目入口）
- 文档：根 README 快速开始补充默认账号 admin/123456 说明
- 插件 / 模块：无

## 涉及对象

- 模块：dashboard、login、menu 种子
- 接口：`/ops/getDashboard`、`/workflow/getBuildList`、`/workflow/getPipelineList`、`/ops/getAlertList`
- 页面：登录页、首页大盘
- 配置：本地库 `sys_security_config.captcha_open=99`（关闭验证码展示，仅本地开发库）

## 已确认约束

- 版权保护：`bottomInfo.vue`（头部版权声明文件）、商业授权卡片、购买商业授权/插件市场按钮、授权购买/插件市场/项目仓库外链全部原样保留，不做删除、隐藏或改写
- 假数据图表（访问人数/新增客户等）、GVA commits 表、插件表仅从首页布局移除，组件文件保留不删
- 修复顺带缺陷：`pinia/modules/router.js` 菜单缺失时 `.children.forEach` 崩溃（两处防御）；`workflow/pipeline/edit.vue` webhook 地址 `{{baseUrl}}` 字面量 bug
