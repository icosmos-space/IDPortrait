<script setup>
defineProps({
  show: { type: Boolean, default: false },
  exportOpt: { type: Object, required: true },
  exportDir: { type: String, default: '' },
})

const emit = defineEmits(['update:show', 'select-dir', 'export'])
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    title="导出证件照"
    style="width: 480px"
    :bordered="false"
    :segmented="{ content: true, footer: 'soft' }"
    :mask-closable="false"
    @update:show="(v) => emit('update:show', v)"
  >
    <div class="modal-stack">
      <n-checkbox v-model:checked="exportOpt.single">单张照片</n-checkbox>
      <n-checkbox v-model:checked="exportOpt.idphoto">证件照</n-checkbox>
      <n-checkbox v-model:checked="exportOpt.layout">排版照</n-checkbox>
      <n-checkbox v-model:checked="exportOpt.social">社交照</n-checkbox>
      <label class="modal-label">输出目录</label>
      <n-input :value="exportDir" readonly placeholder="请选择导出目录" />
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn ghost" type="button" @click="emit('select-dir')">选择目录</button>
        <button class="btn primary" type="button" @click="emit('export')">开始导出</button>
      </div>
    </template>
  </n-modal>
</template>
