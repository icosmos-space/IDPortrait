<script setup>
import { onMounted, ref } from 'vue'
import logoSvg from '../../assets/images/logosvg.svg'

defineProps({
  processing: { type: Boolean, default: false },
})

const emit = defineEmits(['open-image', 'open-camera', 'reset', 'generate', 'export'])

const isWails = ref(false)
const maximised = ref(false)

onMounted(async () => {
  isWails.value = typeof window !== 'undefined' && !!window.runtime
  await refreshMaximised()
})

async function refreshMaximised() {
  if (!window.runtime?.WindowIsMaximised) return
  try {
    maximised.value = await window.runtime.WindowIsMaximised()
  } catch {
    maximised.value = false
  }
}

function minimise() {
  window.runtime?.WindowMinimise?.()
}

async function toggleMaximise() {
  window.runtime?.WindowToggleMaximise?.()
  await refreshMaximised()
}

function closeWindow() {
  window.runtime?.Quit?.()
}
</script>

<template>
  <header class="topbar" @dblclick="toggleMaximise">
    <div class="brand-block">
      <img class="brand-logo" :src="logoSvg" alt="最美证件照" />
      <div class="titles">
        <h1 class="brand">最美证件照</h1>
        <p class="brand-sub">影楼精修 · 合规规格 · 一键出片</p>
      </div>
    </div>

    <div class="actions" @dblclick.stop>
      <button class="btn ghost" type="button" @click="emit('open-image')">打开图片</button>
      <button class="btn ghost" type="button" @click="emit('open-camera')">拍照</button>
      <button class="btn ghost" type="button" @click="emit('reset')">重置</button>
      <button class="btn primary" type="button" :disabled="processing" @click="emit('generate')">
        <span v-if="processing" class="spin" />
        {{ processing ? '精修中…' : '开始生成' }}
      </button>
      <button class="btn ghost" type="button" @click="emit('export')">导出</button>
    </div>

    <div v-if="isWails" class="window-controls" aria-label="窗口控制" @dblclick.stop>
      <button class="win-btn" type="button" title="最小化" aria-label="最小化" @click="minimise">
        <svg viewBox="0 0 12 12" aria-hidden="true"><rect x="1" y="5.5" width="10" height="1" rx="0.5" /></svg>
      </button>
      <button
        class="win-btn"
        type="button"
        :title="maximised ? '还原' : '最大化'"
        :aria-label="maximised ? '还原' : '最大化'"
        @click="toggleMaximise"
      >
        <svg v-if="!maximised" viewBox="0 0 12 12" aria-hidden="true">
          <rect x="1.5" y="1.5" width="9" height="9" fill="none" stroke-width="1" />
        </svg>
        <svg v-else viewBox="0 0 12 12" aria-hidden="true">
          <rect x="3.5" y="1.5" width="7" height="7" fill="none" stroke-width="1" />
          <path d="M1.5 3.5h7v7h-7z" fill="none" stroke-width="1" />
        </svg>
      </button>
      <button class="win-btn win-close" type="button" title="关闭" aria-label="关闭" @click="closeWindow">
        <svg viewBox="0 0 12 12" aria-hidden="true">
          <path d="M2.5 2.5l7 7M9.5 2.5l-7 7" fill="none" stroke-width="1.2" stroke-linecap="round" />
        </svg>
      </button>
    </div>
  </header>
</template>
