<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-button
          type="primary"
          icon="plus"
          @click="openDialog()"
        >
          新建备份
        </el-button>
        <el-input
          v-model="searchInfo.name"
          class="!w-44"
          placeholder="任务名称"
          clearable
          @keyup.enter="onSearch"
        />
        <el-button
          icon="search"
          @click="onSearch"
        >
          查询
        </el-button>
      </div>

      <el-table
        :data="tableData"
        style="width: 100%"
        row-key="ID"
      >
        <el-table-column
          label="任务名称"
          prop="name"
          min-width="140"
        />
        <el-table-column
          label="资产/凭据"
          width="160"
        >
          <template #default="scope">
            <span>资产#{{ scope.row.assetId }} / 凭据#{{ scope.row.credentialId }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="远程路径"
          prop="remotePath"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          label="cron"
          prop="spec"
          width="140"
        />
        <el-table-column
          label="保留份数"
          prop="keepCount"
          width="90"
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
          label="操作"
          width="400"
          fixed="right"
        >
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="VideoPlay"
              @click="onRun(scope.row)"
            >
              立即备份
            </el-button>
            <el-button
              type="primary"
              link
              @click="onToggle(scope.row)"
            >
              {{ scope.row.enabled ? '停用' : '启用' }}
            </el-button>
            <el-button
              type="primary"
              link
              icon="View"
              @click="openRecords(scope.row)"
            >
              记录
            </el-button>
            <el-button
              type="primary"
              link
              icon="edit"
              @click="openDialog(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              type="primary"
              link
              icon="delete"
              @click="onDelete(scope.row)"
            >
              删除
            </el-button>
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

    <!-- 新建/编辑 -->
    <el-drawer
      v-model="dialogVisible"
      :title="isEdit ? '编辑备份任务' : '新建备份任务'"
      size="520px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item
          label="任务名称"
          prop="name"
        >
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item
          label="资产"
          prop="assetId"
        >
          <el-select
            v-model="form.assetId"
            class="!w-full"
            filterable
          >
            <el-option
              v-for="a in assets"
              :key="a.ID"
              :label="`${a.name} (${a.host})`"
              :value="a.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="凭据"
          prop="credentialId"
        >
          <el-select
            v-model="form.credentialId"
            class="!w-full"
            filterable
          >
            <el-option
              v-for="c in credentials"
              :key="c.ID"
              :label="`${c.name} [${c.username}]`"
              :value="c.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="远程路径"
          prop="remotePath"
        >
          <el-input
            v-model="form.remotePath"
            placeholder="如 /etc/nginx/nginx.conf"
          />
        </el-form-item>
        <el-form-item
          label="cron表达式"
          prop="spec"
        >
          <el-input
            v-model="form.spec"
            placeholder="如 0 2 * * * (每天2点)"
          />
        </el-form-item>
        <el-form-item label="含秒位">
          <el-switch v-model="form.withSeconds" />
        </el-form-item>
        <el-form-item label="保留份数">
          <el-input-number
            v-model="form.keepCount"
            :min="1"
            :max="100"
          />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="saving"
            @click="onSave"
          >
            保存
          </el-button>
        </div>
      </template>
    </el-drawer>

    <!-- 备份记录 -->
    <el-drawer
      v-model="resultVisible"
      :title="`备份记录 - ${resultTask?.name || ''}`"
      size="640px"
    >
      <el-table :data="records">
        <el-table-column
          label="状态"
          width="90"
        >
          <template #default="scope">
            <el-tag
              :type="recordStatusType(scope.row.status)"
              size="small"
            >
              {{ recordStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="大小"
          width="100"
        >
          <template #default="scope">{{ formatSize(scope.row.size) }}</template>
        </el-table-column>
        <el-table-column
          label="时间"
          width="170"
        >
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column
          label="详情/路径"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="scope">{{ scope.row.detail || scope.row.localPath }}</template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="90"
        >
          <template #default="scope">
            <el-button
              v-if="scope.row.status === 'success' && scope.row.localPath"
              type="primary"
              link
              @click="onDownload(scope.row)"
            >
              下载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getBackupTaskList,
  getBackupRecordList,
  createBackupTask,
  updateBackupTask,
  deleteBackupTask,
  toggleBackupTask,
  runBackupTask,
  downloadBackupUrl,
  getAssetList,
  getCredentialList
} from '@/api/ops'
import { useUserStore } from '@/pinia/modules/user'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsBackupList' })

const userStore = useUserStore()

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchInfo = ref({ name: '' })

const assets = ref([])
const credentials = ref([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref()
const form = reactive({
  id: 0,
  name: '',
  assetId: null,
  credentialId: null,
  remotePath: '',
  spec: '0 2 * * *',
  withSeconds: false,
  enabled: true,
  keepCount: 7,
  remark: ''
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  assetId: [{ required: true, message: '请选择资产', trigger: 'change' }],
  credentialId: [{ required: true, message: '请选择凭据', trigger: 'change' }],
  remotePath: [{ required: true, message: '请输入远程路径', trigger: 'blur' }],
  spec: [{ required: true, message: '请输入cron表达式', trigger: 'blur' }]
}

const resultVisible = ref(false)
const resultTask = ref(null)
const records = ref([])

const recordStatusLabel = (s) => ({ running: '执行中', success: '成功', failed: '失败', canceled: '已取消' }[s] || s)
const recordStatusType = (s) => ({ running: 'warning', success: 'success', failed: 'danger', canceled: 'info' }[s] || 'info')

const getList = async () => {
  const res = await getBackupTaskList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total
  }
}

const loadOptions = async () => {
  const [aRes, cRes] = await Promise.all([
    getAssetList({ page: 1, pageSize: 100 }),
    getCredentialList({ page: 1, pageSize: 100 })
  ])
  if (aRes.code === 0) assets.value = aRes.data.list || []
  if (cRes.code === 0) credentials.value = cRes.data.list || []
}

const onSearch = () => { page.value = 1; getList() }
const handleCurrentChange = (v) => { page.value = v; getList() }
const handleSizeChange = (v) => { pageSize.value = v; getList() }

const openDialog = async (row) => {
  await loadOptions()
  isEdit.value = !!row
  Object.assign(form, {
    id: row ? row.ID : 0,
    name: row ? row.name : '',
    assetId: row ? row.assetId : null,
    credentialId: row ? row.credentialId : null,
    remotePath: row ? row.remotePath : '',
    spec: row ? row.spec : '0 2 * * *',
    withSeconds: row ? row.withSeconds : false,
    enabled: row ? row.enabled : true,
    keepCount: row ? row.keepCount : 7,
    remark: row ? row.remark : ''
  })
  dialogVisible.value = true
}

const onSave = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const fn = isEdit.value ? updateBackupTask : createBackupTask
      const res = await fn({ ...form })
      if (res.code === 0) {
        ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
        dialogVisible.value = false
        getList()
      }
    } finally {
      saving.value = false
    }
  })
}

const onToggle = async (row) => {
  const res = await toggleBackupTask({ id: row.ID, enabled: !row.enabled })
  if (res.code === 0) {
    ElMessage.success('操作成功')
    getList()
  }
}

const onRun = async (row) => {
  const res = await runBackupTask({ id: row.ID })
  if (res.code === 0) ElMessage.success('已触发')
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除备份任务「${row.name}」?`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteBackupTask({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
    .catch(() => {})
}

const openRecords = async (row) => {
  resultTask.value = row
  resultVisible.value = true
  const res = await getBackupRecordList({ page: 1, pageSize: 50, taskId: row.ID })
  if (res.code === 0) records.value = res.data.list || []
}

const onDownload = (row) => {
  // 带 token 下载(同源, axios 拦截器已注入 x-token, 直接用 window.open 走 cookie/query)
  const url = `${import.meta.env.VITE_BASE_API || ''}${downloadBackupUrl(row.ID)}&token=${userStore.token}`
  window.open(url, '_blank')
}

const formatSize = (n) => {
  if (!n) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024; i++
  }
  return v.toFixed(i === 0 ? 0 : 1) + ' ' + units[i]
}

getList()
</script>
