import service from '@/utils/request'

// ============================== 资产管理 ==============================

export const createAsset = (data) => {
  return service({
    url: '/ops/createAsset',
    method: 'post',
    data
  })
}

export const updateAsset = (data) => {
  return service({
    url: '/ops/updateAsset',
    method: 'put',
    data
  })
}

export const deleteAsset = (data) => {
  return service({
    url: '/ops/deleteAsset',
    method: 'delete',
    data
  })
}

export const getAssetList = (data) => {
  return service({
    url: '/ops/getAssetList',
    method: 'post',
    data
  })
}

// ============================== 凭据管理 ==============================

export const createCredential = (data) => {
  return service({
    url: '/ops/createCredential',
    method: 'post',
    data
  })
}

export const updateCredential = (data) => {
  return service({
    url: '/ops/updateCredential',
    method: 'put',
    data
  })
}

export const deleteCredential = (data) => {
  return service({
    url: '/ops/deleteCredential',
    method: 'delete',
    data
  })
}

export const getCredentialList = (data) => {
  return service({
    url: '/ops/getCredentialList',
    method: 'post',
    data
  })
}

// ============================== 跳板机 ==============================

export const testConnection = (data) => {
  return service({
    url: '/ops/testConnection',
    method: 'post',
    data
  })
}

export const execCommand = (data) => {
  return service({
    url: '/ops/execCommand',
    method: 'post',
    data
  })
}

// ============================== 工单发版 ==============================

export const createTicket = (data) => {
  return service({
    url: '/ops/createTicket',
    method: 'post',
    data
  })
}

export const deleteTicket = (data) => {
  return service({
    url: '/ops/deleteTicket',
    method: 'delete',
    data
  })
}

export const approveTicket = (data) => {
  return service({
    url: '/ops/approveTicket',
    method: 'post',
    data
  })
}

export const cancelTicket = (data) => {
  return service({
    url: '/ops/cancelTicket',
    method: 'post',
    data
  })
}

export const getTicketList = (data) => {
  return service({
    url: '/ops/getTicketList',
    method: 'post',
    data
  })
}

export const findTicket = (params) => {
  return service({
    url: '/ops/findTicket',
    method: 'get',
    params
  })
}

// ============================== 运维大盘 ==============================

export const getDashboard = (params) => {
  return service({
    url: '/ops/getDashboard',
    method: 'get',
    params
  })
}

// ============================== 审计日志 ==============================

export const getAuditList = (data) => {
  return service({
    url: '/ops/getAuditList',
    method: 'post',
    data
  })
}

// ============================== 巡检任务 ==============================

export const createInspectTask = (data) => {
  return service({
    url: '/ops/createInspectTask',
    method: 'post',
    data
  })
}

export const updateInspectTask = (data) => {
  return service({
    url: '/ops/updateInspectTask',
    method: 'put',
    data
  })
}

export const toggleInspectTask = (data) => {
  return service({
    url: '/ops/toggleInspectTask',
    method: 'post',
    data
  })
}

export const runInspectTask = (data) => {
  return service({
    url: '/ops/runInspectTask',
    method: 'post',
    data
  })
}

export const deleteInspectTask = (data) => {
  return service({
    url: '/ops/deleteInspectTask',
    method: 'delete',
    data
  })
}

export const getInspectTaskList = (data) => {
  return service({
    url: '/ops/getInspectTaskList',
    method: 'post',
    data
  })
}

export const getInspectResultList = (data) => {
  return service({
    url: '/ops/getInspectResultList',
    method: 'post',
    data
  })
}

// ============================== 远程文件管理(SFTP) ==============================

export const listDir = (data, dir) => {
  return service({
    url: '/ops/listDir',
    method: 'post',
    data,
    params: dir ? { dir } : {}
  })
}

export const readFile = (data) => {
  return service({
    url: '/ops/readFile',
    method: 'post',
    data
  })
}

export const writeFile = (data) => {
  return service({
    url: '/ops/writeFile',
    method: 'post',
    data
  })
}

export const removeFile = (data) => {
  return service({
    url: '/ops/removeFile',
    method: 'post',
    data
  })
}

export const renameFile = (data) => {
  return service({
    url: '/ops/renameFile',
    method: 'post',
    data
  })
}

export const mkdir = (data) => {
  return service({
    url: '/ops/mkdir',
    method: 'post',
    data
  })
}

// ============================== 工具:状态映射 ==============================

// 工单状态 -> 显示文案 / Element Plus tag 类型
export const ticketStatusMap = {
  pending: { label: '待审批', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已拒绝', type: 'danger' },
  deploying: { label: '部署中', type: 'warning' },
  success: { label: '部署成功', type: 'success' },
  failed: { label: '部署失败', type: 'danger' },
  canceled: { label: '已取消', type: 'info' }
}

export const ticketStatusLabel = (status) =>
  (ticketStatusMap[status] && ticketStatusMap[status].label) || status

export const ticketStatusType = (status) =>
  (ticketStatusMap[status] && ticketStatusMap[status].type) || 'info'

// 资产状态 -> 显示文案 / Element Plus tag 类型
export const assetStatusMap = {
  online: { label: '在线', type: 'success' },
  offline: { label: '离线', type: 'info' },
  maintenance: { label: '维护中', type: 'warning' }
}

export const assetStatusLabel = (status) =>
  (assetStatusMap[status] && assetStatusMap[status].label) || status

export const assetStatusType = (status) =>
  (assetStatusMap[status] && assetStatusMap[status].type) || 'info'
