<script setup>
defineProps({
  processing: { type: Boolean, default: false },
  statusText: { type: String, default: '就绪' },
  statusTone: { type: String, default: 'ok' },
  progressPercent: { type: Number, default: 0 },
})

const emit = defineEmits(['open-image', 'open-camera', 'reset', 'generate', 'export'])
</script>

<template>
  <header class="topbar">
    <div class="brand-block">
      <div class="brand-mark" aria-hidden="true" />
      <div class="titles">
        <h1 class="brand">最美证件照</h1>
        <p class="brand-sub">影楼精修 · 合规规格 · 一键出片</p>
      </div>
    </div>

    <div class="actions">
      <button class="btn ghost" type="button" @click="emit('open-image')">打开图片</button>
      <button class="btn ghost" type="button" @click="emit('open-camera')">拍照</button>
      <button class="btn ghost" type="button" @click="emit('reset')">重置</button>
      <button class="btn primary" type="button" :disabled="processing" @click="emit('generate')">
        <span v-if="processing" class="spin" />
        {{ processing ? '精修中…' : '开始生成' }}
      </button>
      <button class="btn ghost" type="button" @click="emit('export')">导出</button>
    </div>

    <div class="status-block">
      <span class="status-pill" :data-tone="statusTone">{{ statusText }}</span>
      <div class="progress-wrap">
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: `${progressPercent}%` }" />
        </div>
        <span class="progress-num">{{ progressPercent }}%</span>
      </div>
    </div>
  </header>
</template>
