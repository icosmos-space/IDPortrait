<script setup>
defineProps({
  show: { type: Boolean, default: false },
  aboutInfo: { type: Object, required: true },
})

const emit = defineEmits(['update:show'])
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    title="关于软件"
    style="width: 440px"
    :bordered="false"
    :segmented="{ content: true, footer: 'soft' }"
    @update:show="(v) => emit('update:show', v)"
  >
    <div class="about-body">
      <div class="about-brand">{{ aboutInfo.name }}</div>
      <p class="about-desc">{{ aboutInfo.desc }}</p>
      <div class="about-meta">
        <div><span>版本</span><em>v{{ aboutInfo.version }}</em></div>
        <div><span>作者</span><em>{{ aboutInfo.author }}</em></div>
        <div><span>运行</span><em>本地桌面 · Wails</em></div>
      </div>
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn primary" type="button" @click="emit('update:show', false)">知道了</button>
      </div>
    </template>
  </n-modal>
</template>
