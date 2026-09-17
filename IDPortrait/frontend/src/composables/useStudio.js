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
  })
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
    originView.value = { img, faceBox: faceBox || [], landmarks: landmarks || [] }
    hasOrigin.value = true
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
    const res = await IDPhotoService.OpenImageDialog()
    if (!res) return
    await loadImage(res)
  }

  function openCameraModal() {
    showCameraModal.value = true
  }

  function closeCameraModal() {
    showCameraModal.value = false
  }

  function applyCapture({ dataUrl, faceBox, landmarks }) {
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

  async function loadImage(path) {
    const ret = await IDPhotoService.LoadImage(path)
    Object.assign(report, ret.report)
    setOriginView(ret.imgBase64, ret.faceBox, ret.landmarks)
    statusText.value = '已加载图片'
    processTagType.value = 'success'
  }

  function handleDrop(e) {
    const file = e.dataTransfer?.files?.[0]
    if (!file) return
    loadImage(file.path || file.name)
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
  }
}
