import { ref, reactive, onMounted, onBeforeUnmount, watch, computed, h } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { IDPhotoService } from '../services/idphoto'
import {
  RESULT_TABS,
  SOCIAL_OPTIONS,
  ORIGIN_TABS,
  BG_MODES,
  BG_PRESETS,
  CLOTH_OPTIONS,
  ABOUT_INFO,
  APP_VERSION,
  createDefaultParams,
  modePreviewStyle,
  formatVal,
} from '../constants/studio'

export function useStudio() {
  const message = useMessage()
  const dialog = useDialog()

  const expandedPanels = ref(['beauty'])
  const processing = ref(false)
  const progressPercent = ref(0)
  const statusText = ref('就绪')
  const processTagType = ref('success')
  const hasOrigin = ref(false)
  const hasResult = ref(false)
  const activeResultTab = ref('idphoto')
  const activeOriginTab = ref('original')

  const resultTabs = RESULT_TABS
  const originTabs = ORIGIN_TABS

  /** @type {import('vue').Ref<{ img: string, faceBox: number[], landmarks: number[] } | null>} */
  const originView = ref(null)
  /** @type {import('vue').Ref<string | null>} */
  const mattingView = ref(null)
  /** @type {import('vue').Ref<string | null>} */
  const resultView = ref(null)

  const resultSet = reactive({
    single: '',
    layout: '',
    social: '',
    social2: '',
    idphoto: '',
  })

  const currentResultHint = computed(() => {
    return resultTabs.find((t) => t.key === activeResultTab.value)?.hint || '成品预览'
  })

  const activeSocial = ref('social')
  const socialView = computed(() => resultSet[activeSocial.value] || '')
  const socialHint = computed(() => {
    return SOCIAL_OPTIONS.find((t) => t.key === activeSocial.value)?.hint || '社交照'
  })

  const currentOriginHint = computed(() => {
    return originTabs.find((t) => t.key === activeOriginTab.value)?.hint || '原图预览'
  })

  const hasMatting = computed(() => Boolean(mattingView.value))

  const params = reactive(createDefaultParams())

  const bgModes = BG_MODES
  const clothOptions = CLOTH_OPTIONS
  const bgPresets = BG_PRESETS

  const paperSizes = ref([])
  const faceDetectModels = ref([])
  const mattingModels = ref([])
  const specCategories = ref([])
  const allSpecs = ref([])
  const photoSpecMeta = reactive({ default: '', current: '' })
  const paperSpecMeta = reactive({ default: '', current: '' })
  const faceDetectMeta = reactive({ default: '', current: '' })
  const mattingMeta = reactive({ default: '', current: '' })

  const currentPaperLabel = computed(() => {
    return paperSizes.value.find((p) => p.value === params.paperSize)?.title || '纸张'
  })

  const specKeyword = ref('')
  const activeSpecCategory = ref('common')
  let specSearchTimer = null

  const filteredSpecs = computed(() => {
    if (activeSpecCategory.value === 'custom' && !specKeyword.value.trim()) return []
    return allSpecs.value
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
  const showPrintModal = ref(false)
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
  const printCopies = ref(1)
  const printPrinterName = ref('')
  const printPaperSize = ref('6inch')
  const printLandscape = ref(false)
  const printBusy = ref(false)
  const printers = ref([])
  const nativePrint = ref(false)
  const setting = reactive({
    modelDir: '',
    cacheDir: '',
    enableRephotoDetect: true,
    enableAiDetect: true,
    remoteEnabled: false,
    remotePort: 8787,
    progressStyle: 'ghost',
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

  function normalizeFaceCheck(ret) {
    if (!ret || typeof ret !== 'object') {
      return { ok: false, reason: '未检测到人脸，已拒绝' }
    }
    return {
      ok: !!(ret.ok ?? ret.OK),
      reason: ret.reason || ret.Reason || '',
      dataUrl: ret.imgBase64 || ret.ImgBase64 || '',
      faceBox: ret.faceBox || ret.FaceBox || [],
      landmarks: ret.landmarks || ret.Landmarks || [],
      score: ret.score ?? ret.Score ?? 0,
    }
  }

  function showFaceReject(person, previewUrl = '') {
    const many = String(person?.reason || '').includes('多张')
    const title = many ? '人脸太多' : '没有人脸'
    const detail = many
      ? '这张图片里有多张人脸，不能制作证件照。请换一张只有一个人的照片。'
      : '这张图片里没有检测到人脸，不能制作证件照。请换一张正面单人照片。'
    const thumb = previewUrl || person?.dataUrl || ''
    statusText.value = title
    processTagType.value = 'error'
    dialog.error({
      title,
      content: () => h('div', { class: 'face-reject-body' }, [
        thumb
          ? h('div', { class: 'face-reject-thumb-wrap' }, [
              h('img', { class: 'face-reject-thumb', src: thumb, alt: '所选图片' }),
            ])
          : null,
        h('p', { class: 'face-reject-text' }, detail),
      ]),
      positiveText: '知道了',
      closable: false,
      maskClosable: false,
      style: { width: '480px' },
    })
  }

  async function detectBeforePreview(dataUrl) {
    statusText.value = '正在检测人脸…'
    processTagType.value = 'warning'
    processing.value = true
    progressPercent.value = 12
    try {
      const person = normalizeFaceCheck(await IDPhotoService.DetectFace(dataUrl))
      progressPercent.value = person.ok ? 100 : 0
      return person
    } catch (e) {
      progressPercent.value = 0
      throw e
    } finally {
      processing.value = false
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
    mattingView.value = null
    resultView.value = null
    hasOrigin.value = false
    hasResult.value = false
    activeOriginTab.value = 'original'
    resultSet.single = ''
    resultSet.layout = ''
    resultSet.social = ''
    resultSet.social2 = ''
    resultSet.idphoto = ''
  }

  function clearResults() {
    mattingView.value = null
    resultView.value = null
    hasResult.value = false
    resultSet.single = ''
    resultSet.layout = ''
    resultSet.social = ''
    resultSet.social2 = ''
    resultSet.idphoto = ''
    progressPercent.value = 0
  }

  function onProgress(data) {
    const payload = data?.percent !== undefined ? data : data?.[0] || data
    progressPercent.value = payload.percent ?? 0
    statusText.value = payload.msg || statusText.value
  }

  function onDone(res) {
    const payload = normalizeGenerateResult(res)
    processing.value = false
    progressPercent.value = 100
    statusText.value = '生成完成'
    processTagType.value = 'success'
    if (!payload) {
      message.error('生成结果为空')
      return
    }
    if (payload.report) Object.assign(report, payload.report)

    // 保留当前真实原图；仅在没有原图时用返回值回填
    if (!originView.value?.img && payload.originImg) {
      setOriginView(payload.originImg, payload.faceBox, payload.landmarks)
    }

    mattingView.value = payload.mattingImg || ''

    const next = payload.results || {}
    resultSet.single = next.single || payload.resultImg || ''
    resultSet.layout = next.layout || ''
    resultSet.social = next.social || ''
    resultSet.social2 = next.social2 || ''
    resultSet.idphoto = next.idphoto || next.IDPhoto || payload.resultImg || ''
    hasResult.value = Boolean(resultSet.single || resultSet.layout || resultSet.social || resultSet.social2 || resultSet.idphoto)
    if (!hasResult.value) {
      message.warning('未生成成品图')
      processTagType.value = 'warning'
      statusText.value = '生成完成但无成品'
      return
    }
    showActiveResult()
    message.success('生成完成')
  }

  function normalizeGenerateResult(res) {
    if (!res || typeof res !== 'object') return null
    const payload = res.resultImg !== undefined || res.ResultImg !== undefined || res.results || res.Results
      ? res
      : res[0] || res
    if (!payload || typeof payload !== 'object') return null
    const results = payload.results || payload.Results || {}
    const report = payload.report || payload.Report || {}
    return {
      originImg: payload.originImg || payload.OriginImg || '',
      mattingImg: payload.mattingImg || payload.MattingImg || '',
      resultImg: payload.resultImg || payload.ResultImg || '',
      results: {
        single: results.single || results.Single || '',
        layout: results.layout || results.Layout || '',
        social: results.social || results.Social || '',
        social2: results.social2 || results.Social2 || '',
        idphoto: results.idphoto || results.IDPhoto || '',
      },
      faceBox: payload.faceBox || payload.FaceBox || [],
      landmarks: payload.landmarks || payload.Landmarks || [],
      report: {
        faceOk: report.faceOk ?? report.FaceOK ?? false,
        faceScore: report.faceScore ?? report.FaceScore ?? 0,
        isRephoto: report.isRephoto ?? report.IsRephoto ?? false,
        isAiImage: report.isAiImage ?? report.IsAiImage ?? false,
      },
    }
  }

  function switchResultTab(key) {
    if (!resultTabs.some((t) => t.key === key)) return
    activeResultTab.value = key
    if (hasResult.value) showActiveResult()
  }

  function switchSocial(key) {
    if (SOCIAL_OPTIONS.some((t) => t.key === key)) activeSocial.value = key
  }

  function switchOriginTab(key) {
    activeOriginTab.value = key
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

  async function refreshPhotoSpecs() {
    const catalog = await IDPhotoService.GetPhotoSpecs({
      keyword: specKeyword.value.trim(),
      category: specKeyword.value.trim() ? '' : activeSpecCategory.value,
    })
    allSpecs.value = catalog?.list || []
    if (catalog?.categories?.length) {
      specCategories.value = catalog.categories
    }
    photoSpecMeta.default = catalog?.default || ''
    photoSpecMeta.current = catalog?.current || ''
    return catalog
  }

  async function refreshPaperSpecs(keyword = '') {
    const catalog = await IDPhotoService.GetPaperSpecs({ keyword })
    paperSizes.value = catalog?.list || []
    paperSpecMeta.default = catalog?.default || ''
    paperSpecMeta.current = catalog?.current || ''
    return catalog
  }

  function applyPreferredSpecs(photoCatalog, paperCatalog, faceCatalog, mattingCatalog) {
    const photoID = IDPhotoService.preferValue(photoCatalog?.current, photoCatalog?.default)
    const paperID = IDPhotoService.preferValue(paperCatalog?.current, paperCatalog?.default)
    const faceID = IDPhotoService.preferValue(faceCatalog?.current, faceCatalog?.default)
    const mattingID = IDPhotoService.preferValue(mattingCatalog?.current, mattingCatalog?.default)
    if (photoID) params.template = photoID
    if (paperID) params.paperSize = paperID
    if (faceID) params.faceDetectModel = faceID
    if (mattingID) params.mattingModel = mattingID
    if (photoID && photoID !== 'custom') {
      const hit = (photoCatalog?.list || []).find((s) => s.value === photoID)
      if (hit?.categories?.length) {
        activeSpecCategory.value = hit.categories[0]
      }
    }
  }

  async function loadCatalogs() {
    const [photoFull, paperFull, faceFull, mattingFull, watermarkCfg] = await Promise.all([
      IDPhotoService.GetPhotoSpecs({ keyword: '', category: '' }),
      IDPhotoService.GetPaperSpecs({ keyword: '' }),
      IDPhotoService.GetFaceDetectModels({ keyword: '' }),
      IDPhotoService.GetMattingModels({ keyword: '' }),
      IDPhotoService.GetWatermarkConfig(),
    ])
    if (photoFull?.categories?.length) {
      specCategories.value = photoFull.categories
    }
    photoSpecMeta.default = photoFull?.default || ''
    photoSpecMeta.current = photoFull?.current || ''
    paperSpecMeta.default = paperFull?.default || ''
    paperSpecMeta.current = paperFull?.current || ''
    faceDetectMeta.default = faceFull?.default || ''
    faceDetectMeta.current = faceFull?.current || ''
    mattingMeta.default = mattingFull?.default || ''
    mattingMeta.current = mattingFull?.current || ''
    paperSizes.value = paperFull?.list || []
    faceDetectModels.value = faceFull?.list || []
    mattingModels.value = mattingFull?.list || []
    applyPreferredSpecs(photoFull, paperFull, faceFull, mattingFull)
    applyWatermarkConfig(watermarkCfg)
    await refreshPhotoSpecs()
  }

  function applyWatermarkConfig(cfg) {
    const wm = cfg?.current || cfg?.default
    if (!wm) return
    params.enableWatermark = !!wm.enabled
    params.watermarkText = wm.text || '最美证件照'
    params.watermarkColor = wm.color || '#FFFFFF'
    params.watermarkFontSize = Number(wm.fontSize) || 18
    params.watermarkOpacity = Number(wm.opacity) || 0.28
    params.watermarkAngle = Number(wm.angle) || -30
    params.watermarkSpacing = Number(wm.spacing) || 120
  }

  function currentWatermarkPayload() {
    return {
      enabled: !!params.enableWatermark,
      text: params.watermarkText,
      color: params.watermarkColor,
      fontSize: params.watermarkFontSize,
      opacity: params.watermarkOpacity,
      angle: params.watermarkAngle,
      spacing: params.watermarkSpacing,
    }
  }

  let watermarkSaveTimer = null
  let watermarkReady = false
  function scheduleSaveWatermark() {
    if (!watermarkReady) return
    clearTimeout(watermarkSaveTimer)
    watermarkSaveTimer = setTimeout(() => {
      IDPhotoService.SetWatermarkConfig(currentWatermarkPayload()).catch(() => {})
    }, 280)
  }

  onMounted(async () => {
    try {
      setting.progressStyle = localStorage.getItem('idportrait.progressStyle') === 'card' ? 'card' : 'ghost'
    } catch {
      setting.progressStyle = 'ghost'
    }
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
    try {
      await loadCatalogs()
    } catch (e) {
      message.warning(e?.message || '规格列表加载失败，已使用本地兜底')
    } finally {
      watermarkReady = true
    }
  })

  onBeforeUnmount(() => {
    clearTimeout(specSearchTimer)
    clearTimeout(watermarkSaveTimer)
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

  watch(
    () => specKeyword.value,
    () => {
      clearTimeout(specSearchTimer)
      specSearchTimer = setTimeout(() => {
        refreshPhotoSpecs().catch(() => {})
      }, 180)
    },
  )

  watch(
    () => params.paperSize,
    (value) => {
      if (!value) return
      IDPhotoService.SetCurrentPaperSpec(value).catch(() => {})
      paperSpecMeta.current = value
    },
  )

  watch(
    () => params.faceDetectModel,
    (value) => {
      if (!value) return
      IDPhotoService.SetCurrentFaceDetectModel(value).catch(() => {})
      faceDetectMeta.current = value
    },
  )

  watch(
    () => params.mattingModel,
    (value) => {
      if (!value) return
      IDPhotoService.SetCurrentMattingModel(value).catch(() => {})
      mattingMeta.current = value
    },
  )

  watch(
    () => [
      params.enableWatermark,
      params.watermarkText,
      params.watermarkColor,
      params.watermarkFontSize,
      params.watermarkOpacity,
      params.watermarkAngle,
      params.watermarkSpacing,
    ],
    () => scheduleSaveWatermark(),
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

  async function applyCapture({ dataUrl }) {
    if (!dataUrl) {
      message.error('拍照结果为空')
      return
    }
    try {
      const person = await detectBeforePreview(dataUrl)
      if (!person?.ok) {
        showFaceReject(person, dataUrl)
        return
      }
      Object.assign(report, {
        faceOk: true,
        faceScore: person.score || 0.9,
        isRephoto: false,
        isAiImage: false,
      })
      clearResults()
      setOriginView(person.dataUrl || dataUrl, person.faceBox, person.landmarks)
      activeOriginTab.value = 'original'
      statusText.value = '已拍照'
      processTagType.value = 'success'
      closeCameraModal()
      message.success('拍照成功')
    } catch (e) {
      message.error(e?.message || '人脸检测失败')
      statusText.value = '检测失败'
      processTagType.value = 'error'
    }
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
      const person = await detectBeforePreview(ret.imgBase64)
      if (!person?.ok) {
        showFaceReject(person, ret.imgBase64)
        return
      }
      Object.assign(report, { ...ret.report, faceOk: true, faceScore: person.score || ret.report.faceScore })
      clearResults()
      setOriginView(person.dataUrl || ret.imgBase64, person.faceBox, person.landmarks)
      activeOriginTab.value = 'original'
      statusText.value = '已加载图片'
      processTagType.value = 'success'
    } catch (e) {
      processing.value = false
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
    watermarkReady = false
    Object.assign(params, createDefaultParams())
    const photoID = IDPhotoService.preferValue(photoSpecMeta.current, photoSpecMeta.default)
    const paperID = IDPhotoService.preferValue(paperSpecMeta.current, paperSpecMeta.default)
    const faceID = IDPhotoService.preferValue(faceDetectMeta.current, faceDetectMeta.default)
    const mattingID = IDPhotoService.preferValue(mattingMeta.current, mattingMeta.default)
    if (photoID) params.template = photoID
    if (paperID) params.paperSize = paperID
    if (faceID) params.faceDetectModel = faceID
    if (mattingID) params.mattingModel = mattingID
    IDPhotoService.GetWatermarkConfig()
      .then((cfg) => {
        applyWatermarkConfig(cfg)
        watermarkReady = true
      })
      .catch(() => {
        watermarkReady = true
      })
    activeSpecCategory.value = 'common'
    specKeyword.value = ''
    progressPercent.value = 0
    statusText.value = '就绪'
    processing.value = false
    processTagType.value = 'success'
    Object.assign(report, { faceOk: false, faceScore: 0, isRephoto: false, isAiImage: false })
    clearViews()
    refreshPhotoSpecs().catch(() => {})
    message.info('已重置')
  }

  async function runGenerate() {
    if (!originView.value?.img) {
      message.warning('请先打开或拍摄照片')
      return
    }
    processing.value = true
    progressPercent.value = 8
    statusText.value = '开始处理...'
    processTagType.value = 'warning'
    const progressTimer = setInterval(() => {
      if (progressPercent.value < 85) {
        progressPercent.value += 7
      }
    }, 280)
    try {
      const raw = await IDPhotoService.Generate({
        ...params,
        sourceImg: originView.value.img,
      })
      // Wails / HTTP / mock 统一在这里收尾，避免桌面端一直停在「处理中」
      onDone(raw)
    } catch (err) {
      onError({ msg: err?.message || '生成失败' })
    } finally {
      clearInterval(progressTimer)
    }
  }

  function onTemplateChange(value) {
    params.template = value
    if (value && value !== 'custom') {
      photoSpecMeta.current = value
      IDPhotoService.SetCurrentPhotoSpec(value).catch(() => {})
    }
  }

  function applyCustomSpec() {
    params.template = 'custom'
    IDPhotoService.SetCurrentPhotoSpec('').catch(() => {})
    message.success(`已应用自定义尺寸 ${params.customWidth}×${params.customHeight} mm`)
  }

  function selectSpecCategory(key) {
    activeSpecCategory.value = key
    specKeyword.value = ''
    if (key === 'custom') {
      params.template = 'custom'
    }
    refreshPhotoSpecs().catch(() => {})
  }

  function openExportModal() {
    showExportModal.value = true
  }

  function openPrintModal() {
    if (!resultSet.layout) {
      message.warning('请先生成排版照后再打印')
      return
    }
    printCopies.value = 1
    printPaperSize.value = params.paperSize || printPaperSize.value || '6inch'
    printLandscape.value = false
    showPrintModal.value = true
    refreshPrinters().catch(() => {})
  }

  async function refreshPrinters() {
    try {
      const ret = await IDPhotoService.ListPrinters()
      printers.value = Array.isArray(ret?.printers) ? ret.printers : []
      nativePrint.value = !!ret?.native
      if (!printPrinterName.value) {
        const def = printers.value.find((p) => p.isDefault) || printers.value[0]
        printPrinterName.value = def?.name || ''
      } else if (printers.value.length && !printers.value.some((p) => p.name === printPrinterName.value)) {
        const def = printers.value.find((p) => p.isDefault) || printers.value[0]
        printPrinterName.value = def?.name || ''
      }
    } catch {
      printers.value = []
      nativePrint.value = false
    }
  }

  function printLayoutImage(dataUrl, copies = 1) {
    const count = Math.max(1, Math.min(20, Number(copies) || 1))
    const iframe = document.createElement('iframe')
    iframe.setAttribute('aria-hidden', 'true')
    iframe.style.cssText = 'position:fixed;right:0;bottom:0;width:0;height:0;border:0;opacity:0;pointer-events:none'
    document.body.appendChild(iframe)

    const pages = Array.from({ length: count }, (_, i) =>
      `<div class="page"><img id="img-${i}" src="${dataUrl}" alt="排版照" /></div>`,
    ).join('')

    const doc = iframe.contentDocument || iframe.contentWindow?.document
    if (!doc) {
      iframe.remove()
      throw new Error('无法打开打印预览')
    }

    doc.open()
    doc.write(`<!doctype html><html><head><meta charset="utf-8" /><title>打印排版照</title>
<style>
  @page { margin: 8mm; }
  html, body { margin: 0; padding: 0; background: #fff; }
  .page { page-break-after: always; display: flex; align-items: center; justify-content: center; min-height: 100vh; }
  .page:last-child { page-break-after: auto; }
  img { max-width: 100%; max-height: 100vh; object-fit: contain; }
</style></head><body>${pages}</body></html>`)
    doc.close()

    return new Promise((resolve, reject) => {
      const imgs = Array.from(doc.images || [])
      const done = () => {
        try {
          iframe.contentWindow?.focus()
          iframe.contentWindow?.print()
          resolve()
        } catch (err) {
          reject(err)
        } finally {
          setTimeout(() => iframe.remove(), 800)
        }
      }
      if (!imgs.length) {
        done()
        return
      }
      let left = imgs.length
      const tick = () => {
        left -= 1
        if (left <= 0) done()
      }
      imgs.forEach((img) => {
        if (img.complete) tick()
        else {
          img.onload = tick
          img.onerror = tick
        }
      })
    })
  }

  async function doPrint() {
    if (!resultSet.layout) {
      message.warning('暂无排版照可打印')
      return
    }
    printBusy.value = true
    try {
      const canNative = nativePrint.value
      if (canNative) {
        if (!printPrinterName.value) {
          await refreshPrinters()
        }
        if (!printPrinterName.value) {
          throw new Error('未找到可用打印机')
        }
        await IDPhotoService.PrintLayout({
          imageDataUrl: resultSet.layout,
          printerName: printPrinterName.value,
          paperSize: printPaperSize.value || params.paperSize || '6inch',
          copies: printCopies.value,
          landscape: printLandscape.value,
        })
        showPrintModal.value = false
        statusText.value = '打印任务已提交'
        message.success(`已发送到打印机：${printPrinterName.value}`)
      } else {
        await printLayoutImage(resultSet.layout, printCopies.value)
        showPrintModal.value = false
        statusText.value = '已调起打印'
        message.success('已打开打印对话框')
      }
    } catch (e) {
      message.error(e?.message || '打印失败')
    } finally {
      printBusy.value = false
    }
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
    try {
      localStorage.setItem('idportrait.progressStyle', setting.progressStyle === 'ghost' ? 'ghost' : 'card')
    } catch {
      // ignore storage failures
    }
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
    hasMatting,
    activeResultTab,
    activeOriginTab,
    resultTabs,
    originTabs,
    originView,
    mattingView,
    resultView,
    resultSet,
    currentResultHint,
    currentOriginHint,
    activeSocial,
    socialView,
    socialHint,
    params,
    bgModes,
    clothOptions,
    paperSizes,
    faceDetectModels,
    mattingModels,
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
    showPrintModal,
    showSettingModal,
    showAboutModal,
    showUpgradeModal,
    showCameraModal,
    appVersion,
    checkingUpgrade,
    exportDir,
    exportOpt,
    printCopies,
    printPrinterName,
    printPaperSize,
    printLandscape,
    printBusy,
    printers,
    nativePrint,
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
    switchSocial,
    switchOriginTab,
    openExportModal,
    openPrintModal,
    refreshPrinters,
    doPrint,
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
