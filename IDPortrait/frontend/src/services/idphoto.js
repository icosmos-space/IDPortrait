/**
 * 证件照服务层：优先调用 Wails Go 绑定，开发原型阶段回退到本地 mock。
 * 后续 Go 实现 OpenImageDialog / LoadImage / Generate / Export / SelectFolder 后即可无缝切换。
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

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
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
  try {
    const mod = await import('../../wailsjs/go/main/App.js')
    if (typeof mod[method] === 'function') {
      return await mod[method](...args)
    }
  } catch {
    // Go 绑定尚未生成或未实现，走 mock
  }
  return undefined
}

export const IDPhotoService = {
  TEMPLATE_PRESETS,

  async OpenImageDialog() {
    const res = await tryGo('OpenImageDialog')
    if (res !== undefined) return res
    // 原型：模拟选择一张图片路径
    return `mock://sample-${Date.now()}.jpg`
  },

  async LoadImage(path) {
    const res = await tryGo('LoadImage', path)
    if (res !== undefined) return res
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

  async Generate(params) {
    const res = await tryGo('Generate', params)
    if (res !== undefined) return res
    // 原型：通过自定义事件模拟 Go 推送进度（由 App.vue 监听）
    const steps = [
      { percent: 15, msg: '人脸检测中...' },
      { percent: 35, msg: '智能抠图中...' },
      { percent: 55, msg: '美颜处理中...' },
      { percent: 75, msg: '背景合成中...' },
      { percent: 90, msg: '排版输出中...' },
    ]
    for (const step of steps) {
      window.dispatchEvent(new CustomEvent('IDPhoto.OnProgress', { detail: step }))
      await sleep(280)
    }
    const bg = params.bgColor || '#FFFFFF'
    const mode = params.bgMode || 'solid'
    const paperMap = {
      '5inch': '5寸',
      '6inch': '6寸',
      '7inch': '7寸',
      a6: 'A6',
      a5: 'A5',
      a4: 'A4',
    }
    const paperLabel = paperMap[params.paperSize] || '6寸'
    const result = {
      originImg: makeCanvasDataUrl(360, 480, '#f1f5f9', '原图'),
      resultImg: makeCanvasDataUrl(295, 413, bg, '证件照', mode),
      results: {
        single: makeCanvasDataUrl(360, 480, bg, '单张照片', mode),
        layout: makeLayoutDataUrl(bg, `${paperLabel}排版照`, mode),
        social: makeCanvasDataUrl(400, 400, bg, '社交照', mode),
        idphoto: makeCanvasDataUrl(295, 413, bg, '证件照', mode),
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
    window.dispatchEvent(new CustomEvent('IDPhoto.OnDone', { detail: result }))
    return result
  },

  async SelectFolder() {
    const res = await tryGo('SelectFolder')
    if (res !== undefined) return res
    return 'D:/IDPortrait/export'
  },

  async Export(dir, opt) {
    const res = await tryGo('Export', dir, opt)
    if (res !== undefined) return res
    return { ok: true, dir, opt }
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
}
