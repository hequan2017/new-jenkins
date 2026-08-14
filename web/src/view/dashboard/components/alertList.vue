<template>
  <div>
    <el-table :data="tableData" stripe style="width: 100%">
      <el-table-column label="级别" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="levelType(row.level)" size="small">
            {{ levelText(row.level) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="告警内容" show-overflow-tooltip />
      <el-table-column label="来源" width="90" align="center">
        <template #default="{ row }">
          {{ sourceMap[row.source] || row.source }}
        </template>
      </el-table-column>
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ formatDate(row.CreatedAt) }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { getAlertList } from '@/api/ops'
  import { formatDate } from '@/utils/format'

  defineOptions({
    name: 'GvaAlertList'
  })

  const tableData = ref([])

  const sourceMap = {
    inspect: '巡检',
    ticket: '工单',
    backup: '备份',
    manual: '手动'
  }

  const levelType = (l) =>
    ({ info: 'info', warning: 'warning', critical: 'danger' }[l] || 'info')

  const levelText = (l) =>
    ({ info: '提示', warning: '警告', critical: '严重' }[l] || l)

  onMounted(async () => {
    const res = await getAlertList({ page: 1, pageSize: 5 })
    if (res.code === 0) {
      tableData.value = res.data?.list || []
    }
  })
</script>
