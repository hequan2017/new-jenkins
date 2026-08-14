<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <el-select
          v-model="searchInfo.source"
          class="!w-36"
          placeholder="来源"
          clearable
          @change="onSearch"
        >
          <el-option
            v-for="(v, k) in sourceMap"
            :key="k"
            :label="v"
            :value="k"
          />
        </el-select>
        <el-select
          v-model="searchInfo.level"
          class="!w-32"
          placeholder="级别"
          clearable
          @change="onSearch"
        >
          <el-option
            label="信息"
            value="info"
          />
          <el-option
            label="警告"
            value="warning"
          />
          <el-option
            label="严重"
            value="critical"
          />
        </el-select>
        <el-select
          v-model="searchInfo.status"
          class="!w-36"
          placeholder="状态"
          clearable
          @change="onSearch"
        >
          <el-option
            label="未处理"
            value="active"
          />
          <el-option
            label="已处理"
            value="resolved"
          />
          <el-option
            label="已忽略"
            value="ignored"
          />
        </el-select>
        <el-input
          v-model="searchInfo.keyword"
          class="!w-44"
          placeholder="标题/对象"
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
          label="标题"
          prop="title"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column
          label="来源"
          width="110"
        >
          <template #default="scope">
            <el-tag size="small">{{ sourceMap[scope.row.source] || scope.row.source }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="级别"
          width="90"
        >
          <template #default="scope">
            <el-tag
              :type="levelType(scope.row.level)"
              size="small"
            >
              {{ levelLabel(scope.row.level) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="状态"
          width="100"
        >
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'active' ? 'danger' : 'info'"
              size="small"
            >
              {{ statusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="详情"
          prop="detail"
          min-width="240"
          show-overflow-tooltip
        />
        <el-table-column
          label="时间"
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
            <template v-if="scope.row.status === 'active'">
              <el-button
                type="success"
                link
                @click="onHandle(scope.row, 'resolved')"
              >
                标记已处理
              </el-button>
              <el-button
                type="info"
                link
                @click="onHandle(scope.row, 'ignored')"
              >
                忽略
              </el-button>
            </template>
            <span v-else class="text-xs text-gray-400">{{ scope.row.handler }} {{ scope.row.comment }}</span>
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
import { ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import { getAlertList, handleAlert } from '@/api/ops'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'OpsAlertList' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchInfo = ref({ source: '', level: '', status: '', keyword: '' })

const sourceMap = { inspect: '巡检', ticket: '工单', backup: '备份', manual: '手动' }
const levelLabel = (l) => ({ info: '信息', warning: '警告', critical: '严重' }[l] || l)
const levelType = (l) => ({ info: 'info', warning: 'warning', critical: 'danger' }[l] || 'info')
const statusLabel = (s) => ({ active: '未处理', resolved: '已处理', ignored: '已忽略' }[s] || s)

const getList = async () => {
  const res = await getAlertList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total
  }
}

const onSearch = () => { page.value = 1; getList() }
const handleCurrentChange = (v) => { page.value = v; getList() }
const handleSizeChange = (v) => { pageSize.value = v; getList() }

const onHandle = (row, status) => {
  ElMessageBox.prompt('处理意见', status === 'resolved' ? '标记已处理' : '忽略告警', {
    inputValue: ''
  })
    .then(async ({ value }) => {
      const res = await handleAlert({ id: row.ID, status, comment: value || '' })
      if (res.code === 0) {
        getList()
      }
    })
    .catch(() => {})
}

getList()
</script>
