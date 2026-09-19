<script setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { RESULT_TABS, ORIGIN_TABS, SOCIAL_OPTIONS } from '../../constants/studio'

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
  socialView: { type: String, default: '' },
  activeSocial: { type: String, default: 'social' },
  socialHint: { type: String, default: '倾斜相框' },
  diagnostics: { type: Array, default: () => [] },
  beautyStrength: { type: Number, default: 0 },
})

const emit = defineEmits(['drop', 'switch-origin-tab', 'switch-result-tab', 'switch-social'])

const originCanvas = ref(null)
const resultCanvas = ref(null)
const socialCanvas = ref(null)
let resizeObserver = null
const socialSelectOptions = SOCIAL_OPTIONS.map((item) => ({ label: item.label, value: item.key }))

function boxSize(canvas) {
  const box = canvas.parentElement
  if (!box) return { w: 360, h: 480 }
  const style = getComputedStyle(box)
  const padX = (parseFloat(style.paddingLeft) || 0) + (parseFloat(style.paddingRight) || 0)
  const padY = (parseFloat(style.paddingTop) || 0) + (parseFloat(style.paddingBottom) || 0)
  const w = Math.floor(box.clientWidth - padX)
  const h = Math.floor(box.clientHeight - padY)
  if (w < 8 || h < 8) return { w: 360, h: 480 }
  return { w, h }
}

function fitImage(imgW, imgH, maxW, maxH) {
  const scale = Math.min(maxW / imgW, maxH / imgH)
  return {
    w: Math.max(1, Math.round(imgW * scale)),
    h: Math.max(1, Math.round(imgH * scale)),
  }
}

function prepareCanvas(canvas, cssW, cssH) {
  const dpr = window.devicePixelRatio || 1
  canvas.width = Math.max(1, Math.round(cssW * dpr))
  canvas.height = Math.max(1, Math.round(cssH * dpr))
  canvas.style.width = `${cssW}px`
  canvas.style.height = `${cssH}px`
  const ctx = canvas.getContext('2d')
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  return ctx
}

function paintPlaceholder(canvas, title, subtitle) {
  if (!canvas) return
  const { w, h } = boxSize(canvas)
  const ctx = prepareCanvas(canvas, w, h)
  const g = ctx.createLinearGradient(0, 0, 0, h)
  g.addColorStop(0, '#f7f9fb')
  g.addColorStop(1, '#e8eef3')
  ctx.fillStyle = g
  ctx.fillRect(0, 0, w, h)
  ctx.strokeStyle = 'rgba(24,32,40,0.08)'
  ctx.lineWidth = 1
  ctx.strokeRect(0.5, 0.5, w - 1, h - 1)
  ctx.fillStyle = '#8b97a3'
  ctx.font = '600 15px Nunito, sans-serif'
  ctx.textAlign = 'center'
  ctx.fillText(title, w / 2, h / 2 - 8)
  ctx.font = '13px Nunito, sans-serif'
  ctx.fillStyle = '#a3adb8'
  ctx.fillText(subtitle, w / 2, h / 2 + 16)
}

function pointAt(landmarks, index) {
  const i = index * 2
  if (i + 1 >= landmarks.length) return null
  return { x: landmarks[i], y: landmarks[i + 1] }
}

function averagePoints(landmarks, start, end) {
  let x = 0
  let y = 0
  let n = 0
  for (let i = start; i <= end; i += 1) {
    const p = pointAt(landmarks, i)
    if (!p) return null
    x += p.x
    y += p.y
    n += 1
  }
  if (!n) return null
  return { x: x / n, y: y / n }
}

/** 左眼、右眼、鼻尖、左嘴角、右嘴角。68 点取关键索引，已是 5 点则直接用。 */
function fiveKeypoints(landmarks) {
  if (!landmarks?.length) return null
  if (landmarks.length === 10) {
    return [0, 1, 2, 3, 4].map((i) => pointAt(landmarks, i))
  }
  if (landmarks.length < 68 * 2) return null
  const leftEye = averagePoints(landmarks, 36, 41)
  const rightEye = averagePoints(landmarks, 42, 47)
  const nose = pointAt(landmarks, 30)
  const mouthLeft = pointAt(landmarks, 48)
  const mouthRight = pointAt(landmarks, 54)
  if (!leftEye || !rightEye || !nose || !mouthLeft || !mouthRight) return null
  return [leftEye, rightEye, nose, mouthLeft, mouthRight]
}

function drawFiveKeypoints(ctx, landmarks, sx, sy) {
  const pts = fiveKeypoints(landmarks)
  if (!pts) return
  const mapped = pts.map((p) => ({ x: p.x * sx, y: p.y * sy }))
  const links = [
    [0, 1],
    [0, 2],
    [1, 2],
    [2, 3],
    [2, 4],
    [3, 4],
  ]
  ctx.save()
  ctx.strokeStyle = '#c45b7a'
  ctx.fillStyle = '#c45b7a'
  ctx.lineWidth = 1.5
  ctx.lineJoin = 'round'
  ctx.beginPath()
  for (const [a, b] of links) {
    ctx.moveTo(mapped[a].x, mapped[a].y)
    ctx.lineTo(mapped[b].x, mapped[b].y)
  }
  ctx.stroke()
  for (const p of mapped) {
    ctx.beginPath()
    ctx.arc(p.x, p.y, 3, 0, Math.PI * 2)
    ctx.fill()
  }
  ctx.restore()
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
    const img = new Image()
    img.onload = () => {
      const box = boxSize(canvas)
      const fitted = fitImage(img.width, img.height, box.w, box.h)
      const ctx = prepareCanvas(canvas, fitted.w, fitted.h)
      paintCheckerboard(ctx, fitted.w, fitted.h)
      ctx.drawImage(img, 0, 0, fitted.w, fitted.h)
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
  const img = new Image()
  img.onload = () => {
    const box = boxSize(canvas)
    const fitted = fitImage(img.width, img.height, box.w, box.h)
    const ctx = prepareCanvas(canvas, fitted.w, fitted.h)
    ctx.clearRect(0, 0, fitted.w, fitted.h)
    ctx.drawImage(img, 0, 0, fitted.w, fitted.h)
    const sx = fitted.w / img.width
    const sy = fitted.h / img.height
    drawFiveKeypoints(ctx, view.landmarks, sx, sy)
  }
  img.onerror = () => {
    paintPlaceholder(canvas, '原图加载失败', '请重新选择图片')
  }
  img.src = view.img
}

function drawFitted(canvas, imgData, emptyTitle, emptySub) {
  if (!canvas) return
  if (!imgData) {
    paintPlaceholder(canvas, emptyTitle, emptySub)
    return
  }
  const img = new Image()
  img.onload = () => {
    const box = boxSize(canvas)
    const fitted = fitImage(img.width, img.height, box.w, box.h)
    const ctx = prepareCanvas(canvas, fitted.w, fitted.h)
    ctx.clearRect(0, 0, fitted.w, fitted.h)
    ctx.drawImage(img, 0, 0, fitted.w, fitted.h)
  }
  img.src = imgData
}

function drawResult(imgData) {
  const tab = props.resultTabs.find((t) => t.key === props.activeResultTab)
  drawFitted(
    resultCanvas.value,
    imgData,
    tab?.label || '成品',
    props.hasResult ? '暂无该类型成品' : '生成后在此显示',
  )
}

function drawSocial(imgData) {
  const tab = SOCIAL_OPTIONS.find((t) => t.key === props.activeSocial)
  drawFitted(socialCanvas.value, imgData, tab?.label || '社交照', '生成后在此显示')
}

onMounted(() => {
  drawOriginSide()
  drawResult(props.resultView)
  drawSocial(props.socialView)
  const observer = new ResizeObserver(() => {
    drawOriginSide()
    drawResult(props.resultView)
    drawSocial(props.socialView)
  })
  if (originCanvas.value?.parentElement) observer.observe(originCanvas.value.parentElement)
  if (resultCanvas.value?.parentElement) observer.observe(resultCanvas.value.parentElement)
  if (socialCanvas.value?.parentElement) observer.observe(socialCanvas.value.parentElement)
  resizeObserver = observer
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
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

watch(
  () => [props.socialView, props.activeSocial],
  () => drawSocial(props.socialView),
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
          <div class="result-stack">
            <section class="result-pane">
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
            </section>
            <section class="result-pane">
              <header class="frame-head">
                <span>社交照</span>
                <em>{{ socialHint }}</em>
              </header>
              <n-select
                class="social-select"
                size="small"
                :value="activeSocial"
                :options="socialSelectOptions"
                :consistent-menu-width="false"
                @update:value="(value) => emit('switch-social', value)"
              />
              <div class="frame-body">
                <canvas ref="socialCanvas" />
                <p v-if="!socialView" class="frame-hint">生成完成后可查看社交照</p>
              </div>
            </section>
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
