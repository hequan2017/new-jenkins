<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-select
          v-model="searchInfo.action"
          class="!w-44"
          placeholder="动作类型"
          clearable
        >
          <el-option
            v-for="(v, k) in actionMap"
            :key="k"
            :label="v"
            :value="k"
          />
        </el-select>
        <el-select
          v-model="searchInfo.status"
          class="!w-32"
          placeholder="结果"
          clearable
        >
          <el-option
            label="成功"
            value="success"
          />
          <el-option
            label="失败"
            value="failed"
          />
        </el-select>
        <el-input
          v-model="searchInfo.keyword"
          class="!w-48"
          placeholder="操作人/对象/详情"
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
          label="操作人"
          width="140"
        >
          <template #default="scope">
            <span>{{ scope.row.operator || '#' + scope.row.operatorId }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="动作"
          width="140"
        >
          <template #default="scope">
            <el-tag size="small">{{ actionMap[scope.row.action] || scope.row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="对象"
          prop="target"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          label="来源IP"
          prop="ip"
          width="140"
        />
        <el-table-column
          label="结果"
          width="90"
        >
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'success' ? 'success' : 'danger'"
              size="small"
            >
              {{ scope.row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="详情"
          prop="detail"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          label="时间"
          width="170"
        >
          <template #default="scope">{{ formatDate(scope.row.CreatedAt) }}</template>
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
import { ref } from 'vue'
import { getAuditList } from '@/api/ops'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsAuditList' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchInfo = ref({ action: '', status: '', keyword: '' })

const actionMap = {
  login: '登录',
  cmd_exec: '命令执行',
  terminal: '终端会话',
  file_op: '文件操作',
  ticket_op: '工单操作',
  inspect: '巡检',
  asset_op: '资产操作',
  credential_op: '凭据操作'
}

const getList = async () => {
  const res = await getAuditList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
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

getList()
</script>
