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
  showFaceLandmarks: { type: Boolean, default: true },
})

const emit = defineEmits(['drop', 'switch-origin-tab', 'switch-result-tab', 'switch-social', 'update:showFaceLandmarks'])

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

/** 五官关键点，附中文标签。YuNet 五点：右眼、左眼、鼻尖、右嘴角、左嘴角。 */
function fiveKeypoints(landmarks) {
  if (!landmarks?.length) return null
  if (landmarks.length === 10) {
    const labels = ['右眼', '左眼', '鼻尖', '右嘴角', '左嘴角']
    return labels.map((label, i) => {
      const p = pointAt(landmarks, i)
      return p ? { ...p, label } : null
    })
  }
  if (landmarks.length < 68 * 2) return null
  const leftEye = averagePoints(landmarks, 36, 41)
  const rightEye = averagePoints(landmarks, 42, 47)
  const nose = pointAt(landmarks, 30)
  const mouthLeft = pointAt(landmarks, 48)
  const mouthRight = pointAt(landmarks, 54)
  if (!leftEye || !rightEye || !nose || !mouthLeft || !mouthRight) return null
  return [
    { ...leftEye, label: '左眼' },
    { ...rightEye, label: '右眼' },
    { ...nose, label: '鼻尖' },
    { ...mouthLeft, label: '左嘴角' },
    { ...mouthRight, label: '右嘴角' },
  ]
}

function drawFiveKeypoints(ctx, landmarks, sx, sy) {
  const pts = fiveKeypoints(landmarks)
  if (!pts?.every(Boolean)) return
  const mapped = pts.map((p) => ({ x: p.x * sx, y: p.y * sy, label: p.label }))
  const eyeSpan = Math.hypot(mapped[0].x - mapped[1].x, mapped[0].y - mapped[1].y) || 48
  const noseArm = Math.max(20, Math.min(48, eyeSpan * 0.58))
  const sideArm = Math.max(8, Math.min(18, eyeSpan * 0.2))
  const lineW = Math.max(0.65, Math.min(1.05, eyeSpan / 80))
  const fontPx = Math.max(9, Math.min(12, eyeSpan / 9))
  const gap = Math.max(5, fontPx * 0.4)
  const arms = [sideArm, sideArm, noseArm, sideArm, sideArm]
  const cx = mapped.reduce((s, p) => s + p.x, 0) / mapped.length
  const cy = mapped.reduce((s, p) => s + p.y, 0) / mapped.length

  const neon = 'rgba(64, 224, 255, 0.92)'
  const neonDim = 'rgba(64, 224, 255, 0.55)'
  const glow = 'rgba(0, 200, 255, 0.55)'

  ctx.save()
  ctx.lineCap = 'butt'
  ctx.lineJoin = 'miter'
  ctx.font = `500 ${fontPx}px "Cascadia Mono", "JetBrains Mono", "Consolas", "Microsoft YaHei", monospace`
  ctx.textBaseline = 'middle'
  ctx.shadowColor = glow
  ctx.shadowBlur = Math.max(3, lineW * 4)

  // Sci-fi targeting reticle: segmented cross + hollow core + end ticks.
  const drawReticle = (x, y, arm, heavy) => {
    const core = Math.max(2.2, arm * 0.16)
    const tick = Math.max(2.5, arm * 0.22)
    ctx.strokeStyle = neon
    ctx.lineWidth = heavy ? lineW * 1.15 : lineW

    ctx.beginPath()
    // Horizontal segments (gap at center).
    ctx.moveTo(x - arm, y)
    ctx.lineTo(x - core, y)
    ctx.moveTo(x + core, y)
    ctx.lineTo(x + arm, y)
    // Vertical segments.
    ctx.moveTo(x, y - arm)
    ctx.lineTo(x, y - core)
    ctx.moveTo(x, y + core)
    ctx.lineTo(x, y + arm)
    ctx.stroke()

    // End caps (small perpendicular ticks).
    ctx.beginPath()
    ctx.moveTo(x - arm, y - tick * 0.55)
    ctx.lineTo(x - arm, y + tick * 0.55)
    ctx.moveTo(x + arm, y - tick * 0.55)
    ctx.lineTo(x + arm, y + tick * 0.55)
    ctx.moveTo(x - tick * 0.55, y - arm)
    ctx.lineTo(x + tick * 0.55, y - arm)
    ctx.moveTo(x - tick * 0.55, y + arm)
    ctx.lineTo(x + tick * 0.55, y + arm)
    ctx.stroke()

    // Hollow diamond core.
    const d = core * 0.72
    ctx.beginPath()
    ctx.moveTo(x, y - d)
    ctx.lineTo(x + d, y)
    ctx.lineTo(x, y + d)
    ctx.lineTo(x - d, y)
    ctx.closePath()
    ctx.strokeStyle = neonDim
    ctx.stroke()

    if (heavy) {
      // Nose: outer dashed ring for “scan lock” feel.
      const rr = arm * 0.72
      ctx.setLineDash([Math.max(2, arm * 0.12), Math.max(2, arm * 0.1)])
      ctx.beginPath()
      ctx.arc(x, y, rr, 0, Math.PI * 2)
      ctx.strokeStyle = neonDim
      ctx.lineWidth = lineW * 0.85
      ctx.stroke()
      ctx.setLineDash([])
    }
  }

  mapped.forEach((p, i) => {
    const arm = arms[i]
    const isNose = i === 2
    drawReticle(p.x, p.y, arm, isNose)

    const label = p.label || ''
    if (!label) return
    let tx
    let ty
    let align
    if (isNose) {
      align = 'center'
      tx = p.x
      ty = p.y + arm + gap + fontPx * 0.25
    } else {
      const toRight = p.x >= cx
      const toBelow = p.y > cy + eyeSpan * 0.08
      align = toRight ? 'left' : 'right'
      tx = p.x + (toRight ? arm + gap : -(arm + gap))
      ty = p.y + (toBelow ? fontPx * 0.55 : -fontPx * 0.55)
    }
    ctx.textAlign = align
    ctx.shadowBlur = Math.max(2, lineW * 3)
    const tag = isNose ? `◆ ${label}` : `› ${label}`
    ctx.fillStyle = neon
    ctx.fillText(tag, tx, ty)
  })
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
    if (props.showFaceLandmarks) {
      drawFiveKeypoints(ctx, view.landmarks, sx, sy)
    }
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
  () => [props.originView, props.mattingView, props.activeOriginTab, props.hasOrigin, props.hasMatting, props.showFaceLandmarks],
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
            <label
              v-if="activeOriginTab === 'original' && hasOrigin"
              class="landmark-toggle landmark-toggle--float"
            >
              <span>五官标记</span>
              <n-switch
                size="small"
                :value="showFaceLandmarks"
                @update:value="emit('update:showFaceLandmarks', $event)"
              />
            </label>
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
