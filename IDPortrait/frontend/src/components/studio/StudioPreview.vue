<script setup>
import { ref, watch, onMounted } from 'vue'
import { RESULT_TABS, ORIGIN_TABS } from '../../constants/studio'

const props = defineProps({
  originView: { type: Object, default: null },
  mattingView: { type: String, default: null },
  resultView: { type: String, default: null },
  hasOrigin: { type: Boolean, default: false },
  hasMatting: { type: Boolean, default: false },
  hasResult: { type: Boolean, default: false },
  activeOriginTab: { type: String, default: 'original' },
  originTabs: { type: Array, default: () => ORIGIN_TABS },
  activeResultTab: { type: String, default: 'idphoto' },
  resultTabs: { type: Array, default: () => RESULT_TABS },
  currentOriginHint: { type: String, default: '人脸检测' },
  currentResultHint: { type: String, default: '成品预览' },
  diagnostics: { type: Array, default: () => [] },
  beautyStrength: { type: Number, default: 0 },
})

const emit = defineEmits(['drop', 'switch-origin-tab', 'switch-result-tab'])

const originCanvas = ref(null)
const resultCanvas = ref(null)

function paintPlaceholder(canvas, title, subtitle) {
  if (!canvas) return
  canvas.width = 360
  canvas.height = 480
  const ctx = canvas.getContext('2d')
  const g = ctx.createLinearGradient(0, 0, 0, canvas.height)
  g.addColorStop(0, '#f7f9fb')
  g.addColorStop(1, '#e8eef3')
  ctx.fillStyle = g
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  ctx.strokeStyle = 'rgba(24,32,40,0.08)'
  ctx.lineWidth = 1
  ctx.strokeRect(0.5, 0.5, canvas.width - 1, canvas.height - 1)
  ctx.fillStyle = '#8b97a3'
  ctx.font = '600 15px Nunito, sans-serif'
  ctx.textAlign = 'center'
  ctx.fillText(title, canvas.width / 2, canvas.height / 2 - 8)
  ctx.font = '13px Nunito, sans-serif'
  ctx.fillStyle = '#a3adb8'
  ctx.fillText(subtitle, canvas.width / 2, canvas.height / 2 + 16)
}

function paintCheckerboard(ctx, w, h, size = 12) {
  const a = '#eceff3'
  const b = '#f7f9fb'
  for (let y = 0; y < h; y += size) {
    for (let x = 0; x < w; x += size) {
      ctx.fillStyle = ((x / size + y / size) % 2 === 0) ? a : b
      ctx.fillRect(x, y, size, size)
    }
  }
}

function drawOriginSide() {
  const canvas = originCanvas.value
  if (!canvas) return

  if (props.activeOriginTab === 'matting') {
    if (!props.mattingView) {
      paintPlaceholder(
        canvas,
        '等待抠图',
        props.hasOrigin ? '生成后可查看抠图' : '请先打开照片并生成',
      )
      return
    }
    const ctx = canvas.getContext('2d')
    const img = new Image()
    img.onload = () => {
      const maxW = 360
      const maxH = 480
      const scale = Math.min(maxW / img.width, maxH / img.height, 1)
      canvas.width = Math.max(1, Math.round(img.width * scale))
      canvas.height = Math.max(1, Math.round(img.height * scale))
      paintCheckerboard(ctx, canvas.width, canvas.height)
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
    }
    img.onerror = () => {
      paintPlaceholder(canvas, '抠图加载失败', '请重新生成')
    }
    img.src = props.mattingView
    return
  }

  const view = props.originView
  if (!view?.img) {
    paintPlaceholder(canvas, '等待原图', '打开或拖入照片')
    return
  }
  const ctx = canvas.getContext('2d')
  const img = new Image()
  img.onload = () => {
    const maxW = 360
    const maxH = 480
    const scale = Math.min(maxW / img.width, maxH / img.height, 1)
    canvas.width = Math.max(1, Math.round(img.width * scale))
    canvas.height = Math.max(1, Math.round(img.height * scale))
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
    const sx = canvas.width / img.width
    const sy = canvas.height / img.height
    const faceBox = view.faceBox
    if (faceBox?.length === 4) {
      ctx.strokeStyle = '#c45b7a'
      ctx.lineWidth = 2
      ctx.strokeRect(
        faceBox[0] * sx,
        faceBox[1] * sy,
        (faceBox[2] - faceBox[0]) * sx,
        (faceBox[3] - faceBox[1]) * sy,
      )
    }
    const landmarks = view.landmarks
    if (landmarks?.length) {
      ctx.fillStyle = '#c45b7a'
      for (let i = 0; i < landmarks.length; i += 2) {
        ctx.beginPath()
        ctx.arc(landmarks[i] * sx, landmarks[i + 1] * sy, 3, 0, Math.PI * 2)
        ctx.fill()
      }
    }
  }
  img.onerror = () => {
    paintPlaceholder(canvas, '原图加载失败', '请重新选择图片')
  }
  img.src = view.img
}

function drawResult(imgData) {
  const canvas = resultCanvas.value
  if (!canvas) return
  if (!imgData) {
    const tab = props.resultTabs.find((t) => t.key === props.activeResultTab)
    paintPlaceholder(canvas, tab?.label || '成品', props.hasResult ? '暂无该类型成品' : '生成后在此显示')
    return
  }
  const ctx = canvas.getContext('2d')
  const img = new Image()
  img.onload = () => {
    const maxW = props.activeResultTab === 'layout' ? 420 : 360
    const maxH = 480
    const scale = Math.min(maxW / img.width, maxH / img.height, 1)
    canvas.width = Math.max(1, Math.round(img.width * scale))
    canvas.height = Math.max(1, Math.round(img.height * scale))
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
  }
  img.src = imgData
}

onMounted(() => {
  drawOriginSide()
  drawResult(props.resultView)
})

watch(
  () => [props.originView, props.mattingView, props.activeOriginTab, props.hasOrigin, props.hasMatting],
  () => drawOriginSide(),
  { deep: true },
)

watch(
  () => [props.resultView, props.activeResultTab, props.hasResult],
  () => drawResult(props.resultView),
)
</script>

<template>
  <main class="stage">
    <div class="preview-board">
      <div class="board-caption">
        <span>影棚预览</span>
        <em>原图比对 · 成品精修</em>
      </div>
      <div class="preview-row">
        <article class="frame">
          <header class="frame-head">
            <span>原图</span>
            <em>{{ currentOriginHint }}</em>
          </header>
          <div class="result-tabs" role="tablist">
            <button
              v-for="tab in originTabs"
              :key="tab.key"
              type="button"
              role="tab"
              class="result-tab"
              :class="{ active: activeOriginTab === tab.key }"
              :aria-selected="activeOriginTab === tab.key"
              @click="emit('switch-origin-tab', tab.key)"
            >
              {{ tab.label }}
            </button>
          </div>
          <div
            class="frame-body"
            :class="{ matting: activeOriginTab === 'matting' }"
            @dragover.prevent
            @drop.prevent="emit('drop', $event)"
          >
            <canvas ref="originCanvas" />
            <p v-if="!hasOrigin && activeOriginTab === 'original'" class="frame-hint">
              拖放照片到这里，或点「打开图片」
            </p>
            <p v-else-if="activeOriginTab === 'matting' && !hasMatting" class="frame-hint">
              生成完成后可查看抠图结果
            </p>
          </div>
        </article>

        <div class="bridge" aria-hidden="true">
          <span />
        </div>

        <article class="frame result">
          <header class="frame-head">
            <span>成品</span>
            <em>{{ currentResultHint }}</em>
          </header>
          <div class="result-tabs" role="tablist">
            <button
              v-for="tab in resultTabs"
              :key="tab.key"
              type="button"
              role="tab"
              class="result-tab"
              :class="{ active: activeResultTab === tab.key }"
              :aria-selected="activeResultTab === tab.key"
              @click="emit('switch-result-tab', tab.key)"
            >
              {{ tab.label }}
            </button>
          </div>
          <div class="frame-body">
            <canvas ref="resultCanvas" />
            <p v-if="!hasResult" class="frame-hint">生成完成后可切换查看各类成品</p>
          </div>
        </article>
      </div>
    </div>

    <section class="report">
      <div class="report-title">检测诊断</div>
      <div class="report-grid">
        <div
          v-for="item in diagnostics"
          :key="item.key"
          class="report-item"
          :data-ok="item.ok"
        >
          <div class="dot" />
          <div>
            <strong>{{ item.label }}</strong>
            <p>{{ item.text }}</p>
          </div>
        </div>
      </div>
      <p v-if="beautyStrength >= 0.35" class="report-warn">
        美颜强度偏高，可能影响人脸核验通过率
      </p>
    </section>
  </main>
</template>
