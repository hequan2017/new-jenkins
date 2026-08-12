<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex gap-3">
        <el-button type="primary" icon="plus" @click="openEdit(0)">新建流水线</el-button>
        <el-input
          v-model="searchInfo.name"
          placeholder="流水线名称"
          class="!w-48"
          clearable
          @keyup.enter="onSearch"
        />
        <el-select
          v-model="searchInfo.triggerType"
          placeholder="触发方式"
          class="!w-36"
          clearable
        >
          <el-option label="手动" value="manual" />
          <el-option label="定时" value="schedule" />
          <el-option label="Webhook" value="webhook" />
        </el-select>
        <el-button icon="search" @click="onSearch">查询</el-button>
      </div>

      <el-table :data="tableData" style="width: 100%" row-key="ID">
        <el-table-column label="ID" prop="ID" width="80" />
        <el-table-column label="名称" prop="name" min-width="160" />
        <el-table-column label="触发方式" prop="triggerType" width="110">
          <template #default="scope">
            <el-tag size="small">{{ triggerLabel(scope.row.triggerType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">
              {{ scope.row.enabled ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="说明" prop="description" min-width="160" show-overflow-tooltip />
        <el-table-column label="创建时间" width="170">
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="scope">
            <el-button type="primary" link icon="VideoPlay" @click="onTrigger(scope.row)">触发</el-button>
            <el-button type="primary" link icon="edit" @click="openEdit(scope.row.ID)">编辑</el-button>
            <el-button
              type="primary"
              link
              icon="Histogram"
              @click="goBuildHistory(scope.row)"
            >构建历史</el-button
            >
            <el-button type="primary" link icon="delete" @click="onDelete(scope.row)">删除</el-button>
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
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { getPipelineList, deletePipeline, triggerBuild } from '@/api/workflow'
  import { formatDate } from '@/utils/format'

  defineOptions({ name: 'WorkflowPipelineList' })

  const router = useRouter()
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const searchInfo = ref({ name: '', triggerType: '' })

  const triggerLabel = (t) => ({ manual: '手动', schedule: '定时', webhook: 'Webhook' }[t] || t)

  const getList = async () => {
    const res = await getPipelineList({
      page: page.value,
      pageSize: pageSize.value,
      name: searchInfo.value.name,
      triggerType: searchInfo.value.triggerType
    })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total
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

  const openEdit = (id) => {
    router.push({ name: 'WorkflowPipelineEdit', params: { id } })
  }

  const goBuildHistory = (row) => {
    router.push({ name: 'WorkflowBuildList', query: { pipelineId: row.ID } })
  }

  const onTrigger = async (row) => {
    try {
      await ElMessageBox.confirm(`确定触发流水线「${row.name}」?`, '触发确认', { type: 'info' })
    } catch {
      return
    }
    const res = await triggerBuild({ pipelineId: row.ID, params: [] })
    if (res.code === 0) {
      ElMessage.success('已触发')
      router.push({ name: 'WorkflowBuildDetail', params: { id: res.data.buildId } })
    }
  }

  const onDelete = async (row) => {
    try {
      await ElMessageBox.confirm(`确定删除流水线「${row.name}」?`, '删除确认', { type: 'warning' })
    } catch {
      return
    }
    const res = await deletePipeline({ id: row.ID })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      getList()
    }
  }

  onMounted(getList)
</script>
