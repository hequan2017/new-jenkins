<template>
  <div>
    <div class="gva-table-box">
      <!-- 操作 + 搜索条 -->
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-button
          type="primary"
          icon="plus"
          @click="openDialog()"
        >
          新建资产
        </el-button>
        <el-input
          v-model="searchInfo.name"
          class="!w-48"
          placeholder="资产名称"
          clearable
          @keyup.enter="onSearch"
        />
        <el-select
          v-model="searchInfo.status"
          class="!w-36"
          placeholder="状态"
          clearable
        >
          <el-option
            label="在线"
            value="online"
          />
          <el-option
            label="离线"
            value="offline"
          />
          <el-option
            label="维护中"
            value="maintenance"
          />
        </el-select>
        <el-button
          icon="search"
          @click="onSearch"
        >
          查询
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table
        :data="tableData"
        style="width: 100%"
        row-key="ID"
      >
        <el-table-column
          label="名称"
          prop="name"
          min-width="140"
        />
        <el-table-column
          label="主机"
          min-width="160"
        >
          <template #default="scope">
            <span>{{ scope.row.host }}:{{ scope.row.port }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="操作系统"
          prop="os"
          min-width="120"
        />
        <el-table-column
          label="状态"
          width="100"
        >
          <template #default="scope">
            <el-tag
              :type="assetStatusType(scope.row.status)"
              size="small"
            >
              {{ assetStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="备注"
          prop="remark"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          label="创建时间"
          width="170"
        >
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="180"
          fixed="right"
        >
          <template #default="scope">
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

      <!-- 分页 -->
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

    <!-- 新建 / 编辑抽屉 -->
    <el-drawer
      v-model="dialogVisible"
      :title="isEdit ? '编辑资产' : '新建资产'"
      size="480px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="90px"
      >
        <el-form-item
          label="名称"
          prop="name"
        >
          <el-input
            v-model="form.name"
            placeholder="资产名称"
          />
        </el-form-item>
        <el-form-item
          label="主机"
          prop="host"
        >
          <el-input
            v-model="form.host"
            placeholder="IP 或域名"
          />
        </el-form-item>
        <el-form-item
          label="端口"
          prop="port"
        >
          <el-input-number
            v-model="form.port"
            :min="1"
            :max="65535"
          />
        </el-form-item>
        <el-form-item label="操作系统">
          <el-input
            v-model="form.os"
            placeholder="如 CentOS 7 / Ubuntu 22.04"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="form.status"
            class="!w-full"
          >
            <el-option
              label="在线"
              value="online"
            />
            <el-option
              label="离线"
              value="offline"
            />
            <el-option
              label="维护中"
              value="maintenance"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            v-model="tagsText"
            placeholder="多个标签用逗号分隔"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3"
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
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getAssetList,
  createAsset,
  updateAsset,
  deleteAsset,
  assetStatusType,
  assetStatusLabel
} from '@/api/ops'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsAssetList' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchInfo = ref({ name: '', status: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref()
const tagsText = ref('')
const form = reactive({
  ID: 0,
  name: '',
  host: '',
  port: 22,
  os: '',
  status: 'offline',
  tags: [],
  remark: ''
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  host: [{ required: true, message: '请输入主机', trigger: 'blur' }]
}

const getList = async () => {
  const res = await getAssetList({
    page: page.value,
    pageSize: pageSize.value,
    ...searchInfo.value
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

const openDialog = (row) => {
  isEdit.value = !!row
  if (row) {
    Object.assign(form, {
      ID: row.ID,
      name: row.name,
      host: row.host,
      port: row.port,
      os: row.os,
      status: row.status,
      tags: row.tags || [],
      remark: row.remark
    })
    tagsText.value = Array.isArray(row.tags) ? row.tags.join(',') : ''
  } else {
    Object.assign(form, {
      ID: 0,
      name: '',
      host: '',
      port: 22,
      os: '',
      status: 'offline',
      tags: [],
      remark: ''
    })
    tagsText.value = ''
  }
  dialogVisible.value = true
}

const onSave = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    form.tags = tagsText.value
      ? tagsText.value.split(',').map((s) => s.trim()).filter(Boolean)
      : []
    try {
      const fn = isEdit.value ? updateAsset : createAsset
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

const onDelete = (row) => {
  ElMessageBox.confirm(`确认删除资产「${row.name}」?`, '提示', {
    type: 'warning'
  })
    .then(async () => {
      const res = await deleteAsset({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
    .catch(() => {})
}

getList()
</script>
