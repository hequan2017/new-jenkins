<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-button
          type="primary"
          icon="plus"
          @click="openDialog()"
        >
          新建巡检
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
          label="cron"
          prop="spec"
          width="140"
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
          width="360"
          fixed="right"
        >
          <template #default="scope">
            <el-button
              type="primary"
              link
              icon="VideoPlay"
              @click="onRun(scope.row)"
            >
              立即执行
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
              @click="openResults(scope.row)"
            >
              结果
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
      :title="isEdit ? '编辑巡检' : '新建巡检'"
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
          label="巡检命令"
          prop="command"
        >
          <el-input
            v-model="form.command"
            type="textarea"
            :rows="2"
            placeholder="如 df -h 或 free -m"
          />
        </el-form-item>
        <el-form-item label="异常关键字">
          <el-input
            v-model="form.keyword"
            placeholder="多个用逗号分隔, 命中则判定异常"
          />
        </el-form-item>
        <el-form-item
          label="cron表达式"
          prop="spec"
        >
          <el-input
            v-model="form.spec"
            placeholder="如 @every 5m 或 0 */1 * * *"
          />
        </el-form-item>
        <el-form-item label="含秒位">
          <el-switch v-model="form.withSeconds" />
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

    <!-- 结果抽屉 -->
    <el-drawer
      v-model="resultVisible"
      :title="`巡检结果 - ${resultTask?.name || ''}`"
      size="640px"
    >
      <el-table :data="results">
        <el-table-column
          label="状态"
          width="90"
        >
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'ok' ? 'success' : 'danger'"
              size="small"
            >
              {{ scope.row.status === 'ok' ? '正常' : '异常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="时间"
          width="170"
        >
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column
          label="输出/备注"
          min-width="240"
        >
          <template #default="scope">
            <div class="text-xs whitespace-pre-wrap break-all">
              {{ scope.row.remark || scope.row.output }}
            </div>
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
  getInspectTaskList,
  getInspectResultList,
  createInspectTask,
  updateInspectTask,
  deleteInspectTask,
  toggleInspectTask,
  runInspectTask,
  getAssetList,
  getCredentialList
} from '@/api/ops'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsInspectList' })

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
  command: '',
  keyword: '',
  spec: '@every 5m',
  withSeconds: false,
  enabled: true,
  remark: ''
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  assetId: [{ required: true, message: '请选择资产', trigger: 'change' }],
  credentialId: [{ required: true, message: '请选择凭据', trigger: 'change' }],
  command: [{ required: true, message: '请输入巡检命令', trigger: 'blur' }],
  spec: [{ required: true, message: '请输入cron表达式', trigger: 'blur' }]
}

const resultVisible = ref(false)
const resultTask = ref(null)
const results = ref([])

const getList = async () => {
  const res = await getInspectTaskList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
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

const openDialog = async (row) => {
  await loadOptions()
  isEdit.value = !!row
  Object.assign(form, {
    id: row ? row.ID : 0,
    name: row ? row.name : '',
    assetId: row ? row.assetId : null,
    credentialId: row ? row.credentialId : null,
    command: row ? row.command : '',
    keyword: row ? row.keyword : '',
    spec: row ? row.spec : '@every 5m',
    withSeconds: row ? row.withSeconds : false,
    enabled: row ? row.enabled : true,
    remark: row ? row.remark : ''
  })
  dialogVisible.value = true
}

const onSave = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const fn = isEdit.value ? updateInspectTask : createInspectTask
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
  const res = await toggleInspectTask({ id: row.ID, enabled: !row.enabled })
  if (res.code === 0) {
    ElMessage.success('操作成功')
    getList()
  }
}

const onRun = async (row) => {
  const res = await runInspectTask({ id: row.ID })
  if (res.code === 0) {
    ElMessage.success('已触发, 结果稍后查看')
  }
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除巡检任务「${row.name}」?`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteInspectTask({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
    .catch(() => {})
}

const openResults = async (row) => {
  resultTask.value = row
  resultVisible.value = true
  const res = await getInspectResultList({ page: 1, pageSize: 50, taskId: row.ID })
  if (res.code === 0) results.value = res.data.list || []
}

getList()
</script>
