<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, watch, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { IDPhotoService } from './services/idphoto'

const message = useMessage()

const originCanvas = ref(null)
const resultCanvas = ref(null)
const expandedPanels = ref(['beauty'])
const processing = ref(false)
const progressPercent = ref(0)
const statusText = ref('就绪')
const processTagType = ref('success')
const hasOrigin = ref(false)
const hasResult = ref(false)
const activeResultTab = ref('idphoto')

const resultTabs = [
  { key: 'single', label: '单张照片', hint: '精修单图' },
  { key: 'layout', label: '排版照', hint: '6寸打印排版' },
  { key: 'social', label: '社交照', hint: '方形社交尺寸' },
  { key: 'idphoto', label: '证件照', hint: '标准证件规格' },
]

const resultSet = reactive({
  single: '',
  layout: '',
  social: '',
  idphoto: '',
})

const currentResultHint = computed(() => {
  return resultTabs.find((t) => t.key === activeResultTab.value)?.hint || '成品预览'
})

const params = reactive({
  template: 'one_inch',
  bgMode: 'solid',
  bgColor: '#FFFFFF',
  bgPreset: '#FFFFFF',
  customWidth: 25,
  customHeight: 35,
  beautyStrength: 0.22,
  eyeSharp: 0.15,
  skinBright: 0.18,
  enableMakeup: false,
  lipStrength: 0.12,
  refineBrow: true,
  browFill: 0.18,
  refineHair: true,
  hairColorUniform: 0.22,
  hairlineRepair: 0.15,
  enableCloth: false,
  clothType: 'white_shirt',
  clothFit: 0.25,
  addWatermark: false,
  genPrintLayout: true,
  paperSize: '6inch',
  targetFileSize: 200,
  maskFeather: 0.3,
})

const bgModes = [
  { value: 'solid', label: '纯色' },
  { value: 'vertical', label: '上下渐变' },
  { value: 'radial', label: '中心渐变' },
]

function modePreviewStyle(mode) {
  const a = params.bgColor || '#FFFFFF'
  const b = '#FFFFFF'
  if (mode === 'vertical') return { background: `linear-gradient(180deg, ${a} 0%, ${b} 100%)` }
  if (mode === 'radial') return { background: `radial-gradient(circle at 50% 42%, ${a} 0%, ${b} 100%)` }
  return { background: a }
}

const clothOptions = [
  { label: '白衬衫', value: 'white_shirt' },
  { label: '深色西装', value: 'dark_suit' },
  { label: '学士服', value: 'grad' },
  { label: '商务职业照', value: 'business' },
]

const paperSizes = [
  { value: '5inch', title: '5寸', desc: '89×127 mm' },
  { value: '6inch', title: '6寸', desc: '102×152 mm' },
  { value: '7inch', title: '7寸', desc: '127×178 mm' },
  { value: 'a6', title: 'A6', desc: '105×148 mm' },
  { value: 'a5', title: 'A5', desc: '148×210 mm' },
  { value: 'a4', title: 'A4', desc: '210×297 mm' },
]

const currentPaperLabel = computed(() => {
  return paperSizes.find((p) => p.value === params.paperSize)?.title || '6寸'
})

const specCategories = [
  { key: 'common', label: '常用' },
  { key: 'visa', label: '签职' },
  { key: 'id', label: '证件' },
  { key: 'school', label: '升学' },
  { key: 'exam', label: '考试' },
  { key: 'custom', label: '自定义' },
]

const allSpecs = [
  { value: 'one_inch', title: '一寸', desc: '25×35 mm', categories: ['common', 'id'], keywords: '一寸 常用' },
  { value: 'two_inch', title: '二寸', desc: '35×49 mm', categories: ['common', 'id'], keywords: '二寸 常用' },
  { value: 'small_two', title: '小二寸', desc: '33×48 mm', categories: ['common'], keywords: '小二寸' },
  { value: 'large_one', title: '大一寸', desc: '33×48 mm', categories: ['common'], keywords: '大一寸' },
  { value: 'passport', title: '护照', desc: '33×48 mm', categories: ['common', 'visa', 'id'], keywords: '护照 出国' },
  { value: 'visa_us', title: '美签', desc: '51×51 mm', categories: ['visa'], keywords: '美签 美国签证' },
  { value: 'visa_schengen', title: '申根签', desc: '35×45 mm', categories: ['visa'], keywords: '申根 欧洲签证' },
  { value: 'visa_jp', title: '日签', desc: '45×45 mm', categories: ['visa'], keywords: '日签 日本签证' },
  { value: 'id_card', title: '身份证', desc: '26×32 mm', categories: ['id'], keywords: '身份证' },
  { value: 'driver', title: '驾驶证', desc: '22×32 mm', categories: ['id'], keywords: '驾驶证 驾照' },
  { value: 'social', title: '社保卡', desc: '26×32 mm', categories: ['id'], keywords: '社保卡' },
  { value: 'teacher', title: '教资', desc: '25×35 mm · 教师资格', categories: ['exam', 'visa'], keywords: '教资 教师资格' },
  { value: 'civil', title: '公务员', desc: '25×35 mm', categories: ['exam', 'visa'], keywords: '公务员 国考' },
  { value: 'grad', title: '毕业证', desc: '33×48 mm', categories: ['school'], keywords: '毕业证 学历' },
  { value: 'student', title: '学生证', desc: '25×35 mm', categories: ['school'], keywords: '学生证' },
  { value: 'exam_cet', title: '四六级', desc: '144×192 px', categories: ['exam'], keywords: '英语 四六级 CET' },
  { value: 'exam_cs', title: '计算机等级', desc: '144×192 px', categories: ['exam'], keywords: '计算机等级 NCRE' },
  { value: 'exam_nurse', title: '护士资格', desc: '25×35 mm', categories: ['exam'], keywords: '护士资格证' },
]

const bgPresets = [
  { value: '#FFFFFF', label: '白' },
  { value: '#D92121', label: '红' },
  { value: '#0047AB', label: '蓝' },
]

const specKeyword = ref('')
const activeSpecCategory = ref('common')

const filteredSpecs = computed(() => {
  if (activeSpecCategory.value === 'custom') return []
  const q = specKeyword.value.trim().toLowerCase()
  return allSpecs.filter((item) => {
    const inCategory = !q
      ? item.categories.includes(activeSpecCategory.value)
      : true
    if (!inCategory) return false
    if (!q) return true
    const hay = `${item.title} ${item.desc} ${item.keywords}`.toLowerCase()
    return hay.includes(q)
  })
})

const showCustomSpec = computed(() => activeSpecCategory.value === 'custom' && !specKeyword.value.trim())

const report = reactive({
  faceOk: false,
  faceScore: 0,
  isRephoto: false,
  isAiImage: false,
})

const showExportModal = ref(false)
const showSettingModal = ref(false)
const showAboutModal = ref(false)
const showUpgradeModal = ref(false)
const appVersion = '1.0.0'
const checkingUpgrade = ref(false)
const exportDir = ref('')
const exportOpt = reactive({
  single: true,
  layout: true,
  social: true,
  idphoto: true,
})
const setting = reactive({
  modelDir: '',
  cacheDir: '',
  enableRephotoDetect: true,
  enableAiDetect: true,
})
const aboutInfo = {
  name: '最美证件照',
  version: appVersion,
  author: '王新勇(Tacey Wong)',
  desc: '本地影楼级证件照精修与排版工具，支持多规格制证、打印纸排版与合规核验。',
}

const unbinders = []

const statusTone = computed(() => {
  if (processTagType.value === 'error') return 'bad'
  if (processTagType.value === 'warning') return 'warn'
  return 'ok'
})

const diagnostics = computed(() => [
  {
    key: 'face',
    label: '人脸检测',
    ok: report.faceOk,
    text: report.faceOk ? `质量分 ${Number(report.faceScore).toFixed(2)}` : '未检测到人脸 / 多人',
  },
  {
    key: 'rephoto',
    label: '屏幕翻拍',
    ok: !report.isRephoto,
    text: report.isRephoto ? '疑似翻拍照片' : '非翻拍照片',
  },
  {
    key: 'ai',
    label: 'AI 图片',
    ok: !report.isAiImage,
    text: report.isAiImage ? '疑似 AI 生成' : '真实照片',
  },
])

function onProgress(data) {
  const payload = data?.percent !== undefined ? data : data?.[0] || data
  progressPercent.value = payload.percent ?? 0
  statusText.value = payload.msg || statusText.value
}

function onDone(res) {
  const payload = res?.resultImg !== undefined || res?.results ? res : res?.[0] || res
  processing.value = false
  progressPercent.value = 100
  statusText.value = '生成完成'
  processTagType.value = 'success'
  if (payload?.report) Object.assign(report, payload.report)
  drawOriginCanvas(payload.originImg, payload.faceBox || [], payload.landmarks || [])

  const next = payload?.results || {}
  resultSet.single = next.single || payload.resultImg || ''
  resultSet.layout = next.layout || ''
  resultSet.social = next.social || ''
  resultSet.idphoto = next.idphoto || payload.resultImg || ''
  hasResult.value = Boolean(resultSet.single || resultSet.layout || resultSet.social || resultSet.idphoto)
  showActiveResult()
}

function showActiveResult() {
  const img = resultSet[activeResultTab.value]
  if (img) {
    drawResultCanvas(img)
    return
  }
  const tab = resultTabs.find((t) => t.key === activeResultTab.value)
  paintPlaceholder(resultCanvas.value, tab?.label || '成品', '暂无该类型成品')
}

function switchResultTab(key) {
  activeResultTab.value = key
  if (hasResult.value) showActiveResult()
}

function onError(err) {
  const payload = typeof err === 'string' ? { msg: err } : err?.[0] || err || {}
  processing.value = false
  statusText.value = payload.msg || '处理失败'
  processTagType.value = 'error'
  message.error(statusText.value)
}

function onMockProgress(e) {
  onProgress(e.detail)
}
function onMockDone(e) {
  onDone(e.detail)
}
function onMockError(e) {
  onError(e.detail)
}

onMounted(async () => {
  try {
    const { EventsOn } = await import('../wailsjs/runtime/runtime')
    unbinders.push(EventsOn('IDPhoto.OnProgress', onProgress))
    unbinders.push(EventsOn('IDPhoto.OnDone', onDone))
    unbinders.push(EventsOn('IDPhoto.OnError', onError))
  } catch {
    // browser preview
  }
  window.addEventListener('IDPhoto.OnProgress', onMockProgress)
  window.addEventListener('IDPhoto.OnDone', onMockDone)
  window.addEventListener('IDPhoto.OnError', onMockError)

  initEmptyCanvases()
})

onBeforeUnmount(() => {
  unbinders.forEach((off) => typeof off === 'function' && off())
  window.removeEventListener('IDPhoto.OnProgress', onMockProgress)
  window.removeEventListener('IDPhoto.OnDone', onMockDone)
  window.removeEventListener('IDPhoto.OnError', onMockError)
})

watch(
  () => params.bgPreset,
  (color) => {
    if (color) params.bgColor = color
  },
)

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

function initEmptyCanvases() {
  paintPlaceholder(originCanvas.value, '等待原图', '打开或拖入照片')
  paintPlaceholder(resultCanvas.value, '成品预览', '生成后在此显示')
  hasOrigin.value = false
  hasResult.value = false
  resultSet.single = ''
  resultSet.layout = ''
  resultSet.social = ''
  resultSet.idphoto = ''
}

function drawOriginCanvas(imgData, faceBox, landmarks) {
  if (!originCanvas.value || !imgData) return
  const canvas = originCanvas.value
  const ctx = canvas.getContext('2d')
  const img = new Image()
  img.onload = () => {
    canvas.width = 360
    canvas.height = Math.round((360 * img.height) / img.width) || 480
    ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
    const sx = canvas.width / img.width
    const sy = canvas.height / img.height
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
    if (landmarks?.length) {
      ctx.fillStyle = '#c45b7a'
      for (let i = 0; i < landmarks.length; i += 2) {
        ctx.beginPath()
        ctx.arc(landmarks[i] * sx, landmarks[i + 1] * sy, 3, 0, Math.PI * 2)
        ctx.fill()
      }
    }
    hasOrigin.value = true
  }
  img.src = imgData
}

function drawResultCanvas(imgData) {
  if (!resultCanvas.value || !imgData) return
  const canvas = resultCanvas.value
  const ctx = canvas.getContext('2d')
  const img = new Image()
  img.onload = () => {
    const maxW = activeResultTab.value === 'layout' ? 420 : 360
    const maxH = 480
    const scale = Math.min(maxW / img.width, maxH / img.height, 1)
    canvas.width = Math.max(1, Math.round(img.width * scale))
    canvas.height = Math.max(1, Math.round(img.height * scale))
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
    hasResult.value = true
  }
  img.src = imgData
}

async function openImage() {
  const res = await IDPhotoService.OpenImageDialog()
  if (!res) return
  await loadImage(res)
}

async function loadImage(path) {
  const ret = await IDPhotoService.LoadImage(path)
  Object.assign(report, ret.report)
  drawOriginCanvas(ret.imgBase64, ret.faceBox, ret.landmarks)
  statusText.value = '已加载图片'
  processTagType.value = 'success'
}

function handleDrop(e) {
  const file = e.dataTransfer?.files?.[0]
  if (!file) return
  loadImage(file.path || file.name)
}

function resetAll() {
  params.template = 'one_inch'
  params.bgMode = 'solid'
  params.bgColor = '#FFFFFF'
  params.bgPreset = '#FFFFFF'
  params.customWidth = 25
  params.customHeight = 35
  params.beautyStrength = 0.22
  activeSpecCategory.value = 'common'
  specKeyword.value = ''
  params.eyeSharp = 0.15
  params.skinBright = 0.18
  params.enableMakeup = false
  params.lipStrength = 0.12
  params.refineBrow = true
  params.browFill = 0.18
  params.refineHair = true
  params.hairColorUniform = 0.22
  params.hairlineRepair = 0.15
  params.enableCloth = false
  params.clothType = 'white_shirt'
  params.clothFit = 0.25
  params.addWatermark = false
  params.genPrintLayout = true
  params.paperSize = '6inch'
  params.targetFileSize = 200
  params.maskFeather = 0.3
  progressPercent.value = 0
  statusText.value = '就绪'
  processing.value = false
  processTagType.value = 'success'
  Object.assign(report, { faceOk: false, faceScore: 0, isRephoto: false, isAiImage: false })
  initEmptyCanvases()
  message.info('已重置')
}

async function runGenerate() {
  processing.value = true
  progressPercent.value = 0
  statusText.value = '开始处理...'
  processTagType.value = 'warning'
  try {
    await IDPhotoService.Generate({ ...params })
  } catch (err) {
    onError({ msg: err?.message || '生成失败' })
  }
}

function onTemplateChange(value) {
  params.template = value
}

function applyCustomSpec() {
  params.template = 'custom'
  message.success(`已应用自定义尺寸 ${params.customWidth}×${params.customHeight} mm`)
}

function selectSpecCategory(key) {
  activeSpecCategory.value = key
  specKeyword.value = ''
  if (key === 'custom') {
    params.template = 'custom'
  }
}

function openExportModal() {
  showExportModal.value = true
}

async function selectExportDir() {
  const dir = await IDPhotoService.SelectFolder()
  if (dir) exportDir.value = dir
}

async function doExport() {
  if (!exportDir.value) {
    message.warning('请先选择输出目录')
    return
  }
  await IDPhotoService.Export(exportDir.value, { ...exportOpt })
  showExportModal.value = false
  message.success('导出完成（原型模拟）')
}

function openSettingModal() {
  showSettingModal.value = true
}

function openAboutModal() {
  showAboutModal.value = true
}

function openUpgradeModal() {
  showUpgradeModal.value = true
}

async function checkAndUpgrade() {
  checkingUpgrade.value = true
  await new Promise((r) => setTimeout(r, 800))
  checkingUpgrade.value = false
  showUpgradeModal.value = false
  message.info('当前已是最新版本 v' + appVersion)
}

function saveSetting() {
  showSettingModal.value = false
  message.success('设置已保存（原型模拟）')
}

function formatVal(v) {
  return Number(v).toFixed(2)
}
</script>

<template>
  <div class="studio">
    <header class="topbar">
      <div class="brand-block">
        <div class="brand-mark" aria-hidden="true" />
        <div class="titles">
          <h1 class="brand">最美证件照</h1>
          <p class="brand-sub">影楼精修 · 合规规格 · 一键出片</p>
        </div>
      </div>

      <div class="actions">
        <button class="btn ghost" type="button" @click="openImage">打开图片</button>
        <button class="btn ghost" type="button" @click="resetAll">重置</button>
        <button class="btn primary" type="button" :disabled="processing" @click="runGenerate">
          <span v-if="processing" class="spin" />
          {{ processing ? '精修中…' : '开始生成' }}
        </button>
        <button class="btn ghost" type="button" @click="openExportModal">导出</button>
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

    <aside class="rail left">
      <section class="rail-section pin-top">
        <div class="section-label">背景颜色</div>
        <div class="swatches">
          <button
            v-for="c in bgPresets"
            :key="c.value"
            type="button"
            class="swatch"
            :class="{ active: params.bgPreset === c.value }"
            :style="{ '--swatch': c.value }"
            @click="params.bgPreset = c.value"
          >
            {{ c.label }}
          </button>
        </div>
        <n-color-picker v-model:value="params.bgColor" :modes="['hex']" size="small" style="margin-top: 10px" />
        <div class="bg-modes" style="margin-top: 10px">
          <button
            v-for="m in bgModes"
            :key="m.value"
            type="button"
            class="bg-mode"
            :class="{ active: params.bgMode === m.value }"
            @click="params.bgMode = m.value"
          >
            <span class="bg-mode-preview" :style="modePreviewStyle(m.value)" />
            {{ m.label }}
          </button>
        </div>
      </section>

      <section class="rail-section grow specs-section">
        <div class="section-label">证件规格</div>
        <n-input
          v-model:value="specKeyword"
          size="small"
          clearable
          placeholder="搜索规格，如护照、教资"
          class="spec-search"
        />
        <div class="spec-cats" role="tablist">
          <button
            v-for="cat in specCategories"
            :key="cat.key"
            type="button"
            class="spec-cat"
            :class="{ active: activeSpecCategory === cat.key && !specKeyword.trim() }"
            @click="selectSpecCategory(cat.key)"
          >
            {{ cat.label }}
          </button>
        </div>

        <div class="spec-scroll">
          <div v-if="showCustomSpec" class="custom-spec">
            <div class="field">
              <label>宽度 mm</label>
              <n-input-number v-model:value="params.customWidth" :min="10" :max="100" size="small" style="width: 100%" />
            </div>
            <div class="field">
              <label>高度 mm</label>
              <n-input-number v-model:value="params.customHeight" :min="10" :max="140" size="small" style="width: 100%" />
            </div>
            <button class="btn primary custom-apply" type="button" @click="applyCustomSpec">应用自定义</button>
          </div>

          <div v-else class="template-grid">
            <button
              v-for="t in filteredSpecs"
              :key="t.value"
              type="button"
              class="template-card"
              :class="{ active: params.template === t.value }"
              @click="onTemplateChange(t.value)"
            >
              <strong>{{ t.title }}</strong>
              <span>{{ t.desc }}</span>
            </button>
            <p v-if="!filteredSpecs.length" class="spec-empty">未找到匹配规格</p>
          </div>
        </div>
      </section>

      <section class="rail-section pin-footer">
        <button class="footer-link" type="button" @click="openAboutModal">关于</button>
        <button class="footer-link" type="button" @click="openSettingModal">设置</button>
        <button class="footer-version" type="button" @click="openUpgradeModal" title="检查更新">
          v{{ appVersion }}
        </button>
      </section>
    </aside>

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
              <em>人脸检测</em>
            </header>
            <div
              class="frame-body"
              @dragover.prevent
              @drop.prevent="handleDrop"
            >
              <canvas ref="originCanvas" />
              <p v-if="!hasOrigin" class="frame-hint">拖放照片到这里，或点「打开图片」</p>
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
                @click="switchResultTab(tab.key)"
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
        <p v-if="params.beautyStrength >= 0.35" class="report-warn">
          美颜强度偏高，可能影响人脸核验通过率
        </p>
      </section>
    </main>

    <aside class="rail right">
      <div class="section-label">精修参数</div>
      <n-collapse v-model:expanded-names="expandedPanels" accordion display-directive="show">
        <n-collapse-item title="基础美颜" name="beauty">
          <div class="field">
            <div class="field-row">
              <label>美颜强度</label>
              <span>{{ formatVal(params.beautyStrength) }}</span>
            </div>
            <n-slider v-model:value="params.beautyStrength" :max="0.4" :step="0.01" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>眼睛锐化</label>
              <span>{{ formatVal(params.eyeSharp) }}</span>
            </div>
            <n-slider v-model:value="params.eyeSharp" :max="0.3" :step="0.01" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>肤色提亮</label>
              <span>{{ formatVal(params.skinBright) }}</span>
            </div>
            <n-slider v-model:value="params.skinBright" :max="0.3" :step="0.01" />
          </div>
        </n-collapse-item>

        <n-collapse-item title="自动化妆" name="makeup">
          <div class="field switch-row">
            <label>自然妆感</label>
            <n-switch v-model:value="params.enableMakeup" size="small" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>唇色增强</label>
              <span>{{ formatVal(params.lipStrength) }}</span>
            </div>
            <n-slider
              v-model:value="params.lipStrength"
              :max="0.25"
              :step="0.01"
              :disabled="!params.enableMakeup"
            />
          </div>
        </n-collapse-item>

        <n-collapse-item title="眉毛与头发" name="hair">
          <div class="field switch-row">
            <label>眉毛自然增强</label>
            <n-switch v-model:value="params.refineBrow" size="small" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>眉尾补全</label>
              <span>{{ formatVal(params.browFill) }}</span>
            </div>
            <n-slider
              v-model:value="params.browFill"
              :max="0.3"
              :step="0.01"
              :disabled="!params.refineBrow"
            />
          </div>
          <div class="field switch-row">
            <label>碎发整理</label>
            <n-switch v-model:value="params.refineHair" size="small" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>发色统一</label>
              <span>{{ formatVal(params.hairColorUniform) }}</span>
            </div>
            <n-slider
              v-model:value="params.hairColorUniform"
              :max="0.35"
              :step="0.01"
              :disabled="!params.refineHair"
            />
          </div>
          <div class="field">
            <div class="field-row">
              <label>发际线修补</label>
              <span>{{ formatVal(params.hairlineRepair) }}</span>
            </div>
            <n-slider
              v-model:value="params.hairlineRepair"
              :max="0.3"
              :step="0.01"
              :disabled="!params.refineHair"
            />
          </div>
        </n-collapse-item>

        <n-collapse-item title="智能正装" name="cloth">
          <div class="field switch-row">
            <label>智能换正装</label>
            <n-switch v-model:value="params.enableCloth" size="small" />
          </div>
          <div class="field">
            <n-select
              v-model:value="params.clothType"
              :options="clothOptions"
              :disabled="!params.enableCloth"
              size="small"
            />
          </div>
          <div class="field">
            <div class="field-row">
              <label>服装贴合度</label>
              <span>{{ formatVal(params.clothFit) }}</span>
            </div>
            <n-slider
              v-model:value="params.clothFit"
              :max="0.4"
              :step="0.01"
              :disabled="!params.enableCloth"
            />
          </div>
        </n-collapse-item>

        <n-collapse-item title="高级选项" name="adv">
          <div class="field switch-row">
            <label>溯源盲水印</label>
            <n-switch v-model:value="params.addWatermark" size="small" />
          </div>
          <div class="field switch-row">
            <label>6 寸打印排版</label>
                <n-switch v-model:value="params.genPrintLayout" size="small" />
              </div>
              <div v-if="params.genPrintLayout" class="field">
                <label>当前纸张：{{ currentPaperLabel }}</label>
                <n-select
                  v-model:value="params.paperSize"
                  :options="paperSizes.map((p) => ({ label: `${p.title}（${p.desc}）`, value: p.value }))"
                  size="small"
                />
              </div>
              <div class="field">
                <label>目标文件大小 KB</label>
            <n-input-number
              v-model:value="params.targetFileSize"
              :min="50"
              :max="2000"
              size="small"
              style="width: 100%"
            />
          </div>
          <div class="field">
            <div class="field-row">
              <label>边缘羽化</label>
              <span>{{ formatVal(params.maskFeather) }}</span>
            </div>
            <n-slider v-model:value="params.maskFeather" :max="0.6" :step="0.01" />
          </div>
        </n-collapse-item>
      </n-collapse>
    </aside>

    <n-modal
      v-model:show="showExportModal"
      preset="card"
      title="导出证件照"
      style="width: 480px"
      :bordered="false"
      :segmented="{ content: true, footer: 'soft' }"
      :mask-closable="false"
    >
      <div class="modal-stack">
        <n-checkbox v-model:checked="exportOpt.single">单张照片</n-checkbox>
        <n-checkbox v-model:checked="exportOpt.idphoto">证件照</n-checkbox>
        <n-checkbox v-model:checked="exportOpt.layout">排版照</n-checkbox>
        <n-checkbox v-model:checked="exportOpt.social">社交照</n-checkbox>
        <label class="modal-label">输出目录</label>
        <n-input v-model:value="exportDir" readonly placeholder="请选择导出目录" />
      </div>
      <template #footer>
        <div class="modal-actions">
          <button class="btn ghost" type="button" @click="selectExportDir">选择目录</button>
          <button class="btn primary" type="button" @click="doExport">开始导出</button>
        </div>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showSettingModal"
      preset="card"
      title="软件设置"
      style="width: 460px"
      :bordered="false"
      :segmented="{ content: true, footer: 'soft' }"
      :mask-closable="false"
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
          <button class="btn primary" type="button" @click="saveSetting">保存设置</button>
        </div>
      </template>
    </n-modal>
    <n-modal
      v-model:show="showAboutModal"
      preset="card"
      title="关于软件"
      style="width: 440px"
      :bordered="false"
      :segmented="{ content: true, footer: 'soft' }"
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
          <button class="btn primary" type="button" @click="showAboutModal = false">知道了</button>
        </div>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showUpgradeModal"
      preset="card"
      title="版本更新"
      style="width: 420px"
      :bordered="false"
      :segmented="{ content: true, footer: 'soft' }"
      :mask-closable="false"
    >
      <div class="modal-stack">
        <p class="upgrade-text">
          当前版本为 <strong>v{{ appVersion }}</strong>。是否检查并升级到最新版本？
        </p>
      </div>
      <template #footer>
        <div class="modal-actions">
          <button class="btn ghost" type="button" @click="showUpgradeModal = false">暂不升级</button>
          <button class="btn primary" type="button" :disabled="checkingUpgrade" @click="checkAndUpgrade">
            <span v-if="checkingUpgrade" class="spin" />
            {{ checkingUpgrade ? '检查中…' : '检查升级' }}
          </button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.studio {
  width: 100%;
  height: 100%;
  min-height: 0;
  flex: 1 1 auto;
  display: grid;
  grid-template-columns: 228px minmax(0, 1fr) 288px;
  grid-template-rows: 60px minmax(0, 1fr);
  gap: 0;
  padding: 0;
  overflow: hidden;
  animation: studio-in 360ms ease both;
}

@keyframes studio-in {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: none; }
}

.topbar {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: minmax(220px, 1.1fr) auto minmax(200px, 1fr);
  align-items: center;
  gap: 14px;
  padding: 0 16px;
  border: none;
  border-bottom: 1px solid var(--line);
  border-radius: 0;
  background: rgba(255, 252, 253, 0.94);
  backdrop-filter: blur(16px);
  box-shadow: none;
  min-width: 0;
  overflow: hidden;
  z-index: 2;
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.brand-mark {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  flex: none;
  background:
    radial-gradient(circle at 30% 28%, #f3c4d2 0%, transparent 42%),
    linear-gradient(145deg, #d97a95 0%, #c45b7a 48%, #a34460 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.35);
}

.titles {
  min-width: 0;
}

.brand {
  margin: 0;
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.1;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.brand-sub {
  margin: 3px 0 0;
  font-size: 12px;
  color: var(--ink-faint);
  letter-spacing: 0.04em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn {
  height: 34px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid transparent;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
  transition: background 160ms ease, border-color 160ms ease, transform 160ms ease, opacity 160ms ease;
}

.btn:active { transform: translateY(1px); }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }

.btn.primary {
  background: var(--rose);
  color: #fff;
  box-shadow: 0 8px 18px rgba(196, 91, 122, 0.28);
}

.btn.primary:hover:not(:disabled) {
  background: var(--rose-deep);
}

.btn.ghost {
  background: rgba(255, 255, 255, 0.75);
  border-color: var(--line-strong);
  color: var(--ink);
}

.btn.ghost:hover {
  background: #fff;
  border-color: rgba(196, 91, 122, 0.35);
}

.btn.quiet {
  background: transparent;
  color: var(--ink-soft);
}

.btn.quiet:hover {
  color: var(--ink);
  background: rgba(70, 40, 60, 0.04);
}

.spin {
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.status-block {
  justify-self: end;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.status-pill {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 6px 12px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  background: var(--rose-soft);
  color: var(--rose-deep);
}

.status-pill[data-tone='warn'] {
  background: #f7edd4;
  color: #8a6a12;
}

.status-pill[data-tone='bad'] {
  background: #f8e2e2;
  color: #a34444;
}

.progress-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 150px;
}

.progress-track {
  flex: 1;
  height: 6px;
  border-radius: 6px;
  background: rgba(70, 40, 60, 0.08);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  width: 0;
  border-radius: inherit;
  background: linear-gradient(90deg, #c45b7a, #e08aa4);
  transition: width 220ms ease;
}

.progress-num {
  width: 36px;
  text-align: right;
  font-size: 12px;
  color: var(--ink-faint);
  font-variant-numeric: tabular-nums;
}

.rail {
  min-height: 0;
  min-width: 0;
  border: none;
  border-radius: 0;
  background: rgba(255, 252, 253, 0.88);
  backdrop-filter: blur(16px);
  box-shadow: none;
  padding: 14px;
  overflow: hidden;
  animation: panel-in 420ms ease both;
}

.rail.left {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 0;
  overflow: hidden;
  animation: none;
  border-right: 1px solid var(--line);
}

.rail.right {
  overflow-x: hidden;
  overflow-y: auto;
  animation-delay: 100ms;
  border-left: 1px solid var(--line);
  scrollbar-width: thin;
  scrollbar-color: rgba(196, 91, 122, 0.28) transparent;
}

.rail.right::-webkit-scrollbar {
  width: 6px;
}

.rail.right::-webkit-scrollbar-thumb {
  background: rgba(196, 91, 122, 0.28);
  border-radius: 6px;
}

.rail.right::-webkit-scrollbar-track {
  background: transparent;
}

@keyframes panel-in {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: none; }
}

.rail.left .rail-section {
  padding: 12px 14px;
}

.rail.left .pin-top {
  flex: 0 0 auto;
  border-bottom: 1px solid var(--line);
}

.rail.left .pin-footer {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 10px 12px;
  border-top: 1px solid var(--line);
  background: rgba(255, 252, 253, 0.96);
}

.rail.left .specs-section {
  flex: 1 1 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.rail.left .section-label,
.rail.left .spec-search,
.rail.left .spec-cats {
  flex: 0 0 auto;
}

.rail.left .spec-scroll {
  flex: 1 1 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  margin-top: 4px;
  padding-right: 4px;
  scrollbar-width: thin;
  scrollbar-color: rgba(196, 91, 122, 0.28) transparent;
}

.rail.left .spec-scroll::-webkit-scrollbar {
  width: 6px;
}

.rail.left .spec-scroll::-webkit-scrollbar-thumb {
  background: rgba(196, 91, 122, 0.28);
  border-radius: 6px;
}

.rail.left .spec-scroll::-webkit-scrollbar-track {
  background: transparent;
}

.rail-section.grow {
  flex: 1;
  min-height: 0;
}

.section-label,
.report-title {
  margin-bottom: 10px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--ink-faint);
}

.spec-search {
  margin-bottom: 10px;
}

.spec-cats {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}

.spec-cat {
  height: 28px;
  padding: 0 10px;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.75);
  color: var(--ink-soft);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: background 160ms ease, border-color 160ms ease, color 160ms ease;
}

.spec-cat:hover {
  border-color: rgba(196, 91, 122, 0.35);
  color: var(--rose-deep);
}

.spec-cat.active {
  background: var(--rose-soft);
  border-color: rgba(196, 91, 122, 0.4);
  color: var(--rose-deep);
}

.custom-spec {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.custom-apply {
  margin-top: 4px;
  width: 100%;
  justify-content: center;
}

.spec-empty {
  margin: 12px 0 0;
  font-size: 12px;
  color: var(--ink-faint);
  text-align: center;
}

.template-grid {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.template-card {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
  padding: 11px 12px;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.7);
  text-align: left;
  cursor: pointer;
  transition: background 160ms ease, border-color 160ms ease, transform 160ms ease;
}

.template-card:hover {
  transform: translateY(-1px);
  background: var(--rose-soft);
  border-color: rgba(196, 91, 122, 0.18);
}

.template-card.active {
  border-color: rgba(196, 91, 122, 0.4);
  background: var(--rose-soft);
}

.template-card strong {
  font-size: 14px;
  color: var(--ink);
}

.template-card span {
  font-size: 12px;
  color: var(--ink-faint);
}

.stage {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
  animation: panel-in 420ms ease both;
  animation-delay: 70ms;
}

.preview-board {
  position: relative;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 14px 16px 16px;
  overflow: auto;
  border-radius: 0;
  border: none;
  background:
    radial-gradient(ellipse at 50% 0%, rgba(196, 91, 122, 0.08), transparent 55%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.4), rgba(255, 255, 255, 0.18)),
    var(--stage);
  box-shadow: none;
}

.board-caption {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 12px;
}

.board-caption span {
  font-family: var(--font-display);
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: var(--ink);
}

.board-caption em {
  font-style: normal;
  font-size: 12px;
  color: var(--ink-faint);
}

.preview-row {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: stretch;
  justify-content: center;
  gap: 18px;
}

.frame {
  width: min(360px, 42%);
  min-width: 220px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.frame.result {
  width: min(440px, 48%);
}

.frame-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 0 2px;
}

.frame-head span {
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 700;
  color: var(--ink);
}

.frame-head em {
  font-style: normal;
  font-size: 12px;
  color: var(--ink-faint);
}

.frame-body {
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.9);
  min-height: 0;
}

.frame.result .frame-body {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(255, 246, 249, 0.9));
}

.result-tabs {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 4px;
  padding: 2px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid var(--line);
}

.result-tab {
  height: 30px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--ink-soft);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: background 160ms ease, color 160ms ease;
}

.result-tab:hover {
  color: var(--rose-deep);
  background: rgba(196, 91, 122, 0.08);
}

.result-tab.active {
  color: #fff;
  background: var(--rose);
}

.frame-body canvas {
  max-width: 100%;
  max-height: 100%;
  height: auto;
  border-radius: 10px;
  display: block;
  background: #fff;
}

.frame-hint {
  position: absolute;
  left: 14px;
  right: 14px;
  bottom: 16px;
  margin: 0;
  text-align: center;
  font-size: 12px;
  color: var(--ink-faint);
  pointer-events: none;
}

.bridge {
  align-self: center;
  width: 48px;
  display: flex;
  justify-content: center;
}

.bridge span {
  width: 48px;
  height: 2px;
  border-radius: 2px;
  background: linear-gradient(90deg, transparent, rgba(196, 91, 122, 0.55), transparent);
  position: relative;
}

.bridge span::before,
.bridge span::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--rose);
  transform: translateY(-50%);
}

.bridge span::before { left: 0; }
.bridge span::after { right: 0; }

.report {
  flex: 0 0 auto;
  border: none;
  border-top: 1px solid var(--line);
  border-radius: 0;
  background: rgba(255, 252, 253, 0.92);
  backdrop-filter: blur(16px);
  padding: 12px 14px;
  box-shadow: none;
}

.report-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.report-item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid var(--line);
}

.report-item .dot {
  width: 8px;
  height: 8px;
  margin-top: 5px;
  border-radius: 50%;
  background: var(--bad);
  flex: none;
}

.report-item[data-ok='true'] .dot {
  background: var(--ok);
}

.report-item strong {
  display: block;
  font-size: 13px;
  color: var(--ink);
}

.report-item p {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--ink-soft);
}

.report-warn {
  margin: 10px 0 0;
  font-size: 12px;
  color: var(--warn);
}

.field { margin-bottom: 14px; }

.field > label,
.field-row label,
.modal-label {
  display: block;
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--ink-soft);
}

.field-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.field-row span {
  font-size: 12px;
  color: var(--ink-faint);
  font-variant-numeric: tabular-nums;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.switch-row label { margin: 0; }

.swatches {
  display: flex;
  gap: 8px;
}

.bg-modes {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  margin-bottom: 0;
}

.bg-mode {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 4px;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.75);
  color: var(--ink-soft);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: border-color 160ms ease, background 160ms ease, color 160ms ease;
}

.bg-mode:hover {
  border-color: rgba(196, 91, 122, 0.3);
  color: var(--rose-deep);
}

.bg-mode.active {
  border-color: rgba(196, 91, 122, 0.45);
  background: var(--rose-soft);
  color: var(--rose-deep);
}

.bg-mode-preview {
  width: 100%;
  height: 28px;
  border-radius: 6px;
  border: 1px solid rgba(42, 36, 48, 0.08);
}

.swatch {
  flex: 1;
  height: 34px;
  border-radius: 12px;
  border: 1px solid var(--line-strong);
  background: #fff;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  color: var(--ink-soft);
  position: relative;
  padding-left: 28px;
}

.swatch::before {
  content: "";
  position: absolute;
  left: 10px;
  top: 50%;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  transform: translateY(-50%);
  background: var(--swatch);
  border: 1px solid rgba(42, 36, 48, 0.15);
}

.swatch.active {
  border-color: rgba(196, 91, 122, 0.4);
  background: var(--rose-soft);
  color: var(--rose-deep);
}

.modal-stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.footer-link,
.footer-version {
  height: 28px;
  padding: 0 8px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--ink-soft);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: background 160ms ease, color 160ms ease;
}

.footer-link:hover,
.footer-version:hover {
  background: var(--rose-soft);
  color: var(--rose-deep);
}

.footer-version {
  margin-left: auto;
  font-variant-numeric: tabular-nums;
  color: var(--ink-faint);
}

.about-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.about-brand {
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--ink);
}

.about-desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--ink-soft);
}

.about-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border-radius: 12px;
  background: var(--rose-soft);
  border: 1px solid rgba(196, 91, 122, 0.12);
}

.about-meta div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}

.about-meta span {
  color: var(--ink-faint);
}

.about-meta em {
  font-style: normal;
  color: var(--ink);
  font-weight: 600;
}

.upgrade-text {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--ink-soft);
}

.spin.dark {
  border-color: rgba(196, 91, 122, 0.25);
  border-top-color: var(--rose);
}

:deep(.n-collapse .n-collapse-item) {
  margin: 0 0 6px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.65);
  border: 1px solid var(--line);
  overflow: hidden;
}

:deep(.n-collapse .n-collapse-item .n-collapse-item__header) {
  padding: 12px !important;
  font-size: 13px;
}

:deep(.n-collapse .n-collapse-item .n-collapse-item__content-inner) {
  padding: 0 12px 12px !important;
}

:deep(.n-collapse .n-collapse-item:not(:first-child)) {
  border-top: 1px solid var(--line) !important;
}

@media (max-width: 1180px) {
  .studio {
    grid-template-columns: 200px minmax(0, 1fr) 250px;
  }
  .brand-sub { display: none; }
  .report-grid { grid-template-columns: 1fr; }
}
</style>
