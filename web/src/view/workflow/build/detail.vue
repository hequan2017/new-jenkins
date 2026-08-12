<template>
  <div class="gva-table-box">
    <!-- 顶部:构建概要 -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-3">
        <span class="text-base font-medium">构建 #{{ detail.build.buildNo }}</span>
        <el-tag :type="buildStatusType(detail.build.status)">
          {{ buildStatusLabel(detail.build.status) }}
        </el-tag>
        <span class="text-sm text-gray-500">{{ detail.build.message }}</span>
      </div>
      <div class="flex gap-2">
        <el-button
          v-if="detail.build.status === 'running-approval'"
          type="success"
          @click="onApprove(true)"
        >审批通过</el-button
        >
        <el-button
          v-if="detail.build.status === 'running-approval'"
          type="danger"
          @click="onApprove(false)"
        >拒绝</el-button
        >
        <el-button
          v-if="canCancel"
          type="warning"
          @click="onCancel"
        >取消构建</el-button
        >
        <el-button @click="onBack">返回</el-button>
      </div>
    </div>

    <!-- Stage 横条(类 Jenkins 蓝海视图) -->
    <div class="flex items-stretch gap-2 mb-4 flex-wrap">
      <div
        v-for="st in detail.stages"
        :key="st.ID"
        class="border rounded px-4 py-3 cursor-pointer min-w-32 transition"
        :class="stageClass(st)"
        @click="onSelectStage(st)"
      >
        <div class="text-sm font-medium">{{ st.snapshotName }}</div>
        <div class="text-xs mt-1">
          <el-icon v-if="st.status === 'running'" class="is-loading"><Loading /></el-icon>
          {{ buildStatusLabel(st.status) }}
        </div>
      </div>
    </div>

    <!-- Step 列表(选中 stage 下) -->
    <div class="mb-2 text-sm text-gray-600">
      阶段「{{ selectedStageName }}」的步骤
    </div>
    <el-table
      :data="currentSteps"
      size="small"
      highlight-current-row
      @row-click="onSelectStep"
    >
      <el-table-column label="步骤" prop="snapshotName" min-width="160" />
      <el-table-column label="类型" prop="snapshotType" width="90" />
      <el-table-column label="状态" width="110">
        <template #default="scope">
          <el-tag :type="buildStatusType(scope.row.status)" size="small">
            {{ buildStatusLabel(scope.row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="退出码" prop="exitCode" width="90" />
      <el-table-column label="开始" width="170">
        <template #default="scope">{{ formatDate(scope.row.startedAt) }}</template>
      </el-table-column>
      <el-table-column label="结束" width="170">
        <template #default="scope">{{ formatDate(scope.row.finishedAt) }}</template>
      </el-table-column>
    </el-table>

    <!-- 日志区 -->
    <div class="mt-3">
      <div class="flex items-center justify-between mb-2">
        <span class="text-sm">日志{{ selectedStepId ? '(步骤 #' + selectedStepId + ')' : '(选择一个步骤查看)' }}</span>
        <el-button size="small" plain @click="loadHistoryLog">刷新历史日志</el-button>
      </div>
      <div
        ref="logBox"
        class="bg-black text-gray-100 font-mono text-xs p-3 rounded overflow-auto"
        style="height: 360px"
      >
        <div
          v-for="(line, i) in logLines"
          :key="i"
          :class="line.stream === 'stderr' ? 'text-red-400' : line.stream === 'system' ? 'text-yellow-400' : ''"
        >
          <span class="text-gray-500 select-none">[{{ line.stream }}]</span> {{ line.text }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { Loading } from '@element-plus/icons-vue'
  import {
    getBuildDetail,
    getBuildLog,
    cancelBuild,
    approveStage,
    buildStatusLabel,
    buildStatusType
  } from '@/api/workflow'

  defineOptions({ name: 'WorkflowBuildDetail' })

  const route = useRoute()
  const router = useRouter()
  const buildId = Number(route.params.id)

  const detail = reactive({ build: {}, stages: [], steps: [] })
  const selectedStageId = ref(0)
  const selectedStepId = ref(0)
  const logLines = ref([])
  const logBox = ref(null)

  const selectedStageName = computed(() => {
    const st = detail.stages.find((s) => s.ID === selectedStageId.value)
    return st ? st.snapshotName : ''
  })
  const currentSteps = computed(() =>
    selectedStageId.value
      ? detail.steps.filter((s) => s.stageId === selectedStageId.value)
      : []
  )
  const canCancel = computed(() =>
    ['running', 'pending', 'running-approval'].includes(detail.build.status)
  )

  const stageClass = (st) => {
    const base = 'bg-white'
    if (selectedStageId.value === st.ID) return 'ring-2 ring-blue-500 ' + base
    switch (st.status) {
      case 'success':
        return 'bg-green-50 border-green-300'
      case 'failed':
        return 'bg-red-50 border-red-300'
      case 'running':
        return 'bg-blue-50 border-blue-300'
      case 'running-approval':
        return 'bg-yellow-50 border-yellow-300'
      case 'canceled':
        return 'bg-gray-50 border-gray-300'
      default:
        return base
    }
  }

  let es = null
  let pollTimer = null

  const loadDetail = async () => {
    const res = await getBuildDetail({ id: buildId })
    if (res.code === 0) {
      const oldStatus = detail.build.status
      Object.assign(detail, res.data)
      // 默认选第一个 stage
      if (!selectedStageId.value && detail.stages.length) {
        selectedStageId.value = detail.stages[0].ID
      }
      // 构建结束则停止轮询
      if (['success', 'failed', 'canceled'].includes(detail.build.status)) {
        stopPolling()
      } else if (oldStatus && oldStatus !== detail.build.status) {
        // 状态变化时刷新当前 step 历史日志
        if (selectedStepId.value) loadHistoryLog()
      }
    }
  }

  const onSelectStage = (st) => {
    selectedStageId.value = st.ID
    selectedStepId.value = 0
    logLines.value = []
  }

  const onSelectStep = async (row) => {
    selectedStepId.value = row.ID
    logLines.value = []
    await loadHistoryLog()
  }

  const loadHistoryLog = async () => {
    if (!selectedStepId.value) return
    const res = await getBuildLog({
      buildId,
      stepId: selectedStepId.value,
      page: 1,
      pageSize: 500
    })
    if (res.code === 0) {
      logLines.value = (res.data.list || []).map((l) => ({
        stream: l.stream,
        text: l.text,
        ts: l.ts
      }))
      await nextTick()
      if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
    }
  }

  const startSSE = () => {
    // 复用全局 EventSource: 后端 buildStream 按用户维度推送
    // 鉴权: 登录时后端已种 x-token cookie(HttpOnly), EventSource 自动随请求携带,
    // 后端 GetToken 优先读 x-token Header, 回退 Cookie(本项目 JWTAuth 支持)。
    // 故 URL 不拼 token, 依赖 cookie。
    const base = import.meta.env.VITE_BASE_API || ''
    const url = `${base}/workflow/buildStream`
    try {
      es = new EventSource(url, { withCredentials: true })
    } catch (e) {
      // EventSource 不可用时回退轮询(下方 startPolling)
      return
    }
    const onEvent = (name, data) => {
      try {
        const payload = JSON.parse(data)
        // 仅处理本构建事件
        if (Number(payload.buildId) !== buildId) return
        if (name === 'build:status' || name === 'stage:status' || name === 'step:status') {
          loadDetail()
        } else if (name === 'step:log') {
          if (Number(payload.stepId) === selectedStepId.value) {
            logLines.value.push({
              stream: payload.stream,
              text: payload.text,
              ts: payload.ts
            })
            nextTick(() => {
              if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
            })
          }
        }
      } catch {
        // 忽略非 JSON 帧
      }
    }
    es.onmessage = (e) => onEvent('message', e.data)
    ;['build:status', 'stage:status', 'step:status', 'step:log'].forEach((n) => {
      if (es) es.addEventListener(n, (e) => onEvent(n, e.data))
    })
    es.onerror = () => {
      // EventSource 会自动重连; 失败严重时降级轮询
    }
  }

  const startPolling = () => {
    if (pollTimer) return
    pollTimer = setInterval(loadDetail, 3000)
  }
  const stopPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  const onCancel = async () => {
    try {
      await ElMessageBox.confirm('确定取消此构建?', '取消确认', { type: 'warning' })
    } catch {
      return
    }
    const res = await cancelBuild({ id: buildId })
    if (res.code === 0) {
      ElMessage.success('已取消')
      loadDetail()
    }
  }

  const onApprove = async (approve) => {
    const res = await approveStage({ buildId, approve, comment: '' })
    if (res.code === 0) {
      ElMessage.success('审批已提交')
      loadDetail()
    }
  }

  const onBack = () => router.back()

  onMounted(async () => {
    await loadDetail()
    startSSE()
    // SSE 失败兜底 + 状态同步: 同时低频轮询详情
    startPolling()
  })

  onBeforeUnmount(() => {
    if (es) es.close()
    stopPolling()
  })
</script>
