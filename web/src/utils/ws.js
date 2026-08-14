// WebSocket 终端客户端: 封装与后端 /ops/terminal 的连接、收发、重连与清理。
// token 通过 query 传递(WS 握手无法自定义 header), 后端 GetToken 支持 query 回退。
import { useUserStore } from '@/pinia/modules/user'

const buildWsUrl = (path, params = {}) => {
  const base = import.meta.env.VITE_BASE_API || ''
  // 把 http(s) 前缀转成 ws(s)
  let wsBase = base.replace(/^http/i, 'ws')
  if (!wsBase) {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    wsBase = `${proto}//${window.location.host}`
  }
  const userStore = useUserStore()
  const query = new URLSearchParams({ token: userStore.token || '', ...params })
  return `${wsBase}${path}?${query.toString()}`
}

// TerminalSocket 终端 WS 封装
export class TerminalSocket {
  constructor({ assetId, credentialId, cols, rows, onOutput, onClose, onError }) {
    this.url = buildWsUrl('/ops/terminal', { assetId, credentialId, cols, rows })
    this.onOutput = onOutput || (() => {})
    this.onClose = onClose || (() => {})
    this.onError = onError || (() => {})
    this.ws = null
  }

  open() {
    this.ws = new WebSocket(this.url)
    this.ws.onmessage = (evt) => {
      try {
        const msg = JSON.parse(evt.data)
        if (msg.type === 'output') {
          this.onOutput(msg.data)
        } else if (msg.type === 'close') {
          this.onClose(msg.data)
        } else if (msg.type === 'error') {
          this.onError(msg.data)
        }
      } catch (e) {
        // 非 JSON, 直接当输出
        this.onOutput(evt.data)
      }
    }
    this.ws.onclose = () => this.onClose('连接已关闭')
    this.ws.onerror = () => this.onError('连接错误')
  }

  sendInput(data) {
    this._send({ type: 'input', data })
  }

  resize(cols, rows) {
    this._send({ type: 'resize', cols, rows })
  }

  _send(msg) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
    }
  }

  close() {
    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
      this.ws = null
    }
  }
}
