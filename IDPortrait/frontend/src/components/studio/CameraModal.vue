<script setup>
import { ref, watch, nextTick } from 'vue'
import { useMessage } from 'naive-ui'

const props = defineProps({
  show: { type: Boolean, default: false },
})

const emit = defineEmits(['update:show', 'capture'])

const message = useMessage()
const cameraVideo = ref(null)
const cameraStream = ref(null)
const cameraReady = ref(false)
const cameraError = ref('')
const capturing = ref(false)
const fileInput = ref(null)

function stopCamera() {
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

async function startCamera() {
  stopCamera()
  cameraError.value = ''
  cameraReady.value = false
  const reason = cameraUnavailableReason()
  if (reason) {
    cameraError.value = reason
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
    }
  } catch (err) {
    cameraError.value = err?.message || '无法打开摄像头，请检查权限'
    cameraReady.value = false
  }
}

watch(
  () => props.show,
  async (visible) => {
    if (visible) {
      cameraError.value = ''
      cameraReady.value = false
      await nextTick()
      await startCamera()
    } else {
      stopCamera()
    }
  },
)

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
        <div class="camera-guide" aria-hidden="true" />
        <p v-if="cameraError" class="camera-error">{{ cameraError }}</p>
        <p v-else-if="!cameraReady" class="camera-loading">正在打开摄像头…</p>
      </div>
      <p class="camera-tip">请正对镜头，保持面部居中后点击拍照；也可用相册选图</p>
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
          :disabled="!cameraReady || capturing"
          @click="capturePhoto"
        >
          {{ capturing ? '处理中…' : '拍照' }}
        </button>
      </div>
    </template>
  </n-modal>
</template>
