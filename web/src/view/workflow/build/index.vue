<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex gap-3 items-center">
        <el-select
          v-model="searchInfo.pipelineId"
          placeholder="流水线"
          class="!w-40"
          clearable
        >
          <el-option
            v-for="p in pipelines"
            :key="p.ID"
            :label="p.name"
            :value="p.ID"
          />
        </el-select>
        <el-select
          v-model="searchInfo.status"
          placeholder="状态"
          class="!w-36"
          clearable
        >
          <el-option
            v-for="(v, k) in buildStatusMap"
            :key="k"
            :label="v.label"
            :value="k"
          />
        </el-select>
        <el-button icon="search" @click="onSearch">查询</el-button>
      </div>

      <el-table :data="tableData" style="width: 100%" row-key="ID">
        <el-table-column label="构建号" prop="buildNo" width="90" />
        <el-table-column label="流水线ID" prop="pipelineId" width="100" />
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="buildStatusType(scope.row.status)" size="small">
              {{ buildStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="触发方式" prop="trigger" width="100" />
        <el-table-column label="触发人" prop="triggerBy" width="100" />
        <el-table-column label="结果/原因" prop="message" min-width="160" show-overflow-tooltip />
        <el-table-column label="开始时间" width="170">
          <template #default="scope">{{ formatDate(scope.row.startedAt) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="170">
          <template #default="scope">{{ formatDate(scope.row.finishedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button type="primary" link icon="view" @click="goDetail(scope.row)">详情</el-button>
            <el-button
              v-if="canCancel(scope.row.status)"
              type="danger"
              link
              icon="CircleClose"
              @click="onCancel(scope.row)"
            >取消</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getBuildList,
    getPipelineList,
    cancelBuild,
    buildStatusMap,
    buildStatusLabel,
    buildStatusType
  } from '@/api/workflow'
  import { formatDate } from '@/utils/format'

  defineOptions({ name: 'WorkflowBuildList' })

  const router = useRouter()
  const route = useRoute()

  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const pipelines = ref([])
  const searchInfo = ref({
    pipelineId: Number(route.query.pipelineId) || 0,
    status: ''
  })

  const canCancel = (status) => ['running', 'pending', 'running-approval'].includes(status)

  const getList = async () => {
    const res = await getBuildList({
      page: page.value,
      pageSize: pageSize.value,
      pipelineId: searchInfo.value.pipelineId,
      status: searchInfo.value.status
    })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total
    }
  }

  const loadPipelines = async () => {
    const res = await getPipelineList({ page: 1, pageSize: 100 })
    if (res.code === 0) {
      pipelines.value = res.data.list || []
    }
  }

  const onSearch = () => {
    page.value = 1
    getList()
  }
  const handleCurrentChange = (v) => {
    page.value = v
    getList()
  }
  const handleSizeChange = (v) => {
    pageSize.value = v
    getList()
  }

  const goDetail = (row) => {
    router.push({ name: 'WorkflowBuildDetail', params: { id: row.ID } })
  }

  const onCancel = async (row) => {
    try {
      await ElMessageBox.confirm(`确定取消构建 #${row.buildNo}?`, '取消确认', { type: 'warning' })
    } catch {
      return
    }
    const res = await cancelBuild({ id: row.ID })
    if (res.code === 0) {
      ElMessage.success('已取消')
      getList()
    }
  }

  onMounted(() => {
    loadPipelines()
    getList()
  })
</script>
