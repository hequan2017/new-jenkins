<template>
  <div class="gva-table-box">
    <div class="gva-btn-list flex items-center justify-between">
      <span class="text-base font-medium">
        {{ isEdit ? '编辑流水线' : '新建流水线' }}
      </span>
      <div class="flex gap-2">
        <el-button @click="onBack">返回</el-button>
        <el-button type="primary" @click="onSave">保存</el-button>
      </div>
    </div>

    <el-form :model="form" label-width="100px" class="mt-4">
      <el-form-item label="名称">
        <el-input v-model="form.name" placeholder="流水线名称" class="!w-80" />
      </el-form-item>
      <el-form-item label="触发方式">
        <el-select v-model="form.triggerType" class="!w-40">
          <el-option label="手动" value="manual" />
          <el-option label="定时" value="schedule" />
          <el-option label="Webhook" value="webhook" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="form.triggerType === 'schedule'" label="cron 表达式">
        <el-input
          v-model="form.spec"
          placeholder="如 0 * * * * 或 @hourly"
          class="!w-64"
        />
        <el-checkbox v-model="form.withSeconds" class="ml-3">含秒位</el-checkbox>
      </el-form-item>
      <el-form-item v-if="form.triggerType === 'webhook' && form.webhookSecret" label="Webhook">
        <div class="flex items-center gap-2">
          <el-input
            :model-value="`POST {{baseUrl}}/webhook/trigger/${form.ID || '{id}'}`"
            readonly
            class="!w-[360px]"
          />
          <el-input v-model="form.webhookSecret" readonly class="!w-56" />
          <el-button text @click="copyWebhookSecret">复制密钥</el-button>
        </div>
      </el-form-item>
      <el-form-item label="启用">
        <el-switch v-model="form.enabled" />
      </el-form-item>
      <el-form-item label="说明">
        <el-input v-model="form.description" type="textarea" :rows="2" class="!w-[480px]" />
      </el-form-item>
    </el-form>

    <div class="flex items-center justify-between mt-2 mb-2">
      <span class="text-base font-medium">参数定义</span>
      <el-button type="primary" plain icon="plus" size="small" @click="addParam">添加参数</el-button>
    </div>
    <el-table :data="form.paramFields" size="small" class="mb-4">
      <el-table-column label="参数名" width="140">
        <template #default="{ row }">
          <el-input v-model="row.name" size="small" placeholder="如 env" />
        </template>
      </el-table-column>
      <el-table-column label="展示名" width="140">
        <template #default="{ row }">
          <el-input v-model="row.label" size="small" placeholder="如 环境" />
        </template>
      </el-table-column>
      <el-table-column label="类型" width="110">
        <template #default="{ row }">
          <el-select v-model="row.type" size="small">
            <el-option label="string" value="string" />
            <el-option label="number" value="number" />
            <el-option label="bool" value="bool" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="必填" width="70">
        <template #default="{ row }">
          <el-checkbox v-model="row.required" />
        </template>
      </el-table-column>
      <el-table-column label="默认值" min-width="140">
        <template #default="{ row }">
          <el-input v-model="row.default" size="small" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80">
        <template #default="{ $index }">
          <el-button type="danger" link size="small" icon="delete" @click="form.paramFields.splice($index, 1)">删</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="flex items-center justify-between mt-2 mb-2">
      <span class="text-base font-medium">阶段与步骤</span>
      <el-button type="primary" plain icon="plus" @click="addStage">添加阶段</el-button>
    </div>

    <el-collapse v-model="activeStages">
      <el-collapse-item
        v-for="(stage, si) in form.stages"
        :key="si"
        :name="si"
      >
        <template #title>
          <div class="flex items-center gap-3" @click.stop>
            <span class="text-sm">阶段 {{ si + 1 }}:</span>
            <el-input v-model="stage.name" placeholder="阶段名称" class="!w-40" @click.stop />
            <el-checkbox v-model="stage.approval" @click.stop>需审批</el-checkbox>
            <el-checkbox v-model="stage.continueOnError" @click.stop>失败继续</el-checkbox>
            <el-checkbox v-model="stage.parallel" @click.stop>步骤并行</el-checkbox>
            <el-button type="danger" link icon="delete" @click="removeStage(si)">删阶段</el-button>
          </div>
        </template>

        <el-table :data="stage.steps" row-key="localId" size="small">
          <el-table-column label="顺序" width="70">
            <template #default="{ row }">{{ row.order }}</template>
          </el-table-column>
          <el-table-column label="名称" width="160">
            <template #default="{ row }">
              <el-input v-model="row.name" placeholder="步骤名称" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-select v-model="row.type" size="small">
                <el-option label="HTTP" value="http" />
                <el-option label="Shell" value="shell" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="配置(JSON)" min-width="320">
            <template #default="{ row }">
              <el-input
                v-model="row.configText"
                type="textarea"
                :rows="2"
                :placeholder="configPlaceholder(row.type)"
                size="small"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="{ $index }">
              <el-button type="primary" link size="small" @click="moveStep(stage, $index, -1)">↑</el-button>
              <el-button type="primary" link size="small" @click="moveStep(stage, $index, 1)">↓</el-button>
              <el-button type="danger" link size="small" icon="delete" @click="removeStep(stage, $index)">删</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="mt-2">
          <el-button size="small" plain icon="plus" @click="addStep(stage)">添加步骤</el-button>
        </div>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup>
  import { ref, reactive, computed, onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import {
    createPipeline,
    updatePipeline,
    findPipeline
  } from '@/api/workflow'

  defineOptions({ name: 'WorkflowPipelineEdit' })

  const route = useRoute()
  const router = useRouter()
  const isEdit = computed(() => Number(route.params.id) > 0)

  const form = reactive({
    ID: 0,
    name: '',
    triggerType: 'manual',
    spec: '',
    withSeconds: false,
    webhookSecret: '',
    enabled: true,
    description: '',
    paramFields: [], // 参数定义(前端编辑用, 提交时序列化为 paramSchema)
    stages: []
  })
  const activeStages = ref([])

  const baseUrl = window.location.origin

  // 本地自增 id, 仅供前端 row-key / 排序用, 提交时不发送
  let localIdSeed = 0
  const nextLocalId = () => ++localIdSeed

  const configPlaceholder = (type) =>
    type === 'http'
      ? '{"url":"https://...","method":"GET","timeoutSec":30}'
      : type === 'shell'
        ? '{"command":"echo hello","timeoutSec":600}'
        : '先选择类型'

  const addParam = () => {
    form.paramFields.push({
      name: '',
      label: '',
      type: 'string',
      required: false,
      default: ''
    })
  }

  const copyWebhookSecret = () => {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(form.webhookSecret)
      ElMessage.success('密钥已复制')
    }
  }

  const addStage = () => {
    form.stages.push({
      name: '',
      order: form.stages.length + 1,
      approval: false,
      continueOnError: false,
      parallel: false,
      steps: []
    })
    activeStages.value.push(form.stages.length - 1)
  }

  const removeStage = (idx) => {
    form.stages.splice(idx, 1)
    reindexStages()
  }

  const addStep = (stage) => {
    stage.steps.push({
      localId: nextLocalId(),
      name: '',
      type: 'http',
      configText: '',
      order: stage.steps.length + 1
    })
  }

  const removeStep = (stage, idx) => {
    stage.steps.splice(idx, 1)
    reindexSteps(stage)
  }

  const moveStep = (stage, idx, dir) => {
    const target = idx + dir
    if (target < 0 || target >= stage.steps.length) return
    const arr = stage.steps
    ;[arr[idx], arr[target]] = [arr[target], arr[idx]]
    reindexSteps(stage)
  }

  const reindexStages = () => {
    form.stages.forEach((s, i) => (s.order = i + 1))
  }
  const reindexSteps = (stage) => {
    stage.steps.forEach((s, i) => (s.order = i + 1))
  }

  // 把前端临时字段(configText/localId)转为后端期望的结构
  const buildPayload = () => {
    const stages = form.stages.map((s) => ({
      name: s.name,
      order: s.order,
      approval: !!s.approval,
      continueOnError: !!s.continueOnError,
      parallel: !!s.parallel,
      steps: s.steps.map((sp) => {
        let config = null
        try {
          config = sp.configText ? JSON.parse(sp.configText) : null
        } catch {
          throw new Error(`步骤「${sp.name || ''}」的配置不是合法 JSON`)
        }
        return {
          name: sp.name,
          type: sp.type,
          order: sp.order,
          config
        }
      })
    }))
    return {
      ID: form.ID,
      name: form.name,
      triggerType: form.triggerType,
      spec: form.spec,
      withSeconds: form.withSeconds,
      webhookSecret: form.webhookSecret,
      enabled: form.enabled,
      description: form.description,
      paramSchema: form.paramFields.filter((f) => f.name),
      stages
    }
  }

  const validate = () => {
    if (!form.name) return '流水线名称不能为空'
    for (const s of form.stages) {
      if (!s.name) return '每个阶段必须有名称'
      for (const sp of s.steps) {
        if (!sp.name) return '每个步骤必须有名称'
      }
    }
    return ''
  }

  const onSave = async () => {
    const err = validate()
    if (err) {
      ElMessage.warning(err)
      return
    }
    let payload
    try {
      payload = buildPayload()
    } catch (e) {
      ElMessage.warning(e.message)
      return
    }
    const res = isEdit.value
      ? await updatePipeline(payload)
      : await createPipeline(payload)
    if (res.code === 0) {
      ElMessage.success('保存成功')
      router.push({ name: 'WorkflowPipelineList' })
    }
  }

  const onBack = () => router.push({ name: 'WorkflowPipelineList' })

  const loadDetail = async () => {
    if (!isEdit.value) return
    const res = await findPipeline({ id: route.params.id })
    if (res.code === 0) {
      const d = res.data
      form.ID = d.ID
      form.name = d.name
      form.triggerType = d.triggerType
      form.spec = d.spec || ''
      form.withSeconds = !!d.withSeconds
      form.webhookSecret = d.webhookSecret || ''
      form.enabled = d.enabled
      form.description = d.description
      // paramSchema(JSON) -> paramFields(可编辑数组)
      form.paramFields = (d.paramSchema || []).map((f) => ({
        name: f.name || '',
        label: f.label || '',
        type: f.type || 'string',
        required: !!f.required,
        default: f.default || ''
      }))
      form.stages = (d.stages || []).map((s) => ({
        name: s.name,
        order: s.order,
        approval: s.approval,
        continueOnError: s.continueOnError,
        parallel: !!s.parallel,
        steps: (s.steps || []).map((sp) => ({
          localId: nextLocalId(),
          name: sp.name,
          type: sp.type,
          configText: sp.config ? JSON.stringify(sp.config, null, 2) : '',
          order: sp.order
        }))
      }))
      activeStages.value = form.stages.map((_, i) => i)
    }
  }

  onMounted(loadDetail)
</script>
