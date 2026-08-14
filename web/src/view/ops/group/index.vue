<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-button
          type="primary"
          icon="plus"
          @click="openDialog()"
        >
          新建分组
        </el-button>
        <el-input
          v-model="searchInfo.name"
          class="!w-44"
          placeholder="分组名称"
          clearable
          @keyup.enter="onSearch"
        />
        <el-select
          v-model="searchInfo.env"
          class="!w-32"
          placeholder="环境"
          clearable
        >
          <el-option
            label="生产 prod"
            value="prod"
          />
          <el-option
            label="预发 staging"
            value="staging"
          />
          <el-option
            label="开发 dev"
            value="dev"
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
          label="名称"
          prop="name"
          min-width="160"
        />
        <el-table-column
          label="环境"
          width="120"
        >
          <template #default="scope">
            <el-tag
              :type="envType(scope.row.env)"
              size="small"
            >
              {{ envLabel(scope.row.env) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="排序"
          prop="sort"
          width="100"
        />
        <el-table-column
          label="描述"
          prop="desc"
          min-width="200"
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

    <el-drawer
      v-model="dialogVisible"
      :title="isEdit ? '编辑分组' : '新建分组'"
      size="440px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
      >
        <el-form-item
          label="名称"
          prop="name"
        >
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="环境">
          <el-select
            v-model="form.env"
            class="!w-full"
          >
            <el-option
              label="生产 prod"
              value="prod"
            />
            <el-option
              label="预发 staging"
              value="staging"
            />
            <el-option
              label="开发 dev"
              value="dev"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number
            v-model="form.sort"
            :min="0"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="form.desc"
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
  getAssetGroupList,
  createAssetGroup,
  updateAssetGroup,
  deleteAssetGroup
} from '@/api/ops'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsGroupList' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchInfo = ref({ name: '', env: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref()
const form = reactive({ ID: 0, name: '', env: 'prod', sort: 0, desc: '' })
const rules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

const envLabel = (e) => ({ prod: '生产', staging: '预发', dev: '开发' }[e] || e)
const envType = (e) => ({ prod: 'danger', staging: 'warning', dev: 'info' }[e] || '')

const getList = async () => {
  const res = await getAssetGroupList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total
  }
}

const onSearch = () => { page.value = 1; getList() }
const handleCurrentChange = (v) => { page.value = v; getList() }
const handleSizeChange = (v) => { pageSize.value = v; getList() }

const openDialog = (row) => {
  isEdit.value = !!row
  Object.assign(form, {
    ID: row ? row.ID : 0,
    name: row ? row.name : '',
    env: row ? row.env : 'prod',
    sort: row ? row.sort : 0,
    desc: row ? row.desc : ''
  })
  dialogVisible.value = true
}

const onSave = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const fn = isEdit.value ? updateAssetGroup : createAssetGroup
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
  ElMessageBox.confirm(`确认删除分组「${row.name}」?`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteAssetGroup({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
    .catch(() => {})
}

getList()
</script>
