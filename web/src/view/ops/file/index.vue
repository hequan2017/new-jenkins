<template>
  <div>
    <div class="gva-table-box">
      <!-- 资产/凭据选择 + 当前路径 -->
      <div class="gva-btn-list flex items-center gap-3 mb-3 flex-wrap">
        <el-select
          v-model="assetId"
          class="!w-56"
          placeholder="选择资产"
          filterable
          @change="onConnect"
        >
          <el-option
            v-for="a in assets"
            :key="a.ID"
            :label="`${a.name} (${a.host})`"
            :value="a.ID"
          />
        </el-select>
        <el-select
          v-model="credentialId"
          class="!w-48"
          placeholder="选择凭据"
          filterable
        >
          <el-option
            v-for="c in credentials"
            :key="c.ID"
            :label="`${c.name} [${c.username}]`"
            :value="c.ID"
          />
        </el-select>
        <el-button
          icon="Refresh"
          :disabled="!connected"
          @click="loadDir(currentDir)"
        >
          刷新
        </el-button>
        <el-button
          type="primary"
          icon="FolderAdd"
          :disabled="!connected"
          @click="openMkdir"
        >
          新建目录
        </el-button>
        <el-button
          icon="Edit"
          :disabled="!connected"
          @click="openWrite"
        >
          新建文件
        </el-button>
      </div>

      <!-- 面包屑路径 -->
      <div
        v-if="connected"
        class="flex items-center gap-1 mb-3 text-sm flex-wrap"
      >
        <el-button
          link
          type="primary"
          @click="loadDir('/')"
        >
          /
        </el-button>
        <template
          v-for="(seg, idx) in pathSegs"
          :key="idx"
        >
          <span class="text-gray-400">/</span>
          <el-button
            link
            type="primary"
            @click="loadDir(seg.path)"
          >
            {{ seg.name }}
          </el-button>
        </template>
      </div>

      <el-table
        v-loading="loading"
        :data="entries"
        style="width: 100%"
        @row-dblclick="onDblClick"
      >
        <el-table-column
          label="名称"
          min-width="260"
        >
          <template #default="scope">
            <el-icon class="mr-1">
              <Folder v-if="scope.row.isDir" />
              <Document v-else />
            </el-icon>
            <span>{{ scope.row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="大小"
          width="120"
        >
          <template #default="scope">
            <span>{{ scope.row.isDir ? '-' : formatSize(scope.row.size) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="权限"
          prop="mode"
          width="120"
        />
        <el-table-column
          label="修改时间"
          width="200"
        >
          <template #default="scope">{{ formatTime(scope.row.modTime) }}</template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="260"
          fixed="right"
        >
          <template #default="scope">
            <el-button
              v-if="!scope.row.isDir"
              type="primary"
              link
              icon="View"
              @click="onRead(scope.row)"
            >
              查看
            </el-button>
            <el-button
              v-if="!scope.row.isDir"
              type="primary"
              link
              icon="Edit"
              @click="onEdit(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              type="primary"
              link
              icon="Sort"
              @click="onRename(scope.row)"
            >
              重命名
            </el-button>
            <el-button
              type="danger"
              link
              icon="delete"
              @click="onRemove(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 文件查看/编辑 -->
    <el-dialog
      v-model="fileVisible"
      :title="fileEditFlag ? `编辑 ${filePath}` : `查看 ${filePath}`"
      width="780px"
    >
      <el-input
        v-model="fileContent"
        type="textarea"
        :rows="20"
        :readonly="!fileEditFlag"
        class="font-mono"
      />
      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="fileVisible = false">关闭</el-button>
          <el-button
            v-if="fileEditFlag"
            type="primary"
            :loading="fileSaving"
            @click="onWrite"
          >
            保存
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Folder, Document } from '@element-plus/icons-vue'
import {
  getAssetList,
  getCredentialList,
  listDir,
  readFile,
  writeFile,
  removeFile,
  renameFile,
  mkdir
} from '@/api/ops'

defineOptions({ name: 'OpsFile' })

const assets = ref([])
const credentials = ref([])
const assetId = ref(null)
const credentialId = ref(null)
const connected = computed(() => !!assetId.value && !!credentialId.value)

const currentDir = ref('/')
const entries = ref([])
const loading = ref(false)

const pathSegs = computed(() => {
  const parts = currentDir.value.split('/').filter(Boolean)
  let acc = ''
  return parts.map((p) => {
    acc += '/' + p
    return { name: p, path: acc }
  })
})

const loadOptions = async () => {
  const [aRes, cRes] = await Promise.all([
    getAssetList({ page: 1, pageSize: 100 }),
    getCredentialList({ page: 1, pageSize: 100 })
  ])
  if (aRes.code === 0) assets.value = aRes.data.list || []
  if (cRes.code === 0) credentials.value = cRes.data.list || []
}

const loadDir = async (dir) => {
  if (!connected.value) return
  loading.value = true
  try {
    const res = await listDir(
      { assetId: assetId.value, credentialId: credentialId.value },
      dir
    )
    if (res.code === 0) {
      entries.value = res.data.entries || []
      currentDir.value = res.data.dir || dir
    }
  } finally {
    loading.value = false
  }
}

const onConnect = () => {
  if (connected.value) loadDir('/')
}

const onDblClick = (row) => {
  if (row.isDir) {
    const next = currentDir.value.replace(/\/$/, '') + '/' + row.name
    loadDir(next)
  }
}

// 文件查看/编辑
const fileVisible = ref(false)
const filePath = ref('')
const fileContent = ref('')
const fileEditFlag = ref(false)
const fileSaving = ref(false)

const onRead = async (row) => {
  const fp = joinPath(row.name)
  const res = await readFile({ assetId: assetId.value, credentialId: credentialId.value, path: fp })
  if (res.code === 0) {
    filePath.value = fp
    fileContent.value = res.data.content
    fileEditFlag.value = false
    fileVisible.value = true
  }
}

const onEdit = async (row) => {
  await onRead(row)
  fileEditFlag.value = true
}

const openWrite = () => {
  filePath.value = joinPath('newfile.txt')
  fileContent.value = ''
  fileEditFlag.value = true
  fileVisible.value = true
}

const onWrite = async () => {
  fileSaving.value = true
  try {
    const res = await writeFile({
      assetId: assetId.value,
      credentialId: credentialId.value,
      path: filePath.value,
      content: fileContent.value
    })
    if (res.code === 0) {
      ElMessage.success('保存成功')
      fileVisible.value = false
      loadDir(currentDir.value)
    }
  } finally {
    fileSaving.value = false
  }
}

const onRemove = (row) => {
  ElMessageBox.confirm(`确认删除「${row.name}」?`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await removeFile({
        assetId: assetId.value,
        credentialId: credentialId.value,
        path: joinPath(row.name),
        isDir: row.isDir
      })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadDir(currentDir.value)
      }
    })
    .catch(() => {})
}

const onRename = (row) => {
  ElMessageBox.prompt('输入新名称', '重命名', { inputValue: row.name })
    .then(async ({ value }) => {
      const res = await renameFile({
        assetId: assetId.value,
        credentialId: credentialId.value,
        oldPath: joinPath(row.name),
        newPath: joinPath(value)
      })
      if (res.code === 0) {
        ElMessage.success('操作成功')
        loadDir(currentDir.value)
      }
    })
    .catch(() => {})
}

const openMkdir = () => {
  ElMessageBox.prompt('输入目录名', '新建目录', { inputValue: '' })
    .then(async ({ value }) => {
      const res = await mkdir({
        assetId: assetId.value,
        credentialId: credentialId.value,
        dir: joinPath(value)
      })
      if (res.code === 0) {
        ElMessage.success('创建成功')
        loadDir(currentDir.value)
      }
    })
    .catch(() => {})
}

const joinPath = (name) => {
  const base = currentDir.value.replace(/\/$/, '')
  return base + '/' + name
}

const formatSize = (n) => {
  if (!n) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return v.toFixed(i === 0 ? 0 : 1) + ' ' + units[i]
}

const formatTime = (t) => {
  if (!t) return ''
  try {
    return new Date(t).toLocaleString()
  } catch {
    return t
  }
}

loadOptions()
</script>
