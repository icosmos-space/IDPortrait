import { ref, reactive, onMounted, onBeforeUnmount, watch, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { IDPhotoService } from '../services/idphoto'
import {
  RESULT_TABS,
  BG_MODES,
  BG_PRESETS,
  CLOTH_OPTIONS,
  PAPER_SIZES,
  SPEC_CATEGORIES,
  ALL_SPECS,
  ABOUT_INFO,
  APP_VERSION,
  createDefaultParams,
  modePreviewStyle,
  formatVal,
} from '../constants/studio'

export function useStudio() {
  const message = useMessage()

  const expandedPanels = ref(['beauty'])
  const processing = ref(false)
  const progressPercent = ref(0)
  const statusText = ref('就绪')
  const processTagType = ref('success')
  const hasOrigin = ref(false)
  const hasResult = ref(false)
  const activeResultTab = ref('idphoto')

  const resultTabs = RESULT_TABS

  /** @type {import('vue').Ref<{ img: string, faceBox: number[], landmarks: number[] } | null>} */
  const originView = ref(null)
  /** @type {import('vue').Ref<string | null>} */
  const resultView = ref(null)

  const resultSet = reactive({
    single: '',
    layout: '',
    social: '',
    idphoto: '',
  })

  const currentResultHint = computed(() => {
    return resultTabs.find((t) => t.key === activeResultTab.value)?.hint || '成品预览'
  })

  const params = reactive(createDefaultParams())

  const bgModes = BG_MODES
  const clothOptions = CLOTH_OPTIONS
  const paperSizes = PAPER_SIZES
  const specCategories = SPEC_CATEGORIES
  const allSpecs = ALL_SPECS
  const bgPresets = BG_PRESETS

  const currentPaperLabel = computed(() => {
    return paperSizes.find((p) => p.value === params.paperSize)?.title || '6寸'
  })

  const specKeyword = ref('')
  const activeSpecCategory = ref('common')

  const filteredSpecs = computed(() => {
    if (activeSpecCategory.value === 'custom') return []
    const q = specKeyword.value.trim().toLowerCase()
    return allSpecs.filter((item) => {
      const inCategory = !q ? item.categories.includes(activeSpecCategory.value) : true
      if (!inCategory) return false
      if (!q) return true
      const hay = `${item.title} ${item.desc} ${item.keywords}`.toLowerCase()
      return hay.includes(q)
    })
  })

  const showCustomSpec = computed(
    () => activeSpecCategory.value === 'custom' && !specKeyword.value.trim(),
  )

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
  const showCameraModal = ref(false)
  const appVersion = APP_VERSION
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
    remoteEnabled: false,
    remotePort: 8787,
  })
  const remoteStatus = reactive({
    running: false,
    port: 8787,
    url: '',
    addr: '',
  })
  const remoteBusy = ref(false)
  const aboutInfo = ABOUT_INFO

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

  function setOriginView(img, faceBox = [], landmarks = []) {
    if (!img) {
      originView.value = null
      hasOrigin.value = false
      return
    }
    originView.value = {
      img,
      faceBox: Array.isArray(faceBox) ? faceBox : [],
      landmarks: Array.isArray(landmarks) ? landmarks : [],
      _ts: Date.now(),
    }
    hasOrigin.value = true
  }

  function normalizeLoadResult(ret) {
    if (!ret || typeof ret !== 'object') return null
    const report = ret.report || ret.Report || {}
    return {
      imgBase64: ret.imgBase64 || ret.ImgBase64 || ret.img || '',
      faceBox: ret.faceBox || ret.FaceBox || [],
      landmarks: ret.landmarks || ret.Landmarks || [],
      report: {
        faceOk: report.faceOk ?? report.FaceOK ?? false,
        faceScore: report.faceScore ?? report.FaceScore ?? 0,
        isRephoto: report.isRephoto ?? report.IsRephoto ?? false,
        isAiImage: report.isAiImage ?? report.IsAiImage ?? false,
      },
    }
  }

  function pickImageFile() {
    return new Promise((resolve) => {
      const input = document.createElement('input')
      input.type = 'file'
      input.accept = 'image/*'
      input.style.display = 'none'
      const cleanup = () => {
        input.remove()
      }
      input.addEventListener('change', () => {
        const file = input.files?.[0]
        cleanup()
        if (!file) {
          resolve('')
          return
        }
        const reader = new FileReader()
        reader.onload = () => resolve(String(reader.result || ''))
        reader.onerror = () => resolve('')
        reader.readAsDataURL(file)
      })
      input.addEventListener('cancel', () => {
        cleanup()
        resolve('')
      })
      document.body.appendChild(input)
      input.click()
    })
  }

  function readFileAsDataURL(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result || ''))
      reader.onerror = () => reject(new Error('读取图片失败'))
      reader.readAsDataURL(file)
    })
  }

  function showActiveResult() {
    const img = resultSet[activeResultTab.value]
    if (img) {
      resultView.value = img
      hasResult.value = true
      return
    }
    resultView.value = null
  }

  function clearViews() {
    originView.value = null
    resultView.value = null
    hasOrigin.value = false
    hasResult.value = false
    resultSet.single = ''
    resultSet.layout = ''
    resultSet.social = ''
    resultSet.idphoto = ''
  }

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
    setOriginView(payload.originImg, payload.faceBox || [], payload.landmarks || [])

    const next = payload?.results || {}
    resultSet.single = next.single || payload.resultImg || ''
    resultSet.layout = next.layout || ''
    resultSet.social = next.social || ''
    resultSet.idphoto = next.idphoto || payload.resultImg || ''
    hasResult.value = Boolean(resultSet.single || resultSet.layout || resultSet.social || resultSet.idphoto)
    showActiveResult()
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
      const { EventsOn } = await import('../../wailsjs/runtime/runtime')
      unbinders.push(EventsOn('IDPhoto.OnProgress', onProgress))
      unbinders.push(EventsOn('IDPhoto.OnDone', onDone))
      unbinders.push(EventsOn('IDPhoto.OnError', onError))
    } catch {
      // browser preview
    }
    window.addEventListener('IDPhoto.OnProgress', onMockProgress)
    window.addEventListener('IDPhoto.OnDone', onMockDone)
    window.addEventListener('IDPhoto.OnError', onMockError)
    clearViews()
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

  function previewModeStyle(mode) {
    return modePreviewStyle(mode, params.bgColor)
  }

  async function openImage() {
    try {
      const hasWailsDialog = typeof window !== 'undefined' && !!window.go?.main?.App?.OpenImageDialog
      if (hasWailsDialog) {
        const path = await IDPhotoService.OpenImageDialog()
        if (!path) return
        await loadImage(path)
        return
      }
      const dataUrl = await pickImageFile()
      if (!dataUrl) return
      await loadImage(dataUrl)
    } catch (e) {
      message.error(e?.message || '打开图片失败')
    }
  }

  function openCameraModal() {
    showCameraModal.value = true
  }

  function closeCameraModal() {
    showCameraModal.value = false
  }

  function applyCapture({ dataUrl, faceBox, landmarks }) {
    if (!dataUrl) {
      message.error('拍照结果为空')
      return
    }
    Object.assign(report, {
      faceOk: true,
      faceScore: 0.91,
      isRephoto: false,
      isAiImage: false,
    })
    setOriginView(dataUrl, faceBox, landmarks)
    statusText.value = '已拍照'
    processTagType.value = 'success'
    closeCameraModal()
    message.success('拍照成功')
  }

  async function loadImage(pathOrDataUrl) {
    try {
      let payload = pathOrDataUrl
      // 浏览器拖入的 File 或本地路径无法被远程服务读取时，先转 data URL
      if (typeof pathOrDataUrl === 'object' && pathOrDataUrl?.dataUrl) {
        payload = pathOrDataUrl.dataUrl
      }
      const raw = await IDPhotoService.LoadImage(payload)
      const ret = normalizeLoadResult(raw)
      if (!ret?.imgBase64) {
        throw new Error('未获取到图片数据')
      }
      Object.assign(report, ret.report)
      setOriginView(ret.imgBase64, ret.faceBox, ret.landmarks)
      statusText.value = '已加载图片'
      processTagType.value = 'success'
    } catch (e) {
      message.error(e?.message || '加载图片失败')
      statusText.value = '加载失败'
      processTagType.value = 'error'
    }
  }

  async function handleDrop(e) {
    const file = e.dataTransfer?.files?.[0]
    if (!file) return
    if (!String(file.type || '').startsWith('image/') && !/\.(jpe?g|png|bmp|webp|gif)$/i.test(file.name || '')) {
      message.warning('请拖入图片文件')
      return
    }
    try {
      // WebView / 浏览器通常没有 file.path，统一读成 data URL
      if (file.path) {
        await loadImage(file.path)
        return
      }
      const dataUrl = await readFileAsDataURL(file)
      await loadImage(dataUrl)
    } catch (err) {
      message.error(err?.message || '拖放加载失败')
    }
  }

  function resetAll() {
    Object.assign(params, createDefaultParams())
    activeSpecCategory.value = 'common'
    specKeyword.value = ''
    progressPercent.value = 0
    statusText.value = '就绪'
    processing.value = false
    processTagType.value = 'success'
    Object.assign(report, { faceOk: false, faceScore: 0, isRephoto: false, isAiImage: false })
    clearViews()
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

  function applyRemoteStatus(st) {
    if (!st || typeof st !== 'object') return
    remoteStatus.running = !!st.running
    remoteStatus.port = st.port || setting.remotePort
    remoteStatus.url = st.url || st.addr || ''
    remoteStatus.addr = st.addr || st.url || ''
    setting.remoteEnabled = !!st.running
    if (st.port) setting.remotePort = st.port
  }

  async function refreshRemoteStatus() {
    const st = await IDPhotoService.GetRemoteStatus()
    applyRemoteStatus(st)
    return st
  }

  async function startRemote() {
    remoteBusy.value = true
    try {
      const st = await IDPhotoService.StartRemoteServer(Number(setting.remotePort) || 8787)
      applyRemoteStatus(st)
      setting.remoteEnabled = true
      if (st?.url || st?.addr) {
        message.success(`远程服务已启动：${st.url || st.addr}`)
      } else {
        message.success('远程服务已启动')
      }
    } catch (e) {
      setting.remoteEnabled = false
      message.error(e?.message || '启动远程服务失败')
    } finally {
      remoteBusy.value = false
    }
  }

  async function stopRemote() {
    remoteBusy.value = true
    try {
      await IDPhotoService.StopRemoteServer()
      applyRemoteStatus({ running: false, port: setting.remotePort, url: '', addr: '' })
      setting.remoteEnabled = false
      message.info('远程服务已关闭')
    } catch (e) {
      message.error(e?.message || '关闭远程服务失败')
      await refreshRemoteStatus()
    } finally {
      remoteBusy.value = false
    }
  }

  function openSettingModal() {
    showSettingModal.value = true
    refreshRemoteStatus()
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
    message.success('设置已保存')
  }

  return {
    expandedPanels,
    processing,
    progressPercent,
    statusText,
    processTagType,
    hasOrigin,
    hasResult,
    activeResultTab,
    resultTabs,
    originView,
    resultView,
    resultSet,
    currentResultHint,
    params,
    bgModes,
    clothOptions,
    paperSizes,
    currentPaperLabel,
    specCategories,
    allSpecs,
    bgPresets,
    specKeyword,
    activeSpecCategory,
    filteredSpecs,
    showCustomSpec,
    report,
    showExportModal,
    showSettingModal,
    showAboutModal,
    showUpgradeModal,
    showCameraModal,
    appVersion,
    checkingUpgrade,
    exportDir,
    exportOpt,
    setting,
    remoteStatus,
    remoteBusy,
    aboutInfo,
    statusTone,
    diagnostics,
    previewModeStyle,
    formatVal,
    openImage,
    openCameraModal,
    closeCameraModal,
    applyCapture,
    loadImage,
    handleDrop,
    resetAll,
    runGenerate,
    onTemplateChange,
    applyCustomSpec,
    selectSpecCategory,
    switchResultTab,
    openExportModal,
    selectExportDir,
    doExport,
    openSettingModal,
    openAboutModal,
    openUpgradeModal,
    checkAndUpgrade,
    saveSetting,
    refreshRemoteStatus,
    startRemote,
    stopRemote,
  }
}
