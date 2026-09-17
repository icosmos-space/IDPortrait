<script setup>
defineProps({
  show: { type: Boolean, default: false },
  setting: { type: Object, required: true },
})

const emit = defineEmits(['update:show', 'save'])
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    title="软件设置"
    style="width: 460px"
    :bordered="false"
    :segmented="{ content: true, footer: 'soft' }"
    :mask-closable="false"
    @update:show="(v) => emit('update:show', v)"
  >
    <div class="modal-stack">
      <label class="modal-label">模型目录</label>
      <n-input v-model:value="setting.modelDir" placeholder="选择或填写模型目录" />
      <label class="modal-label">缓存目录</label>
      <n-input v-model:value="setting.cacheDir" placeholder="选择或填写缓存目录" />
      <div class="field switch-row">
        <label>屏幕翻拍检测</label>
        <n-switch v-model:value="setting.enableRephotoDetect" size="small" />
      </div>
      <div class="field switch-row">
        <label>AI 图片预检</label>
        <n-switch v-model:value="setting.enableAiDetect" size="small" />
      </div>
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn primary" type="button" @click="emit('save')">保存设置</button>
      </div>
    </template>
  </n-modal>
</template>
