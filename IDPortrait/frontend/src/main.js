import { createApp, h } from 'vue'
import naive, { NConfigProvider, NDialogProvider, NMessageProvider, zhCN, dateZhCN } from 'naive-ui'
import App from './App.vue'

const themeOverrides = {
  common: {
    primaryColor: '#c45b7a',
    primaryColorHover: '#d16b89',
    primaryColorPressed: '#a34460',
    primaryColorSuppl: '#c45b7a',
    successColor: '#3d8b63',
    warningColor: '#a67c2a',
    errorColor: '#c45b5b',
    borderRadius: '12px',
    fontFamily: '"Nunito", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif',
  },
  Button: {
    heightMedium: '34px',
    paddingMedium: '0 14px',
    fontWeight: '600',
    borderRadiusMedium: '12px',
  },
  Slider: {
    fillColor: '#c45b7a',
    fillColorHover: '#d16b89',
    handleSize: '16px',
  },
  Card: {
    borderRadius: '16px',
  },
  Collapse: {
    titleFontWeight: '600',
  },
}

const fill = {
  width: '100%',
  height: '100%',
  minHeight: 0,
  display: 'flex',
  flexDirection: 'column',
}

const app = createApp({
  render() {
    return h(
      NConfigProvider,
      {
        locale: zhCN,
        dateLocale: dateZhCN,
        themeOverrides,
        style: fill,
      },
      {
        default: () =>
          h(
            NMessageProvider,
            { style: fill },
            {
              default: () =>
                h(
                  NDialogProvider,
                  { style: fill },
                  { default: () => h(App) },
                ),
            },
          ),
      },
    )
  },
})

app.use(naive)
app.mount('#app')
