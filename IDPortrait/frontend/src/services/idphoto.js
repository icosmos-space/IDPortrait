/**
 * 证件照服务层：
 * 1) Wails 桌面：优先调用 Go 绑定
 * 2) 远程浏览器：同源 /api（Echo）
 * 3) 纯前端开发：本地 mock
 */

const TEMPLATE_PRESETS = {
  one_inch: { bgColor: '#FFFFFF', label: '一寸', w: 25, h: 35 },
  two_inch: { bgColor: '#FFFFFF', label: '二寸', w: 35, h: 49 },
  small_two: { bgColor: '#FFFFFF', label: '小二寸', w: 33, h: 48 },
  large_one: { bgColor: '#FFFFFF', label: '大一寸', w: 33, h: 48 },
  passport: { bgColor: '#FFFFFF', label: '护照', w: 33, h: 48 },
  visa_us: { bgColor: '#FFFFFF', label: '美签', w: 51, h: 51 },
  visa_schengen: { bgColor: '#FFFFFF', label: '申根签', w: 35, h: 45 },
  visa_jp: { bgColor: '#FFFFFF', label: '日签', w: 45, h: 45 },
  id_card: { bgColor: '#FFFFFF', label: '身份证', w: 26, h: 32 },
  driver: { bgColor: '#FFFFFF', label: '驾驶证', w: 22, h: 32 },
  social: { bgColor: '#FFFFFF', label: '社保卡', w: 26, h: 32 },
  teacher: { bgColor: '#FFFFFF', label: '教资', w: 25, h: 35 },
  civil: { bgColor: '#FFFFFF', label: '公务员', w: 25, h: 35 },
  grad: { bgColor: '#FFFFFF', label: '毕业证', w: 33, h: 48 },
  student: { bgColor: '#FFFFFF', label: '学生证', w: 25, h: 35 },
  exam_cet: { bgColor: '#FFFFFF', label: '四六级', w: 144, h: 192, unit: 'px' },
  exam_cs: { bgColor: '#FFFFFF', label: '计算机等级', w: 144, h: 192, unit: 'px' },
  exam_nurse: { bgColor: '#FFFFFF', label: '护士资格', w: 25, h: 35 },
  custom: { bgColor: '#FFFFFF', label: '自定义尺寸' },
}

let httpMode = null // null | true | false

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function isWailsRuntime() {
  return typeof window !== 'undefined' && !!(window.go && window.go.main && window.go.main.App)
}

async function detectHttpMode() {
  if (httpMode !== null) return httpMode
  if (isWailsRuntime()) {
    httpMode = false
    return false
  }
  try {
    const res = await fetch('/api/health', { method: 'GET' })
    httpMode = res.ok
  } catch {
    httpMode = false
  }
  return httpMode
}

async function apiJSON(path, body, method = 'POST') {
  const res = await fetch(`/api${path}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body && method !== 'GET' ? JSON.stringify(body) : undefined,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || `HTTP ${res.status}`)
  }
  return data
}

function preferValue(current, fallback) {
  return current || fallback || ''
}

function normalizePhotoCatalog(raw) {
  if (!raw || typeof raw !== 'object') return null
  const list = raw.list || raw.List || []
  const categories = raw.categories || raw.Categories || []
  return {
    list: list.map((item) => ({
      value: item.value || item.Value,
      title: item.title || item.Title,
      desc: item.desc || item.Desc,
      categories: item.categories || item.Categories || [],
      keywords: item.keywords || item.Keywords || '',
      widthMm: item.widthMm ?? item.WidthMM,
      heightMm: item.heightMm ?? item.HeightMM,
      widthPx: item.widthPx ?? item.WidthPx,
      heightPx: item.heightPx ?? item.HeightPx,
      unit: item.unit || item.Unit || 'mm',
    })),
    default: raw.default || raw.Default || '',
    current: raw.current || raw.Current || '',
    categories: categories.map((c) => ({
      key: c.key || c.Key,
      label: c.label || c.Label,
    })),
  }
}

function normalizePaperCatalog(raw) {
  if (!raw || typeof raw !== 'object') return null
  const list = raw.list || raw.List || []
  return {
    list: list.map((item) => ({
      value: item.value || item.Value,
      title: item.title || item.Title,
      desc: item.desc || item.Desc,
      widthMm: item.widthMm ?? item.WidthMM,
      heightMm: item.heightMm ?? item.HeightMM,
    })),
    default: raw.default || raw.Default || '',
    current: raw.current || raw.Current || '',
  }
}

const FALLBACK_PHOTO_SPECS = {
  list: [
    { value: 'one_inch', title: '一寸', desc: '25×35 mm', categories: ['common', 'id'], keywords: '一寸 常用' },
    { value: 'two_inch', title: '二寸', desc: '35×49 mm', categories: ['common', 'id'], keywords: '二寸 常用' },
    { value: 'passport', title: '护照', desc: '33×48 mm', categories: ['common', 'visa', 'id'], keywords: '护照 出国' },
  ],
  default: 'one_inch',
  current: '',
  categories: [
    { key: 'common', label: '常用' },
    { key: 'visa', label: '签职' },
    { key: 'id', label: '证件' },
    { key: 'school', label: '升学' },
    { key: 'exam', label: '考试' },
    { key: 'custom', label: '自定义' },
  ],
}

function normalizeModelCatalog(raw) {
  if (!raw || typeof raw !== 'object') return null
  const list = raw.list || raw.List || []
  return {
    list: list.map((item) => ({
      value: item.value || item.Value,
      title: item.title || item.Title,
      desc: item.desc || item.Desc,
      keywords: item.keywords || item.Keywords || '',
    })),
    default: raw.default || raw.Default || '',
    current: raw.current || raw.Current || '',
  }
}

const FALLBACK_FACE_DETECT_MODELS = {
  list: [
    { value: 'retinaface', title: 'RetinaFace', desc: '高精度人脸框与五点' },
    { value: 'scrfd', title: 'SCRFD', desc: '轻量快速，适合实时' },
    { value: 'mediapipe', title: 'MediaPipe', desc: '移动端友好' },
  ],
  default: 'retinaface',
  current: '',
}

const FALLBACK_MATTING_MODELS = {
  list: [
    { value: 'modnet', title: 'MODNet', desc: '人像抠图，边缘自然' },
    { value: 'u2net', title: 'U²-Net', desc: '通用显著物体分割' },
    { value: 'birefnet', title: 'BiRefNet', desc: '高细节抠图' },
  ],
  default: 'modnet',
  current: '',
}

const FALLBACK_WATERMARK = {
  enabled: false,
  text: '最美证件照',
  color: '#FFFFFF',
  fontSize: 18,
  opacity: 0.28,
  angle: -30,
  spacing: 120,
}

function normalizeWatermarkSettings(raw) {
  if (!raw || typeof raw !== 'object') return null
  return {
    enabled: !!(raw.enabled ?? raw.Enabled),
    text: raw.text || raw.Text || FALLBACK_WATERMARK.text,
    color: raw.color || raw.Color || FALLBACK_WATERMARK.color,
    fontSize: Number(raw.fontSize ?? raw.FontSize ?? FALLBACK_WATERMARK.fontSize),
    opacity: Number(raw.opacity ?? raw.Opacity ?? FALLBACK_WATERMARK.opacity),
    angle: Number(raw.angle ?? raw.Angle ?? FALLBACK_WATERMARK.angle),
    spacing: Number(raw.spacing ?? raw.Spacing ?? FALLBACK_WATERMARK.spacing),
  }
}

function normalizeWatermarkConfig(raw) {
  if (!raw || typeof raw !== 'object') {
    return { default: FALLBACK_WATERMARK, current: null }
  }
  return {
    default: normalizeWatermarkSettings(raw.default || raw.Default) || FALLBACK_WATERMARK,
    current: normalizeWatermarkSettings(raw.current || raw.Current),
  }
}

function makePlaceholderThumb(label, bg = '#e8eef5') {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="80">
    <rect width="64" height="80" fill="${bg}"/>
    <text x="32" y="44" text-anchor="middle" font-size="10" fill="#64748b">${label}</text>
  </svg>`
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
}

function fillBackground(ctx, w, h, bg, mode = 'solid') {
  const end = '#FFFFFF'
  if (mode === 'vertical') {
    const g = ctx.createLinearGradient(0, 0, 0, h)
    g.addColorStop(0, bg)
    g.addColorStop(1, end)
    ctx.fillStyle = g
  } else if (mode === 'radial') {
    const g = ctx.createRadialGradient(w / 2, h * 0.42, 8, w / 2, h * 0.42, Math.max(w, h) * 0.75)
    g.addColorStop(0, bg)
    g.addColorStop(1, end)
    ctx.fillStyle = g
  } else {
    ctx.fillStyle = bg
  }
  ctx.fillRect(0, 0, w, h)
}

function makeCanvasDataUrl(w, h, bg, text, mode = 'solid') {
  const canvas = document.createElement('canvas')
  canvas.width = w
  canvas.height = h
  const ctx = canvas.getContext('2d')
  fillBackground(ctx, w, h, bg, mode)
  ctx.fillStyle = '#334155'
  ctx.beginPath()
  ctx.ellipse(w / 2, h * 0.42, w * 0.22, h * 0.18, 0, 0, Math.PI * 2)
  ctx.fill()
  ctx.fillStyle = '#475569'
  ctx.beginPath()
  ctx.ellipse(w / 2, h * 0.85, w * 0.32, h * 0.28, 0, Math.PI, 0)
  ctx.fill()
  ctx.fillStyle = '#0f172a'
  ctx.font = '14px sans-serif'
  ctx.textAlign = 'center'
  ctx.fillText(text, w / 2, h - 12)
  return canvas.toDataURL('image/png')
}

function loadImageElement(src) {
  return new Promise((resolve, reject) => {
    if (!src) {
      reject(new Error('empty image'))
      return
    }
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('image load failed'))
    img.src = src
  })
}

async function makeMattingDataUrl(source) {
  try {
    const img = await loadImageElement(source)
    const w = img.naturalWidth || img.width
    const h = img.naturalHeight || img.height
    const canvas = document.createElement('canvas')
    canvas.width = Math.max(1, w)
    canvas.height = Math.max(1, h)
    const ctx = canvas.getContext('2d')
    ctx.drawImage(img, 0, 0)
    const data = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const cx = canvas.width * 0.5
    const cy = canvas.height * 0.42
    const rx = canvas.width * 0.36
    const ry = canvas.height * 0.48
    const feather = 0.12
    for (let y = 0; y < canvas.height; y++) {
      for (let x = 0; x < canvas.width; x++) {
        const nx = (x - cx) / rx
        const ny = (y - cy) / ry
        const d = Math.sqrt(nx * nx + ny * ny)
        let a = 0
        if (d <= 1 - feather) a = 1
        else if (d < 1) {
          const t = (d - (1 - feather)) / feather
          a = 1 - t * t * (3 - 2 * t)
        }
        data.data[(y * canvas.width + x) * 4 + 3] = Math.round(a * 255)
      }
    }
    ctx.putImageData(data, 0, 0)
    return canvas.toDataURL('image/png')
  } catch {
    return source || makeCanvasDataUrl(360, 480, '#f1f5f9', '抠图')
  }
}

function makeLayoutDataUrl(bg, label, mode = 'solid') {
  const canvas = document.createElement('canvas')
  canvas.width = 600
  canvas.height = 400
  const ctx = canvas.getContext('2d')
  ctx.fillStyle = '#f8fafc'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  const tileW = 90
  const tileH = 126
  const gap = 12
  const cols = 5
  const rows = 2
  const startX = 40
  const startY = 50
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      const x = startX + c * (tileW + gap)
      const y = startY + r * (tileH + gap)
      const tmp = document.createElement('canvas')
      tmp.width = tileW
      tmp.height = tileH
      const tctx = tmp.getContext('2d')
      fillBackground(tctx, tileW, tileH, bg, mode)
      tctx.fillStyle = '#475569'
      tctx.beginPath()
      tctx.ellipse(tileW / 2, tileH * 0.38, 18, 22, 0, 0, Math.PI * 2)
      tctx.fill()
      tctx.beginPath()
      tctx.ellipse(tileW / 2, tileH * 0.82, 28, 26, 0, Math.PI, 0)
      tctx.fill()
      ctx.drawImage(tmp, x, y)
    }
  }
  ctx.fillStyle = '#0f172a'
  ctx.font = '14px sans-serif'
  ctx.textAlign = 'center'
  ctx.fillText(label, canvas.width / 2, 28)
  return canvas.toDataURL('image/png')
}

async function tryGo(method, ...args) {
  // wailsjs/App.js 始终能 import，但浏览器里没有 window.go → 读 main 会抛错
  if (!isWailsRuntime()) {
    return undefined
  }
  let mod
  try {
    mod = await import('../../wailsjs/go/main/App.js')
  } catch {
    // 非 Wails 环境（纯浏览器预览）才回退 mock / HTTP
    return undefined
  }
  if (typeof mod[method] !== 'function') {
    return undefined
  }
  // Go 方法已绑定：错误必须抛出，禁止吞掉后落到 mock（否则会把检验后的原图当成抠图）
  return await mod[method](...args)
}

export const IDPhotoService = {
  TEMPLATE_PRESETS,

  async OpenImageDialog() {
    const res = await tryGo('OpenImageDialog')
    if (res !== undefined) return res
    // 浏览器 / 远程页：交给前端 input[type=file]
    return ''
  },

  async LoadImage(path) {
    // 已是 data URL 时直接用于预览，避免再走占位图
    if (typeof path === 'string' && path.startsWith('data:image/')) {
      const goRes = await tryGo('LoadImage', path)
      if (goRes !== undefined) return goRes
      if (await detectHttpMode()) {
        try {
          return await apiJSON('/load-image', { path })
        } catch {
          // 超大图 POST 失败时本地兜底预览
        }
      }
      return {
        imgBase64: path,
        faceBox: [90, 80, 270, 300],
        landmarks: [140, 160, 220, 160, 180, 210, 150, 250, 210, 250],
        report: {
          faceOk: true,
          faceScore: 0.92,
          isRephoto: false,
          isAiImage: false,
        },
      }
    }

    const res = await tryGo('LoadImage', path)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/load-image', { path })
    }
    const name = String(path).split(/[/\\]/).pop() || '示例照片'
    return {
      imgBase64: makeCanvasDataUrl(360, 480, '#f1f5f9', name),
      faceBox: [90, 80, 270, 300],
      landmarks: [140, 160, 220, 160, 180, 210, 150, 250, 210, 250],
      report: {
        faceOk: true,
        faceScore: 0.92,
        isRephoto: false,
        isAiImage: false,
      },
    }
  },

  async DetectFace(image, options = {}) {
    // Skip* defaults to false → all gates ON (safe if bindings drop fields).
    const opts = {
      skipPose: options.skipPose === true || options.checkPose === false,
      skipTilt: options.skipTilt === true || options.checkTilt === false,
      skipBlur: options.skipBlur === true || options.checkBlur === false,
      skipMosaic: options.skipMosaic === true || options.checkMosaic === false,
      skipParse: options.skipParse === true || options.checkParse === false,
    }
    const res = await tryGo('DetectFace', image, opts)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/detect-face', { path: image, options: opts })
    }
    throw new Error('人脸检测服务不可用')
  },

  async Generate(params) {
    const res = await tryGo('Generate', params)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      window.dispatchEvent(new CustomEvent('IDPhoto.OnProgress', { detail: { percent: 25, msg: '远程处理中...' } }))
      const result = await apiJSON('/generate', params)
      window.dispatchEvent(new CustomEvent('IDPhoto.OnProgress', { detail: { percent: 90, msg: '整理成品...' } }))
      return result
    }
    const steps = [
      { percent: 15, msg: '人脸检测中...' },
      { percent: 35, msg: '智能抠图中...' },
      { percent: 55, msg: '美颜处理中...' },
      { percent: 75, msg: '背景合成中...' },
      { percent: 90, msg: '排版输出中...' },
    ]
    for (const step of steps) {
      window.dispatchEvent(new CustomEvent('IDPhoto.OnProgress', { detail: step }))
      await sleep(180)
    }
    const bg = params.bgColor || '#FFFFFF'
    const mode = params.bgMode || 'solid'
    const source = params.sourceImg || ''
    const paperMap = {
      '5inch': '5寸',
      '6inch': '6寸',
      '7inch': '7寸',
      a6: 'A6',
      a5: 'A5',
      a4: 'A4',
    }
    const paperLabel = paperMap[params.paperSize] || '6寸'
    const idphoto = source || makeCanvasDataUrl(295, 413, bg, '证件照', mode)
    const single = source || makeCanvasDataUrl(360, 480, bg, '单张照片', mode)
    const mattingImg = source
      ? await makeMattingDataUrl(source)
      : makeCanvasDataUrl(360, 480, '#f1f5f9', '抠图')
    return {
      originImg: source || makeCanvasDataUrl(360, 480, '#f1f5f9', '原图'),
      mattingImg,
      resultImg: idphoto,
      results: {
        single,
        layout: makeLayoutDataUrl(bg, `${paperLabel}排版照`, mode),
        social: source || makeCanvasDataUrl(1080, 1400, bg, '社交一', mode),
        social2: source || makeCanvasDataUrl(1080, 1440, bg, '社交二', mode),
        idphoto,
      },
      faceBox: [90, 80, 270, 300],
      landmarks: [140, 160, 220, 160, 180, 210, 150, 250, 210, 250],
      report: {
        faceOk: true,
        faceScore: 0.94,
        isRephoto: false,
        isAiImage: false,
      },
    }
  },

  async SelectFolder() {
    const res = await tryGo('SelectFolder')
    if (res !== undefined) return res
    return 'D:/IDPortrait/export'
  },

  async Export(dir, opt) {
    const res = await tryGo('Export', dir, opt)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/export', { dir, options: opt })
    }
    return { ok: true, dir, opt }
  },

  async ListPrinters() {
    const res = await tryGo('ListPrinters')
    if (res !== undefined) {
      return { native: true, printers: Array.isArray(res) ? res : [] }
    }
    return { native: false, printers: [] }
  },

  async PrintLayout(opt) {
    const res = await tryGo('PrintLayout', opt)
    if (res !== undefined) return res
    throw new Error('本机打印仅支持桌面客户端')
  },

  async GetHistory() {
    const res = await tryGo('GetHistory')
    if (res !== undefined) return res
    return [
      { name: '示例_一寸.jpg', thumb: makePlaceholderThumb('一寸'), path: 'mock://demo-1.jpg' },
      { name: '示例_护照.jpg', thumb: makePlaceholderThumb('护照', '#dbeafe'), path: 'mock://demo-2.jpg' },
      { name: '示例_教资.jpg', thumb: makePlaceholderThumb('教资', '#fee2e2'), path: 'mock://demo-3.jpg' },
    ]
  },

  async StartRemoteServer(port) {
    const res = await tryGo('StartRemoteServer', port)
    if (res !== undefined) return res
    throw new Error('远程服务仅可在桌面客户端中启动')
  },

  async StopRemoteServer() {
    const res = await tryGo('StopRemoteServer')
    if (res !== undefined) return res
    return null
  },

  async GetRemoteStatus() {
    const res = await tryGo('GetRemoteStatus')
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return { running: true, port: Number(location.port) || 8787, url: location.origin, addr: location.origin }
    }
    return { running: false, port: 8787, url: '', addr: '' }
  },

  async GetPhotoSpecs(query = {}) {
    const q = { keyword: query.keyword || '', category: query.category || '' }
    const res = await tryGo('GetPhotoSpecs', q)
    if (res !== undefined) return normalizePhotoCatalog(res) || FALLBACK_PHOTO_SPECS
    if (await detectHttpMode()) {
      const qs = new URLSearchParams()
      if (q.keyword) qs.set('keyword', q.keyword)
      if (q.category) qs.set('category', q.category)
      const suffix = qs.toString() ? `?${qs}` : ''
      const raw = await apiJSON(`/photo-specs${suffix}`, null, 'GET')
      return normalizePhotoCatalog(raw) || FALLBACK_PHOTO_SPECS
    }
    // offline mock: local filter
    const keyword = String(q.keyword || '').trim().toLowerCase()
    const category = String(q.category || '').trim()
    let list = FALLBACK_PHOTO_SPECS.list
    if (keyword) {
      list = list.filter((item) => `${item.title} ${item.desc} ${item.keywords}`.toLowerCase().includes(keyword))
    } else if (category && category !== 'custom') {
      list = list.filter((item) => item.categories.includes(category))
    }
    return { ...FALLBACK_PHOTO_SPECS, list }
  },

  async GetPaperSpecs(query = {}) {
    const q = { keyword: query.keyword || '', category: query.category || '' }
    const res = await tryGo('GetPaperSpecs', q)
    if (res !== undefined) return normalizePaperCatalog(res) || FALLBACK_PAPER_SPECS
    if (await detectHttpMode()) {
      const qs = new URLSearchParams()
      if (q.keyword) qs.set('keyword', q.keyword)
      const suffix = qs.toString() ? `?${qs}` : ''
      const raw = await apiJSON(`/paper-specs${suffix}`, null, 'GET')
      return normalizePaperCatalog(raw) || FALLBACK_PAPER_SPECS
    }
    const keyword = String(q.keyword || '').trim().toLowerCase()
    let list = FALLBACK_PAPER_SPECS.list
    if (keyword) {
      list = list.filter((item) => `${item.title} ${item.desc}`.toLowerCase().includes(keyword))
    }
    return { ...FALLBACK_PAPER_SPECS, list }
  },

  async SetCurrentPhotoSpec(value) {
    const res = await tryGo('SetCurrentPhotoSpec', value)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/photo-specs/current', { value })
    }
    return null
  },

  async SetCurrentPaperSpec(value) {
    const res = await tryGo('SetCurrentPaperSpec', value)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/paper-specs/current', { value })
    }
    return null
  },

  async GetFaceDetectModels(query = {}) {
    const q = { keyword: query.keyword || '' }
    const res = await tryGo('GetFaceDetectModels', q)
    if (res !== undefined) return normalizeModelCatalog(res) || FALLBACK_FACE_DETECT_MODELS
    if (await detectHttpMode()) {
      const qs = new URLSearchParams()
      if (q.keyword) qs.set('keyword', q.keyword)
      const suffix = qs.toString() ? `?${qs}` : ''
      const raw = await apiJSON(`/face-detect-models${suffix}`, null, 'GET')
      return normalizeModelCatalog(raw) || FALLBACK_FACE_DETECT_MODELS
    }
    const keyword = String(q.keyword || '').trim().toLowerCase()
    let list = FALLBACK_FACE_DETECT_MODELS.list
    if (keyword) {
      list = list.filter((item) => `${item.title} ${item.desc}`.toLowerCase().includes(keyword))
    }
    return { ...FALLBACK_FACE_DETECT_MODELS, list }
  },

  async GetMattingModels(query = {}) {
    const q = { keyword: query.keyword || '' }
    const res = await tryGo('GetMattingModels', q)
    if (res !== undefined) return normalizeModelCatalog(res) || FALLBACK_MATTING_MODELS
    if (await detectHttpMode()) {
      const qs = new URLSearchParams()
      if (q.keyword) qs.set('keyword', q.keyword)
      const suffix = qs.toString() ? `?${qs}` : ''
      const raw = await apiJSON(`/matting-models${suffix}`, null, 'GET')
      return normalizeModelCatalog(raw) || FALLBACK_MATTING_MODELS
    }
    const keyword = String(q.keyword || '').trim().toLowerCase()
    let list = FALLBACK_MATTING_MODELS.list
    if (keyword) {
      list = list.filter((item) => `${item.title} ${item.desc}`.toLowerCase().includes(keyword))
    }
    return { ...FALLBACK_MATTING_MODELS, list }
  },

  async SetCurrentFaceDetectModel(value) {
    const res = await tryGo('SetCurrentFaceDetectModel', value)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/face-detect-models/current', { value })
    }
    return null
  },

  async SetCurrentMattingModel(value) {
    const res = await tryGo('SetCurrentMattingModel', value)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/matting-models/current', { value })
    }
    return null
  },

  async GetWatermarkConfig() {
    const res = await tryGo('GetWatermarkConfig')
    if (res !== undefined) return normalizeWatermarkConfig(res)
    if (await detectHttpMode()) {
      const raw = await apiJSON('/watermark-config', null, 'GET')
      return normalizeWatermarkConfig(raw)
    }
    return {
      default: FALLBACK_WATERMARK,
      current: null,
    }
  },

  async SetWatermarkConfig(settings) {
    const payload = {
      enabled: !!settings.enabled,
      text: settings.text || '',
      color: settings.color || '#FFFFFF',
      fontSize: Number(settings.fontSize) || 18,
      opacity: Number(settings.opacity) || 0.28,
      angle: Number(settings.angle) || 0,
      spacing: Number(settings.spacing) || 120,
    }
    const res = await tryGo('SetWatermarkConfig', payload)
    if (res !== undefined) return res
    if (await detectHttpMode()) {
      return apiJSON('/watermark-config', payload)
    }
    return null
  },

  preferValue,
}
