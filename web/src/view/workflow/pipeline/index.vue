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
        <el-table-column label="操作" width="420" fixed="right">
          <template #default="scope">
            <el-button type="primary" link icon="VideoPlay" @click="onTrigger(scope.row)">触发</el-button>
            <el-button
              type="primary"
              link
              :icon="scope.row.enabled ? 'CircleClose' : 'Open'"
              @click="onToggle(scope.row)"
            >{{ scope.row.enabled ? '停用' : '启用' }}</el-button
            >
            <el-button type="primary" link icon="edit" @click="openEdit(scope.row.ID)">编辑</el-button>
            <el-button type="primary" link icon="CopyDocument" @click="onClone(scope.row)">克隆</el-button>
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

    <!-- 参数收集对话框: 触发带参数的流水线时弹出 -->
    <el-dialog
      v-model="paramDialogVisible"
      :title="`触发参数 - ${paramNameCache}`"
      width="480px"
    >
      <el-form label-width="120px">
        <el-form-item
          v-for="p in paramForm"
          :key="p.name"
          :label="p.label"
          :required="p.required"
        >
          <el-input
            v-if="p.type !== 'bool'"
            v-model="p.value"
            :placeholder="p.required ? '必填' : ''"
          />
          <el-switch v-else v-model="p.value" active-value="true" inactive-value="false" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="onParamCancel">取消</el-button>
        <el-button type="primary" @click="onParamConfirm">触发</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    getPipelineList,
    deletePipeline,
    triggerBuild,
    findPipeline,
    togglePipeline,
    clonePipeline
  } from '@/api/workflow'
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
    // 查详情看是否定义了参数, 有则弹表单收集
    const dres = await findPipeline({ id: row.ID })
    let params = []
    if (dres.code === 0) {
      const schema = dres.data.paramSchema || []
      if (schema.length > 0) {
        const collected = await collectParams(row.name, schema)
        if (collected === null) return // 用户取消
        params = collected
      }
    }
    const res = await triggerBuild({ pipelineId: row.ID, params })
    if (res.code === 0) {
      ElMessage.success('已触发')
      router.push({ name: 'WorkflowBuildDetail', params: { id: res.data.buildId } })
    }
  }

  // 参数收集对话框: schema 为参数定义数组, 返回 []ParamValue 或 null(取消)
  const paramDialogVisible = ref(false)
  const paramForm = ref([])
  const paramSchemaCache = ref([])
  const paramNameCache = ref('')
  let paramResolve = null
  const collectParams = (name, schema) => {
    paramNameCache.value = name
    paramSchemaCache.value = schema
    paramForm.value = schema.map((f) => ({
      name: f.name,
      label: f.label || f.name,
      value: f.default || '',
      required: !!f.required,
      type: f.type || 'string'
    }))
    return new Promise((resolve) => {
      paramResolve = resolve
      paramDialogVisible.value = true
    })
  }
  const onParamConfirm = () => {
    // 必填校验
    for (const p of paramForm.value) {
      if (p.required && !p.value) {
        ElMessage.warning(`参数「${p.label}」必填`)
        return
      }
    }
    paramDialogVisible.value = false
    if (paramResolve) {
      paramResolve(
        paramForm.value.map((p) => ({ name: p.name, value: String(p.value) }))
      )
      paramResolve = null
    }
  }
  const onParamCancel = () => {
    paramDialogVisible.value = false
    if (paramResolve) {
      paramResolve(null)
      paramResolve = null
    }
  }

  const onToggle = async (row) => {
    const res = await togglePipeline({ id: row.ID, enabled: !row.enabled })
    if (res.code === 0) {
      ElMessage.success('操作成功')
      getList()
    }
  }

  const onClone = async (row) => {
    try {
      const { value: newName } = await ElMessageBox.prompt('请输入新流水线名称', '克隆流水线', {
        inputValue: row.name + '-copy',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      const res = await clonePipeline({ id: row.ID, newName: newName || '' })
      if (res.code === 0) {
        ElMessage.success('克隆成功')
        getList()
      }
    } catch {
      // 用户取消
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
