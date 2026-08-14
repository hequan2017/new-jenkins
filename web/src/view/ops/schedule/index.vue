<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <span class="text-base font-medium">调度任务中心</span>
        <el-select
          v-model="source"
          class="!w-40"
          placeholder="来源"
          @change="getList"
        >
          <el-option
            label="全部"
            value=""
          />
          <el-option
            label="流水线(workflow)"
            value="workflow"
          />
          <el-option
            label="巡检(inspect)"
            value="inspect"
          />
          <el-option
            label="备份(backup)"
            value="backup"
          />
        </el-select>
        <el-button
          icon="Refresh"
          @click="getList"
        >
          刷新
        </el-button>
        <span class="text-xs text-gray-400 ml-auto">共 {{ total }} 个调度任务</span>
      </div>

      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        row-key="key"
      >
        <el-table-column
          label="来源"
          width="160"
        >
          <template #default="scope">
            <el-tag :type="sourceType(scope.row.source)" size="small">
              {{ sourceLabel(scope.row.source) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="任务名称"
          prop="name"
          min-width="200"
        />
        <el-table-column
          label="cron 表达式"
          prop="spec"
          width="160"
        />
        <el-table-column
          label="关联对象"
          prop="refName"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          label="状态"
          width="100"
        >
          <template #default="scope">
            <el-tag
              :type="scope.row.enabled ? 'success' : 'info'"
              size="small"
            >
              {{ scope.row.enabled ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="说明"
          width="220"
        >
          <template #default>
            <span class="text-xs text-gray-400">各来源任务请到对应模块手动触发/启停</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getScheduleList } from '@/api/ops'

defineOptions({ name: 'OpsSchedule' })

const tableData = ref([])
const total = ref(0)
const source = ref('')
const loading = ref(false)

const sourceLabel = (s) => ({ workflow: '流水线', inspect: '巡检', backup: '备份' }[s] || s)
const sourceType = (s) => ({ workflow: 'primary', inspect: 'warning', backup: 'success' }[s] || '')

const getList = async () => {
  loading.value = true
  try {
    const res = await getScheduleList({ source: source.value })
    if (res.code === 0) {
      const list = res.data.list || []
      // 加 key 防止 row-key 冲突
      tableData.value = list.map((it) => ({ ...it, key: `${it.source}-${it.id}` }))
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

getList()
</script>
