<template>
  <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-4">
    <div
      v-for="card in cards"
      :key="card.label"
      class="rounded-xl border border-slate-200/80 bg-white px-5 py-4 shadow-sm dark:border-slate-700 dark:bg-slate-900"
    >
      <div class="flex items-center justify-between">
        <div class="text-xs tracking-wide text-muted-foreground">
          {{ card.label }}
        </div>
        <el-icon class="text-lg text-muted-foreground">
          <component :is="card.icon" />
        </el-icon>
      </div>
      <div class="mt-3 text-2xl font-mono text-base-text">
        {{ card.value }}
      </div>
      <div class="mt-1 text-xs text-muted-foreground">
        {{ card.caption }}
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { Monitor, DataLine, Bell, Tickets } from '@element-plus/icons-vue'
  import { getDashboard } from '@/api/ops'

  defineOptions({
    name: 'GvaStatCards'
  })

  const dash = ref(null)

  const load = async () => {
    const res = await getDashboard()
    if (res.code === 0) {
      dash.value = res.data || {}
    }
  }

  onMounted(load)

  const cards = computed(() => {
    const d = dash.value || {}
    const build = d.buildRecent || {}
    const ticket = d.ticketStat || {}
    const assetTotal = d.assetTotal ?? 0
    const assetOnline = d.assetOnline ?? 0
    const success = build.success ?? 0
    const failed = build.failed ?? 0
    return [
      {
        label: '资产在线',
        icon: Monitor,
        value: `${assetOnline} / ${assetTotal}`,
        caption: '在线资产 / 资产总数'
      },
      {
        label: '近期构建',
        icon: DataLine,
        value: success + failed,
        caption: `成功 ${success} · 失败 ${failed}`
      },
      {
        label: '活跃告警',
        icon: Bell,
        value: d.alertActive ?? 0,
        caption: `今日新增 ${d.alertToday ?? 0}`
      },
      {
        label: '工单动态',
        icon: Tickets,
        value: (ticket.pending ?? 0) + (ticket.deploying ?? 0),
        caption: `待审批 ${ticket.pending ?? 0} · 发布中 ${ticket.deploying ?? 0}`
      }
    ]
  })
</script>
