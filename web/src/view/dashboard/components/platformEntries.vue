<template>
  <div class="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-4">
    <button
      v-for="item in entries"
      :key="item.path"
      class="group flex w-full items-center gap-3 rounded-lg border border-black/10 bg-white/70 p-3 text-left transition-all duration-200 hover:border-[var(--el-color-primary)] hover:shadow-sm dark:border-white/10 dark:bg-white/[0.02]"
      type="button"
      @click="toPath(item)"
    >
      <span
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-base-text transition-colors group-hover:bg-[var(--el-color-primary)] group-hover:text-white"
      >
        <el-icon><component :is="item.icon" /></el-icon>
      </span>
      <span class="min-w-0">
        <span class="block text-sm text-base-text">{{ item.title }}</span>
        <span class="block truncate text-xs text-muted-foreground">
          {{ item.desc }}
        </span>
      </span>
    </button>
  </div>
</template>

<script setup>
  import {
    Connection,
    DataAnalysis,
    Monitor,
    Cpu,
    Tickets,
    Bell,
    Compass,
    FolderOpened
  } from '@element-plus/icons-vue'
  import { useRouter } from 'vue-router'

  defineOptions({
    name: 'GvaPlatformEntries'
  })

  const router = useRouter()

  const toPath = (item) => {
    router.push({ name: item.path })
  }

  const entries = [
    {
      icon: Connection,
      title: '流水线管理',
      desc: '声明式 CI/CD 编排',
      path: 'WorkflowPipelineList'
    },
    {
      icon: DataAnalysis,
      title: '构建历史',
      desc: '状态 · 日志 · 审批',
      path: 'WorkflowBuildList'
    },
    {
      icon: Monitor,
      title: '运维大盘',
      desc: '资产与运行总览',
      path: 'OpsDashboard'
    },
    {
      icon: Cpu,
      title: '资产管理',
      desc: '主机 · 分组 · 凭据',
      path: 'OpsAssetList'
    },
    {
      icon: Tickets,
      title: '工单发版',
      desc: '申请 · 审批 · 发布',
      path: 'OpsTicketList'
    },
    {
      icon: Bell,
      title: '告警中心',
      desc: '巡检与备份告警',
      path: 'OpsAlertList'
    },
    {
      icon: Compass,
      title: '跳板机',
      desc: '命令执行 · 终端',
      path: 'OpsBastion'
    },
    {
      icon: FolderOpened,
      title: '文件管理',
      desc: 'SFTP 在线管理',
      path: 'OpsFile'
    }
  ]
</script>
