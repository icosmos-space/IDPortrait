<script setup>
import { ref, watch, nextTick, onBeforeUnmount } from 'vue'
import { useMessage } from 'naive-ui'
import { detectFaces, loadFaceModels, qualifyDetections, toCaptureGeometry } from '../../services/faceDetect'

const props = defineProps({
  show: { type: Boolean, default: false },
})

const emit = defineEmits(['update:show', 'capture'])

const message = useMessage()
const cameraVideo = ref(null)
const overlayCanvas = ref(null)
const cameraStream = ref(null)
const cameraReady = ref(false)
const cameraError = ref('')
const capturing = ref(false)
const fileInput = ref(null)
const faceOk = ref(false)
const faceReason = ref('正在加载人脸模型…')

let detectToken = 0
let latestFace = null

const CONTOURS = [
  [0, 16, false],
  [17, 21, false],
  [22, 26, false],
  [27, 30, false],
  [31, 35, false],
  [36, 41, true],
  [42, 47, true],
  [48, 59, true],
  [60, 67, true],
]

function stopDetect() {
  detectToken += 1
  latestFace = null
  faceOk.value = false
  clearOverlay()
}

function clearOverlay() {
  const canvas = overlayCanvas.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  ctx?.clearRect(0, 0, canvas.width, canvas.height)
}

function stopCamera() {
  stopDetect()
  if (cameraStream.value) {
    cameraStream.value.getTracks().forEach((t) => t.stop())
    cameraStream.value = null
  }
  if (cameraVideo.value) {
    cameraVideo.value.srcObject = null
  }
  cameraReady.value = false
}

function cameraUnavailableReason() {
  if (typeof window !== 'undefined' && window.isSecureContext === false) {
    return '当前页面不是安全连接（需 HTTPS）。请使用设置中的 https 地址访问，并在浏览器中信任证书后再试。'
  }
  if (!navigator.mediaDevices?.getUserMedia) {
    return '当前环境不支持摄像头。请改用 HTTPS 访问，或从相册选择照片。'
  }
  return ''
}

function drawOverlay(detections, ok) {
  const video = cameraVideo.value
  const canvas = overlayCanvas.value
  if (!video || !canvas) return
  const w = video.videoWidth
  const h = video.videoHeight
  if (!w || !h) return
  if (canvas.width !== w) canvas.width = w
  if (canvas.height !== h) canvas.height = h
  const ctx = canvas.getContext('2d')
  ctx.clearRect(0, 0, w, h)
  const color = ok ? '#3dbe7a' : '#ff6b8a'
  const radius = Math.max(1.6, w / 420)
  ctx.lineWidth = Math.max(1.2, w / 480)
  ctx.strokeStyle = color
  ctx.fillStyle = color

  for (const face of detections) {
    const box = face.detection.box
    ctx.strokeRect(box.x, box.y, box.width, box.height)
    const pts = face.landmarks.positions
    ctx.beginPath()
    for (const [start, end, closed] of CONTOURS) {
      ctx.moveTo(pts[start].x, pts[start].y)
      for (let i = start + 1; i <= end; i += 1) {
        ctx.lineTo(pts[i].x, pts[i].y)
      }
      if (closed) ctx.closePath()
    }
    ctx.stroke()
    for (const p of pts) {
      ctx.beginPath()
      ctx.arc(p.x, p.y, radius, 0, Math.PI * 2)
      ctx.fill()
    }
  }
}

async function runDetectLoop(token) {
  const video = cameraVideo.value
  if (!video || token !== detectToken) return
  if (video.readyState < 2) {
    setTimeout(() => runDetectLoop(token), 120)
    return
  }
  try {
    const detections = await detectFaces(video)
    if (token !== detectToken) return
    const verdict = qualifyDetections(detections, video.videoWidth, video.videoHeight)
    latestFace = verdict.ok ? verdict.face : null
    faceOk.value = verdict.ok
    faceReason.value = verdict.reason
    drawOverlay(detections, verdict.ok)
  } catch (err) {
    if (token !== detectToken) return
    latestFace = null
    faceOk.value = false
    faceReason.value = err?.message || '人脸检测失败'
  }
  if (token === detectToken) {
    setTimeout(() => runDetectLoop(token), 90)
  }
}

async function startDetect() {
  const token = ++detectToken
  faceOk.value = false
  faceReason.value = '正在加载人脸模型…'
  try {
    await loadFaceModels()
  } catch (err) {
    if (token !== detectToken) return
    faceReason.value = err?.message || '人脸模型加载失败'
    return
  }
  if (token !== detectToken) return
  faceReason.value = '正在检测人脸…'
  runDetectLoop(token)
}

async function startCamera() {
  stopCamera()
  cameraError.value = ''
  cameraReady.value = false
  faceReason.value = '正在打开摄像头…'
  const reason = cameraUnavailableReason()
  if (reason) {
    cameraError.value = reason
    faceReason.value = reason
    return
  }
  try {
    const stream = await navigator.mediaDevices.getUserMedia({
      video: {
        facingMode: 'user',
        width: { ideal: 1280 },
        height: { ideal: 720 },
      },
      audio: false,
    })
    cameraStream.value = stream
    if (cameraVideo.value) {
      cameraVideo.value.srcObject = stream
      await cameraVideo.value.play()
      cameraReady.value = true
      await startDetect()
    }
  } catch (err) {
    cameraError.value = err?.message || '无法打开摄像头，请检查权限'
    cameraReady.value = false
    faceReason.value = cameraError.value
  }
}

watch(
  () => props.show,
  async (visible) => {
    if (visible) {
      cameraError.value = ''
      cameraReady.value = false
      faceOk.value = false
      faceReason.value = '正在打开摄像头…'
      await nextTick()
      await startCamera()
    } else {
      stopCamera()
    }
  },
)

onBeforeUnmount(stopCamera)

function close() {
  emit('update:show', false)
}

function pickFromAlbum() {
  fileInput.value?.click()
}

function onFilePicked(e) {
  const file = e.target.files?.[0]
  e.target.value = ''
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    const dataUrl = String(reader.result || '')
    if (!dataUrl) {
      message.error('读取图片失败')
      return
    }
    const img = new Image()
    img.onload = () => {
      const w = img.naturalWidth || 1280
      const h = img.naturalHeight || 720
      const faceBox = [
        Math.round(w * 0.28),
        Math.round(h * 0.16),
        Math.round(w * 0.72),
        Math.round(h * 0.72),
      ]
      const cx = (faceBox[0] + faceBox[2]) / 2
      const cy = (faceBox[1] + faceBox[3]) / 2
      const landmarks = [
        cx - w * 0.08, cy - h * 0.06,
        cx + w * 0.08, cy - h * 0.06,
        cx, cy + h * 0.02,
        cx - w * 0.06, cy + h * 0.1,
        cx + w * 0.06, cy + h * 0.1,
      ]
      emit('capture', { dataUrl, faceBox, landmarks })
    }
    img.onerror = () => message.error('图片无法预览')
    img.src = dataUrl
  }
  reader.onerror = () => message.error('读取图片失败')
  reader.readAsDataURL(file)
}

function capturePhoto() {
  if (!cameraVideo.value || !cameraReady.value) {
    message.warning('摄像头尚未就绪')
    return
  }
  if (!faceOk.value || !latestFace) {
    message.warning(faceReason.value || '当前姿态不合格，不能拍照')
    return
  }
  capturing.value = true
  const video = cameraVideo.value
  const canvas = document.createElement('canvas')
  const w = video.videoWidth || 1280
  const h = video.videoHeight || 720
  canvas.width = w
  canvas.height = h
  const ctx = canvas.getContext('2d')
  ctx.translate(w, 0)
  ctx.scale(-1, 1)
  ctx.drawImage(video, 0, 0, w, h)
  const dataUrl = canvas.toDataURL('image/jpeg', 0.92)
  const { faceBox, landmarks } = toCaptureGeometry(latestFace, w)
  capturing.value = false
  emit('capture', { dataUrl, faceBox, landmarks })
}
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    title="摄像头拍照"
    style="width: 640px"
    :bordered="false"
    :segmented="{ content: true, footer: 'soft' }"
    :mask-closable="false"
    @update:show="(v) => emit('update:show', v)"
    @after-leave="stopCamera"
  >
    <div class="camera-body">
      <div class="camera-viewport">
        <video
          ref="cameraVideo"
          class="camera-video"
          autoplay
          playsinline
          muted
        />
        <canvas ref="overlayCanvas" class="camera-overlay" />
        <div class="camera-guide" aria-hidden="true" />
        <p
          v-if="!cameraError"
          class="camera-status"
          :data-ok="faceOk"
        >
          {{ faceReason }}
        </p>
        <p v-if="cameraError" class="camera-error">{{ cameraError }}</p>
        <p v-else-if="!cameraReady" class="camera-loading">正在打开摄像头…</p>
      </div>
      <p class="camera-tip">请单人正对镜头、面部居中且端正；检测到合格姿态后才能拍照</p>
      <input
        ref="fileInput"
        type="file"
        accept="image/*"
        capture="user"
        hidden
        @change="onFilePicked"
      />
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn ghost" type="button" @click="close">取消</button>
        <button class="btn ghost" type="button" @click="pickFromAlbum">相册选图</button>
        <button class="btn ghost" type="button" @click="startCamera">重试</button>
        <button
          class="btn primary"
          type="button"
          :disabled="!cameraReady || capturing || !faceOk"
          @click="capturePhoto"
        >
          {{ capturing ? '处理中…' : faceOk ? '拍照' : '不合格' }}
        </button>
      </div>
    </template>
  </n-modal>
</template>
