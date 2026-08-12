import service from '@/utils/request'

// ============================== 流水线定义 ==============================

export const createPipeline = (data) => {
  return service({
    url: '/workflow/createPipeline',
    method: 'post',
    data
  })
}

export const updatePipeline = (data) => {
  return service({
    url: '/workflow/updatePipeline',
    method: 'put',
    data
  })
}

export const deletePipeline = (data) => {
  return service({
    url: '/workflow/deletePipeline',
    method: 'delete',
    data
  })
}

export const togglePipeline = (data) => {
  return service({
    url: '/workflow/togglePipeline',
    method: 'post',
    data
  })
}

export const clonePipeline = (data) => {
  return service({
    url: '/workflow/clonePipeline',
    method: 'post',
    data
  })
}

export const getPipelineList = (data) => {
  return service({
    url: '/workflow/getPipelineList',
    method: 'post',
    data
  })
}

export const findPipeline = (params) => {
  return service({
    url: '/workflow/findPipeline',
    method: 'get',
    params
  })
}

// ============================== 构建 ==============================

export const triggerBuild = (data) => {
  return service({
    url: '/workflow/triggerBuild',
    method: 'post',
    data
  })
}

export const cancelBuild = (data) => {
  return service({
    url: '/workflow/cancelBuild',
    method: 'post',
    data
  })
}

export const retryBuild = (data) => {
  return service({
    url: '/workflow/retryBuild',
    method: 'post',
    data
  })
}

export const approveStage = (data) => {
  return service({
    url: '/workflow/approveStage',
    method: 'post',
    data
  })
}

export const getBuildList = (data) => {
  return service({
    url: '/workflow/getBuildList',
    method: 'post',
    data
  })
}

export const getBuildDetail = (params) => {
  return service({
    url: '/workflow/getBuildDetail',
    method: 'get',
    params
  })
}

export const getBuildLog = (params) => {
  return service({
    url: '/workflow/getBuildLog',
    method: 'get',
    params
  })
}

// ============================== 工具:状态映射 ==============================

// 构建状态 -> 显示文案 / Element Plus tag 类型
export const buildStatusMap = {
  pending: { label: '待执行', type: 'info' },
  running: { label: '执行中', type: 'warning' },
  'running-approval': { label: '待审批', type: 'warning' },
  success: { label: '成功', type: 'success' },
  failed: { label: '失败', type: 'danger' },
  canceled: { label: '已取消', type: 'info' }
}

export const buildStatusLabel = (status) =>
  (buildStatusMap[status] && buildStatusMap[status].label) || status

export const buildStatusType = (status) =>
  (buildStatusMap[status] && buildStatusMap[status].type) || 'info'
