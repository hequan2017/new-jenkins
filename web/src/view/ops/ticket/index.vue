<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-button
          type="primary"
          icon="plus"
          @click="openCreate"
        >
          申请工单
        </el-button>
        <el-select
          v-model="searchInfo.scope"
          class="!w-36"
          placeholder="范围"
          @change="onSearch"
        >
          <el-option
            label="全部工单"
            value="all"
          />
          <el-option
            label="我的申请"
            value="mine"
          />
        </el-select>
        <el-select
          v-model="searchInfo.status"
          class="!w-32"
          placeholder="状态"
          clearable
          @change="onSearch"
        >
          <el-option
            v-for="(v, k) in ticketStatusMap"
            :key="k"
            :label="v.label"
            :value="k"
          />
        </el-select>
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
          label="标题"
          prop="title"
          min-width="180"
        />
        <el-table-column
          label="流水线"
          width="120"
        >
          <template #default="scope">#{{ scope.row.pipelineId }}</template>
        </el-table-column>
        <el-table-column
          label="状态"
          width="110"
        >
          <template #default="scope">
            <el-tag
              :type="ticketStatusType(scope.row.status)"
              size="small"
            >
              {{ ticketStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="申请人"
          width="100"
        >
          <template #default="scope">#{{ scope.row.applicantId }}</template>
        </el-table-column>
        <el-table-column
          label="构建"
          width="100"
        >
          <template #default="scope">
            <el-button
              v-if="scope.row.buildId"
              type="primary"
              link
              @click="goBuild(scope.row.buildId)"
            >
              #{{ scope.row.buildId }}
            </el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="申请时间"
          width="170"
        >
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="220"
          fixed="right"
        >
          <template #default="scope">
            <el-button
              v-if="scope.row.status === 'pending'"
              type="success"
              link
              icon="Select"
              @click="openApprove(scope.row, true)"
            >
              通过
            </el-button>
            <el-button
              v-if="scope.row.status === 'pending'"
              type="danger"
              link
              icon="CloseBold"
              @click="openApprove(scope.row, false)"
            >
              拒绝
            </el-button>
            <el-button
              v-if="scope.row.status === 'pending'"
              type="primary"
              link
              icon="CircleClose"
              @click="onCancel(scope.row)"
            >
              取消
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

    <!-- 申请工单 -->
    <el-drawer
      v-model="createVisible"
      title="申请发版工单"
      size="520px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="90px"
      >
        <el-form-item
          label="标题"
          prop="title"
        >
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item
          label="流水线"
          prop="pipelineId"
        >
          <el-select
            v-model="form.pipelineId"
            class="!w-full"
            filterable
            placeholder="选择要发版的流水线"
            @change="onPipelineChange"
          >
            <el-option
              v-for="p in pipelines"
              :key="p.ID"
              :label="p.name"
              :value="p.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          v-if="pipelineParams.length"
          label="参数"
        >
          <div class="flex flex-col gap-2 w-full">
            <div
              v-for="(f, idx) in pipelineParams"
              :key="idx"
              class="flex items-center gap-2"
            >
              <span class="text-sm w-32 truncate">{{ f.label || f.name }}</span>
              <el-input
                v-if="f.type !== 'bool'"
                v-model="paramValues[f.name]"
                class="flex-1"
                :placeholder="f.default || ''"
              />
              <el-switch
                v-else
                v-model="paramValues[f.name]"
              />
            </div>
          </div>
        </el-form-item>
        <el-form-item label="申请说明">
          <el-input
            v-model="form.applyReason"
            type="textarea"
            :rows="3"
          />
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
          <el-button @click="createVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="saving"
            @click="onSave"
          >
            提交申请
          </el-button>
        </div>
      </template>
    </el-drawer>

    <!-- 审批意见 -->
    <el-dialog
      v-model="approveVisible"
      :title="approveAction ? '通过工单' : '拒绝工单'"
      width="460px"
    >
      <el-form label-width="80px">
        <el-form-item label="审批意见">
          <el-input
            v-model="approveComment"
            type="textarea"
            :rows="3"
            :placeholder="approveAction ? '通过后将自动触发流水线构建' : '请填写拒绝原因'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="approveVisible = false">取消</el-button>
          <el-button
            :type="approveAction ? 'success' : 'danger'"
            :loading="approving"
            @click="onApprove"
          >
            确认
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getTicketList,
  createTicket,
  deleteTicket,
  approveTicket,
  cancelTicket,
  ticketStatusMap,
  ticketStatusType,
  ticketStatusLabel
} from '@/api/ops'
import { getPipelineList, findPipeline } from '@/api/workflow'
import { useUserStore } from '@/pinia/modules/user'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsTicketList' })

const router = useRouter()
const userStore = useUserStore()

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchInfo = ref({ scope: 'all', status: '' })

const createVisible = ref(false)
const saving = ref(false)
const formRef = ref()
const pipelines = ref([])
const pipelineParams = ref([])
const paramValues = reactive({})
const form = reactive({
  title: '',
  pipelineId: null,
  applyReason: '',
  remark: ''
})
const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  pipelineId: [{ required: true, message: '请选择流水线', trigger: 'change' }]
}

const approveVisible = ref(false)
const approving = ref(false)
const approveAction = ref(true)
const approveComment = ref('')
const currentTicket = ref(null)

const getList = async () => {
  const params = {
    page: page.value,
    pageSize: pageSize.value,
    status: searchInfo.value.status
  }
  if (searchInfo.value.scope === 'mine') {
    params.applicantId = userStore.userInfo.ID || userStore.userInfo.id
  }
  const res = await getTicketList(params)
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

const loadPipelines = async () => {
  const res = await getPipelineList({ page: 1, pageSize: 100 })
  if (res.code === 0) pipelines.value = (res.data.list || []).filter((p) => p.enabled)
}

const openCreate = async () => {
  await loadPipelines()
  Object.assign(form, { title: '', pipelineId: null, applyReason: '', remark: '' })
  pipelineParams.value = []
  Object.keys(paramValues).forEach((k) => delete paramValues[k])
  createVisible.value = true
}

const onPipelineChange = async (id) => {
  Object.keys(paramValues).forEach((k) => delete paramValues[k])
  if (!id) {
    pipelineParams.value = []
    return
  }
  const res = await findPipeline({ id })
  if (res.code === 0 && res.data) {
    const schema = res.data.paramSchema || []
    pipelineParams.value = schema
    schema.forEach((f) => {
      paramValues[f.name] = f.type === 'bool' ? f.default === 'true' : f.default || ''
    })
  }
}

const onSave = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const params = pipelineParams.value.map((f) => ({
        name: f.name,
        value: String(paramValues[f.name] ?? '')
      }))
      const res = await createTicket({ ...form, params })
      if (res.code === 0) {
        ElMessage.success('已提交, 等待审批')
        createVisible.value = false
        getList()
      }
    } finally {
      saving.value = false
    }
  })
}

const openApprove = (row, action) => {
  currentTicket.value = row
  approveAction.value = action
  approveComment.value = ''
  approveVisible.value = true
}

const onApprove = async () => {
  approving.value = true
  try {
    const res = await approveTicket({
      id: currentTicket.value.ID,
      approve: approveAction.value,
      comment: approveComment.value
    })
    if (res.code === 0) {
      ElMessage.success(approveAction.value ? '已通过并触发流水线' : '已拒绝')
      approveVisible.value = false
      getList()
    }
  } finally {
    approving.value = false
  }
}

const onCancel = (row) => {
  ElMessageBox.confirm(`确认取消工单「${row.title}」?`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await cancelTicket({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已取消')
        getList()
      }
    })
    .catch(() => {})
}

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除工单「${row.title}」?`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteTicket({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
    .catch(() => {})
}

const goBuild = (id) => {
  router.push({ name: 'WorkflowBuildDetail', params: { id } })
}

getList()
</script>
