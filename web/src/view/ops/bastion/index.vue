<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center gap-3 mb-4 flex-wrap">
        <span class="text-base font-medium">跳板机</span>
        <el-select
          v-model="assetId"
          class="!w-56"
          placeholder="选择资产"
          filterable
        >
          <el-option
            v-for="a in assets"
            :key="a.ID"
            :label="`${a.name} (${a.host}:${a.port})`"
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
          icon="Connection"
          :disabled="!assetId || !credentialId"
          :loading="testing"
          @click="onTest"
        >
          连接测试
        </el-button>
      </div>

      <!-- 命令执行 -->
      <div class="flex items-center gap-3 mb-3 flex-wrap">
        <el-input
          v-model="command"
          class="!w-[480px]"
          placeholder="输入要执行的命令, 如 uname -a"
          @keyup.enter="onExec"
        />
        <el-button
          type="primary"
          icon="VideoPlay"
          :disabled="!assetId || !credentialId || !command"
          :loading="executing"
          @click="onExec"
        >
          执行命令
        </el-button>
        <el-button
          type="success"
          icon="Monitor"
          :disabled="!assetId || !credentialId"
          @click="openTerminal"
        >
          打开终端
        </el-button>
      </div>

      <el-input
        v-model="output"
        type="textarea"
        :rows="14"
        readonly
        class="font-mono"
        placeholder="命令输出将显示在这里"
      />
    </div>

    <!-- 交互终端对话框 -->
    <el-dialog
      v-model="termVisible"
      title="SSH 终端"
      width="820px"
      @closed="closeTerminal"
    >
      <div
        ref="termWrap"
        class="bg-black rounded"
        style="height: 420px"
      />
      <div class="text-xs text-gray-400 mt-2">
        提示: 终端窗口尺寸自适应, 关闭对话框即断开会话。
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, nextTick, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { getAssetList, getCredentialList, testConnection, execCommand } from '@/api/ops'
import { TerminalSocket } from '@/utils/ws'

defineOptions({ name: 'OpsBastion' })

const assets = ref([])
const credentials = ref([])
const assetId = ref(null)
const credentialId = ref(null)
const command = ref('')
const output = ref('')

const testing = ref(false)
const executing = ref(false)

const termVisible = ref(false)
const termWrap = ref(null)
let term = null
let fitAddon = null
let socket = null

const loadOptions = async () => {
  const [aRes, cRes] = await Promise.all([
    getAssetList({ page: 1, pageSize: 100 }),
    getCredentialList({ page: 1, pageSize: 100 })
  ])
  if (aRes.code === 0) assets.value = aRes.data.list || []
  if (cRes.code === 0) credentials.value = cRes.data.list || []
}

const onTest = async () => {
  testing.value = true
  try {
    const res = await testConnection({
      assetId: assetId.value,
      credentialId: credentialId.value
    })
    if (res.code === 0) {
      ElMessage.success('连接成功')
    }
  } finally {
    testing.value = false
  }
}

const onExec = async () => {
  executing.value = true
  output.value = `$ ${command.value}\n`
  try {
    const res = await execCommand({
      assetId: assetId.value,
      credentialId: credentialId.value,
      command: command.value
    })
    if (res.code === 0) {
      output.value += res.data.output || ''
      if (res.data.error) output.value += `\n[error] ${res.data.error}\n`
    }
  } finally {
    executing.value = false
  }
}

const openTerminal = async () => {
  termVisible.value = true
  await nextTick()
  term = new Terminal({
    theme: { background: '#000000', foreground: '#e5e5e5' },
    fontSize: 13,
    cursorBlink: true
  })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(termWrap.value)
  fitAddon.fit()

  const cols = term.cols
  const rows = term.rows
  socket = new TerminalSocket({
    assetId: assetId.value,
    credentialId: credentialId.value,
    cols,
    rows,
    onOutput: (data) => term && term.write(data),
    onClose: (info) => {
      if (term) term.write(`\r\n\x1b[33m[${info}]\x1b[0m\r\n`)
    },
    onError: (err) => {
      if (term) term.write(`\r\n\x1b[31m[错误] ${err}\x1b[0m\r\n`)
      ElMessage.error(err)
    }
  })
  socket.open()
  term.onData((d) => socket && socket.sendInput(d))
  term.onResize(({ cols: c, rows: r }) => socket && socket.resize(c, r))
}

const closeTerminal = () => {
  if (socket) {
    socket.close()
    socket = null
  }
  if (term) {
    term.dispose()
    term = null
  }
}

onBeforeUnmount(() => {
  closeTerminal()
})

loadOptions()
</script>
