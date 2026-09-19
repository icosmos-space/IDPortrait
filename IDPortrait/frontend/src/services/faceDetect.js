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

export async function detectFaces(video) {
  const api = await loadFaceModels()
  const options = new api.TinyFaceDetectorOptions({
    inputSize: 320,
    scoreThreshold: 0.45,
  })
  return api.detectAllFaces(video, options).withFaceLandmarks()
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
