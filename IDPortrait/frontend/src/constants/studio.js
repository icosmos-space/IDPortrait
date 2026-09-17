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

export const PAPER_SIZES = [
  { value: '5inch', title: '5寸', desc: '89×127 mm' },
  { value: '6inch', title: '6寸', desc: '102×152 mm' },
  { value: '7inch', title: '7寸', desc: '127×178 mm' },
  { value: 'a6', title: 'A6', desc: '105×148 mm' },
  { value: 'a5', title: 'A5', desc: '148×210 mm' },
  { value: 'a4', title: 'A4', desc: '210×297 mm' },
]

export const SPEC_CATEGORIES = [
  { key: 'common', label: '常用' },
  { key: 'visa', label: '签职' },
  { key: 'id', label: '证件' },
  { key: 'school', label: '升学' },
  { key: 'exam', label: '考试' },
  { key: 'custom', label: '自定义' },
]

export const ALL_SPECS = [
  { value: 'one_inch', title: '一寸', desc: '25×35 mm', categories: ['common', 'id'], keywords: '一寸 常用' },
  { value: 'two_inch', title: '二寸', desc: '35×49 mm', categories: ['common', 'id'], keywords: '二寸 常用' },
  { value: 'small_two', title: '小二寸', desc: '33×48 mm', categories: ['common'], keywords: '小二寸' },
  { value: 'large_one', title: '大一寸', desc: '33×48 mm', categories: ['common'], keywords: '大一寸' },
  { value: 'passport', title: '护照', desc: '33×48 mm', categories: ['common', 'visa', 'id'], keywords: '护照 出国' },
  { value: 'visa_us', title: '美签', desc: '51×51 mm', categories: ['visa'], keywords: '美签 美国签证' },
  { value: 'visa_schengen', title: '申根签', desc: '35×45 mm', categories: ['visa'], keywords: '申根 欧洲签证' },
  { value: 'visa_jp', title: '日签', desc: '45×45 mm', categories: ['visa'], keywords: '日签 日本签证' },
  { value: 'id_card', title: '身份证', desc: '26×32 mm', categories: ['id'], keywords: '身份证' },
  { value: 'driver', title: '驾驶证', desc: '22×32 mm', categories: ['id'], keywords: '驾驶证 驾照' },
  { value: 'social', title: '社保卡', desc: '26×32 mm', categories: ['id'], keywords: '社保卡' },
  { value: 'teacher', title: '教资', desc: '25×35 mm · 教师资格', categories: ['exam', 'visa'], keywords: '教资 教师资格' },
  { value: 'civil', title: '公务员', desc: '25×35 mm', categories: ['exam', 'visa'], keywords: '公务员 国考' },
  { value: 'grad', title: '毕业证', desc: '33×48 mm', categories: ['school'], keywords: '毕业证 学历' },
  { value: 'student', title: '学生证', desc: '25×35 mm', categories: ['school'], keywords: '学生证' },
  { value: 'exam_cet', title: '四六级', desc: '144×192 px', categories: ['exam'], keywords: '英语 四六级 CET' },
  { value: 'exam_cs', title: '计算机等级', desc: '144×192 px', categories: ['exam'], keywords: '计算机等级 NCRE' },
  { value: 'exam_nurse', title: '护士资格', desc: '25×35 mm', categories: ['exam'], keywords: '护士资格证' },
]

export const ABOUT_INFO = {
  name: '最美证件照',
  version: APP_VERSION,
  author: '王新勇(Tacey Wong)',
  desc: '本地影楼级证件照精修与排版工具，支持多规格制证、打印纸排版与合规核验。',
}

export function createDefaultParams() {
  return {
    template: 'one_inch',
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
    paperSize: '6inch',
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
