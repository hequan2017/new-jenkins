<template>
  <div v-loading="loading">
    <!-- 数据卡片 -->
    <div class="flex gap-4 mb-4 flex-wrap">
      <div
        v-for="card in cards"
        :key="card.label"
        class="border rounded p-4 min-w-40 flex-1"
      >
        <div class="text-sm text-gray-500">{{ card.label }}</div>
        <div class="text-2xl font-medium mt-2">{{ card.value }}</div>
        <div
          v-if="card.sub"
          class="text-xs mt-1"
          :class="card.subClass"
        >
          {{ card.sub }}
        </div>
      </div>
    </div>

    <div class="flex gap-4 flex-wrap">
      <!-- 工单状态分布 -->
      <div class="border rounded p-4 flex-1 min-w-80">
        <div class="text-base font-medium mb-3">工单状态分布</div>
        <div
          ref="ticketChart"
          style="height: 280px"
        />
      </div>
      <!-- 近期构建状态 -->
      <div class="border rounded p-4 flex-1 min-w-80">
        <div class="text-base font-medium mb-3">近期构建状态(近20次)</div>
        <div
          ref="buildChart"
          style="height: 280px"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getDashboard } from '@/api/ops'
import { ticketStatusLabel } from '@/api/ops'

defineOptions({ name: 'OpsDashboard' })

const loading = ref(false)
const data = reactive({})
const ticketChart = ref(null)
const buildChart = ref(null)

const cards = ref([])

const buildCards = () => {
  const ticketStat = data.ticketStat || {}
  const ticketTotal = Object.values(ticketStat).reduce((a, b) => a + b, 0)
  const pending = ticketStat.pending || 0
  cards.value = [
    { label: '资产总数', value: data.assetTotal || 0, sub: `在线 ${data.assetOnline || 0}`, subClass: 'text-green-600' },
    { label: '凭据数量', value: data.credentialCnt || 0 },
    { label: '工单总数', value: ticketTotal, sub: `待审批 ${pending}`, subClass: pending ? 'text-orange-500' : 'text-gray-400' },
    { label: '今日审计', value: data.auditToday || 0 },
    { label: '今日巡检异常', value: data.alertToday || 0, subClass: data.alertToday ? 'text-red-500' : 'text-gray-400' }
  ]
}

const renderCharts = () => {
  // 工单状态饼图
  const ticketData = Object.entries(data.ticketStat || {}).map(([k, v]) => ({
    name: ticketStatusLabel(k),
    value: v
  }))
  if (ticketChart.value) {
    const tChart = echarts.init(ticketChart.value)
    tChart.setOption({
      tooltip: { trigger: 'item' },
      legend: { bottom: 0 },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        data: ticketData.length ? ticketData : [{ name: '无数据', value: 1 }]
      }]
    })
  }
  // 构建状态柱图
  const buildData = data.buildRecent || {}
  const buildChartInstance = buildChart.value && echarts.init(buildChart.value)
  if (buildChartInstance) {
    const statusMap = { running: '执行中', success: '成功', failed: '失败', pending: '待执行', canceled: '已取消' }
    buildChartInstance.setOption({
      tooltip: {},
      xAxis: { type: 'category', data: Object.keys(buildData).map((k) => statusMap[k] || k) },
      yAxis: { type: 'value', minInterval: 1 },
      series: [{ type: 'bar', data: Object.values(buildData), itemStyle: { color: '#409eff' } }]
    })
  }
}

const load = async () => {
  loading.value = true
  try {
    const res = await getDashboard()
    if (res.code === 0) {
      Object.assign(data, res.data)
      buildCards()
      await nextTick()
      renderCharts()
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
