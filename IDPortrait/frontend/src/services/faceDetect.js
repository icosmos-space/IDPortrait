import * as faceapi from '@vladmandic/face-api'

const MODEL_URI = `${import.meta.env.BASE_URL}models`.replace(/\/{2,}/g, '/')

let loading = null

export function loadFaceModels() {
  if (!loading) {
    loading = (async () => {
      if (faceapi.tf?.setBackend) {
        await faceapi.tf.setBackend('webgl')
        await faceapi.tf.ready()
      }
      await Promise.all([
        faceapi.nets.tinyFaceDetector.loadFromUri(MODEL_URI),
        faceapi.nets.faceLandmark68Net.loadFromUri(MODEL_URI),
      ])
      return faceapi
    })().catch((err) => {
      loading = null
      throw err
    })
  }
  return loading
}

export async function detectFaces(input, { inputSize = 320, scoreThreshold = 0.45 } = {}) {
  const api = await loadFaceModels()
  const options = new api.TinyFaceDetectorOptions({
    inputSize,
    scoreThreshold,
  })
  return api.detectAllFaces(input, options).withFaceLandmarks()
}

function loadImageElement(src) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('图片无法读取'))
    img.src = src
  })
}

function rotateImage(img, degrees) {
  const w = img.naturalWidth || img.width
  const h = img.naturalHeight || img.height
  const canvas = document.createElement('canvas')
  const swap = degrees === 90 || degrees === 270
  canvas.width = swap ? h : w
  canvas.height = swap ? w : h
  const ctx = canvas.getContext('2d')
  ctx.translate(canvas.width / 2, canvas.height / 2)
  ctx.rotate((degrees * Math.PI) / 180)
  ctx.drawImage(img, -w / 2, -h / 2)
  return canvas
}

/** 静止图片：只接受恰好一张人脸。方向不对时再试 90/180/270，并转到人脸朝上。 */
export async function requirePerson(dataUrl, onProgress) {
  const report = (percent, msg) => {
    if (typeof onProgress === 'function') onProgress({ percent, msg })
  }
  report(8, '正在加载人脸模型…')
  await loadFaceModels()
  report(22, '正在读取照片…')
  const img = await loadImageElement(dataUrl)
  const angles = [0, 90, 270, 180]
  const labels = ['正在检测人脸…', '正在调整方向…', '正在继续检测…', '正在确认人脸…']
  let best = null
  let sawMultiple = false
  for (let i = 0; i < angles.length; i += 1) {
    const angle = angles[i]
    report(28 + Math.round((i / angles.length) * 62), labels[i])
    const source = angle === 0 ? img : rotateImage(img, angle)
    const detections = await detectFaces(source, { inputSize: 416, scoreThreshold: 0.35 })
    if (!detections?.length) continue
    if (detections.length > 1) {
      sawMultiple = true
      continue
    }
    const face = detections[0]
    const score = face.detection.score ?? 0
    const box = face.detection.box
    const area = box.width * box.height
    if (!best || score > best.score || (score === best.score && area > best.area)) {
      best = {
        ok: true,
        score,
        area,
        ...toImageGeometry(face),
        dataUrl: angle === 0 ? dataUrl : source.toDataURL('image/jpeg', 0.92),
      }
    }
    if (angle === 0 && score >= 0.55) break
  }
  report(100, best ? '检测完成' : '检测结束')
  if (best) return best
  if (sawMultiple) return { ok: false, reason: '检测到多张人脸，已拒绝' }
  return { ok: false, reason: '未检测到人脸，已拒绝' }
}

export function toImageGeometry(face) {
  const box = face.detection.box
  const faceBox = [
    Math.round(box.x),
    Math.round(box.y),
    Math.round(box.x + box.width),
    Math.round(box.y + box.height),
  ]
  const landmarks = []
  for (const p of face.landmarks.positions) {
    landmarks.push(p.x, p.y)
  }
  return { faceBox, landmarks }
}

function avg(points) {
  let x = 0
  let y = 0
  for (const p of points) {
    x += p.x
    y += p.y
  }
  const n = points.length || 1
  return { x: x / n, y: y / n }
}

export function qualifyDetections(detections, width, height) {
  if (!detections?.length) {
    return { ok: false, reason: '未检测到人脸', face: null }
  }
  if (detections.length > 1) {
    return { ok: false, reason: '检测到多张人脸，请只保留一人', face: detections[0] }
  }

  const face = detections[0]
  const score = face.detection?.score ?? 0
  if (score < 0.55) {
    return { ok: false, reason: '人脸不清晰，请正对镜头', face }
  }

  const box = face.detection.box
  const faceH = box.height / height
  const faceW = box.width / width
  if (faceH < 0.3 || faceW < 0.18) {
    return { ok: false, reason: '人脸过小，请靠近一些', face }
  }
  if (faceH > 0.72 || faceW > 0.62) {
    return { ok: false, reason: '人脸过大，请稍远一些', face }
  }
  if (
    box.x < width * 0.04
    || box.y < height * 0.02
    || box.x + box.width > width * 0.96
    || box.y + box.height > height * 0.98
  ) {
    return { ok: false, reason: '面部超出画面，请调整位置', face }
  }

  const cx = (box.x + box.width / 2) / width
  const cy = (box.y + box.height / 2) / height
  if (Math.abs(cx - 0.5) > 0.12) {
    return { ok: false, reason: '请将面部移到画面中央', face }
  }
  if (cy < 0.3 || cy > 0.58) {
    return { ok: false, reason: '请调整头部上下位置', face }
  }

  const lm = face.landmarks
  const leftEye = avg(lm.getLeftEye())
  const rightEye = avg(lm.getRightEye())
  const tilt = Math.abs(Math.atan2(rightEye.y - leftEye.y, rightEye.x - leftEye.x) * 180 / Math.PI)
  if (tilt > 10) {
    return { ok: false, reason: '头部倾斜过大，请端正', face }
  }

  const eyeDist = Math.hypot(rightEye.x - leftEye.x, rightEye.y - leftEye.y) || 1
  const nose = lm.getNose()
  const noseTip = nose[Math.min(3, nose.length - 1)]
  const midX = (leftEye.x + rightEye.x) / 2
  if (Math.abs(noseTip.x - midX) / eyeDist > 0.32) {
    return { ok: false, reason: '请正对镜头，不要侧脸', face }
  }

  return { ok: true, reason: '姿态合格，可以拍照', face }
}

/** Map detector coords onto the horizontally flipped capture image. */
export function toCaptureGeometry(face, width) {
  const box = face.detection.box
  const faceBox = [
    Math.round(width - (box.x + box.width)),
    Math.round(box.y),
    Math.round(width - box.x),
    Math.round(box.y + box.height),
  ]
  const landmarks = []
  for (const p of face.landmarks.positions) {
    landmarks.push(width - p.x, p.y)
  }
  return { faceBox, landmarks }
}
