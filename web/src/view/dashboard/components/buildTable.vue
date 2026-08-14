<template>
  <div>
    <el-table :data="tableData" stripe style="width: 100%">
      <el-table-column prop="pipelineName" label="流水线" show-overflow-tooltip />
      <el-table-column label="构建号" width="90" align="center">
        <template #default="{ row }">#{{ row.buildNo }}</template>
      </el-table-column>
      <el-table-column label="状态" width="130" align="center">
        <template #default="{ row }">
          <el-tag :type="buildStatusType(row.status)" size="small">
            {{ buildStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="trigger" label="触发" width="90" align="center" />
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ formatDate(row.CreatedAt) }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import {
    getBuildList,
    getPipelineList,
    buildStatusLabel,
    buildStatusType
  } from '@/api/workflow'
  import { formatDate } from '@/utils/format'

  defineOptions({
    name: 'GvaBuildTable'
  })

  const tableData = ref([])

  onMounted(async () => {
    const [buildRes, pipelineRes] = await Promise.all([
      getBuildList({ page: 1, pageSize: 6 }),
      getPipelineList({ page: 1, pageSize: 100 })
    ])
    if (buildRes.code !== 0) {
      return
    }
    const nameMap = {}
    if (pipelineRes.code === 0) {
      const pipelines = pipelineRes.data?.list || []
      pipelines.forEach((p) => {
        nameMap[p.ID] = p.name
      })
    }
    tableData.value = (buildRes.data?.list || []).map((b) => ({
      ...b,
      pipelineName: nameMap[b.pipelineId] || `#${b.pipelineId}`
    }))
  })
</script>
