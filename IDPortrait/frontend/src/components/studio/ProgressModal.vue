<script setup>
import logoSvg from '../../assets/images/logosvg.svg'

defineProps({
  show: { type: Boolean, default: false },
  statusText: { type: String, default: '处理中…' },
  progressPercent: { type: Number, default: 0 },
  variant: { type: String, default: 'card' },
})
</script>

<template>
  <n-modal
    :show="show"
    :mask-closable="false"
    :close-on-esc="false"
    :closable="false"
    transform-origin="center"
  >
    <div class="progress-modal" :class="{ ghost: variant === 'ghost' }" role="status" aria-live="polite">
      <div class="progress-logo-wrap">
        <span class="progress-logo-ring" aria-hidden="true" />
        <img class="progress-logo" :src="logoSvg" alt="" />
      </div>
      <p class="progress-modal-title">{{ statusText || '处理中…' }}</p>
      <div class="progress-modal-bar">
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: `${progressPercent}%` }" />
        </div>
        <span class="progress-num">{{ progressPercent }}%</span>
      </div>
    </div>
  </n-modal>
</template>
