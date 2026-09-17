export const APP_VERSION = '1.0.0'

export const RESULT_TABS = [
  { key: 'single', label: '单张照片', hint: '精修单图' },
  { key: 'layout', label: '排版照', hint: '6寸打印排版' },
  { key: 'social', label: '社交照', hint: '方形社交尺寸' },
  { key: 'idphoto', label: '证件照', hint: '标准证件规格' },
]

export const BG_MODES = [
  { value: 'solid', label: '纯色' },
  { value: 'vertical', label: '上下渐变' },
  { value: 'radial', label: '中心渐变' },
]

export const BG_PRESETS = [
  { value: '#FFFFFF', label: '白' },
  { value: '#D92121', label: '红' },
  { value: '#0047AB', label: '蓝' },
]

export const CLOTH_OPTIONS = [
  { label: '白衬衫', value: 'white_shirt' },
  { label: '深色西装', value: 'dark_suit' },
  { label: '学士服', value: 'grad' },
  { label: '商务职业照', value: 'business' },
]

export const ABOUT_INFO = {
  name: '最美证件照',
  version: APP_VERSION,
  author: '王新勇(Tacey Wong)',
  desc: '本地影楼级证件照精修与排版工具，支持多规格制证、打印纸排版与合规核验。',
}

export function createDefaultParams() {
  return {
    template: '',
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
    paperSize: '',
    enableTargetFileSize: false,
    targetFileSize: 200,
    maskFeather: 0.3,
  }
}

export function modePreviewStyle(mode, bgColor = '#FFFFFF') {
  const a = bgColor || '#FFFFFF'
  const b = '#FFFFFF'
  if (mode === 'vertical') return { background: `linear-gradient(180deg, ${a} 0%, ${b} 100%)` }
  if (mode === 'radial') return { background: `radial-gradient(circle at 50% 42%, ${a} 0%, ${b} 100%)` }
  return { background: a }
}

export function formatVal(v) {
  return Number(v).toFixed(2)
}
