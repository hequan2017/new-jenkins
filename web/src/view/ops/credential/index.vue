<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-button
          type="primary"
          icon="plus"
          @click="openDialog()"
        >
          新建凭据
        </el-button>
        <el-input
          v-model="searchInfo.name"
          class="!w-48"
          placeholder="凭据名称"
          clearable
          @keyup.enter="onSearch"
        />
        <el-select
          v-model="searchInfo.type"
          class="!w-36"
          placeholder="类型"
          clearable
        >
          <el-option
            label="密码"
            value="password"
          />
          <el-option
            label="私钥"
            value="key"
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
          min-width="140"
        />
        <el-table-column
          label="类型"
          width="100"
        >
          <template #default="scope">
            <el-tag size="small">
              {{ scope.row.type === 'key' ? '私钥' : '密码' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="用户名"
          prop="username"
          min-width="120"
        />
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
      :title="isEdit ? '编辑凭据' : '新建凭据'"
      size="480px"
    >
      <el-alert
        v-if="isEdit"
        type="info"
        :closable="false"
        class="mb-4"
        title="密码 / 私钥留空表示不修改"
      />
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
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item
          label="类型"
          prop="type"
        >
          <el-select
            v-model="form.type"
            class="!w-full"
          >
            <el-option
              label="密码"
              value="password"
            />
            <el-option
              label="私钥"
              value="key"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="用户名"
          prop="username"
        >
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="form.secret"
            type="password"
            show-password
            :placeholder="isEdit ? '留空不修改' : '登录密码'"
          />
        </el-form-item>
        <template v-if="form.type === 'key'">
          <el-form-item label="私钥">
            <el-input
              v-model="form.secret"
              type="textarea"
              :rows="6"
              placeholder="PEM 格式私钥, 编辑时留空不修改"
            />
          </el-form-item>
          <el-form-item label="私钥口令">
            <el-input
              v-model="form.passphrase"
              type="password"
              show-password
              placeholder="无口令可留空"
            />
          </el-form-item>
        </template>
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
import { ref, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCredentialList,
  createCredential,
  updateCredential,
  deleteCredential
} from '@/api/ops'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsCredentialList' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchInfo = ref({ name: '', type: '' })

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref()
const form = reactive({
  id: 0,
  name: '',
  type: 'password',
  username: '',
  secret: '',
  passphrase: '',
  remark: ''
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
}

const getList = async () => {
  const res = await getCredentialList({
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
  Object.assign(form, {
    id: row ? row.ID : 0,
    name: row ? row.name : '',
    type: row ? row.type : 'password',
    username: row ? row.username : '',
    secret: '',
    passphrase: '',
    remark: row ? row.remark : ''
  })
  dialogVisible.value = true
}

// 切换类型时清空敏感字段, 避免误用
watch(
  () => form.type,
  () => {
    form.secret = ''
    form.passphrase = ''
  }
)

const onSave = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    if (!isEdit.value && !form.secret) {
      ElMessage.warning('请填写密码或私钥')
      return
    }
    saving.value = true
    try {
      const fn = isEdit.value ? updateCredential : createCredential
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
  ElMessageBox.confirm(`确认删除凭据「${row.name}」?`, '提示', {
    type: 'warning'
  })
    .then(async () => {
      const res = await deleteCredential({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getList()
      }
    })
    .catch(() => {})
}

getList()
</script>
