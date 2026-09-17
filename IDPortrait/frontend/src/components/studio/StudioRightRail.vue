<script setup>
import { CLOTH_OPTIONS, formatVal } from '../../constants/studio'

defineProps({
  params: { type: Object, required: true },
  expandedPanels: { type: Array, required: true },
  clothOptions: { type: Array, default: () => CLOTH_OPTIONS },
  paperSizes: { type: Array, default: () => [] },
  faceDetectModels: { type: Array, default: () => [] },
  mattingModels: { type: Array, default: () => [] },
  currentPaperLabel: { type: String, default: '纸张' },
})

const emit = defineEmits(['update:expandedPanels'])
</script>

<template>
  <aside class="rail right">
    <div class="section-label">精修参数</div>
    <n-collapse
      class="right-collapse"
      :expanded-names="expandedPanels"
      accordion
      display-directive="show"
      :trigger-areas="['main', 'arrow']"
      @update:expanded-names="(v) => emit('update:expandedPanels', v)"
    >
      <n-collapse-item title="基础美颜" name="beauty">
        <div class="field">
          <div class="field-row">
            <label>美颜强度</label>
            <span>{{ formatVal(params.beautyStrength) }}</span>
          </div>
          <n-slider v-model:value="params.beautyStrength" :max="0.4" :step="0.01" />
        </div>
        <div class="field">
          <div class="field-row">
            <label>眼睛锐化</label>
            <span>{{ formatVal(params.eyeSharp) }}</span>
          </div>
          <n-slider v-model:value="params.eyeSharp" :max="0.3" :step="0.01" />
        </div>
        <div class="field">
          <div class="field-row">
            <label>肤色提亮</label>
            <span>{{ formatVal(params.skinBright) }}</span>
          </div>
          <n-slider v-model:value="params.skinBright" :max="0.3" :step="0.01" />
        </div>
      </n-collapse-item>

      <n-collapse-item title="自动化妆" name="makeup">
        <div class="field switch-row">
          <label>自然妆感</label>
          <n-switch v-model:value="params.enableMakeup" size="small" />
        </div>
        <div class="field">
          <div class="field-row">
            <label>唇色增强</label>
            <span>{{ formatVal(params.lipStrength) }}</span>
          </div>
          <n-slider
            v-model:value="params.lipStrength"
            :max="0.25"
            :step="0.01"
            :disabled="!params.enableMakeup"
          />
        </div>
      </n-collapse-item>

      <n-collapse-item title="眉毛与头发" name="hair">
        <div class="field switch-row">
          <label>眉毛自然增强</label>
          <n-switch v-model:value="params.refineBrow" size="small" />
        </div>
        <div class="field">
          <div class="field-row">
            <label>眉尾补全</label>
            <span>{{ formatVal(params.browFill) }}</span>
          </div>
          <n-slider
            v-model:value="params.browFill"
            :max="0.3"
            :step="0.01"
            :disabled="!params.refineBrow"
          />
        </div>
        <div class="field switch-row">
          <label>碎发整理</label>
          <n-switch v-model:value="params.refineHair" size="small" />
        </div>
        <div class="field">
          <div class="field-row">
            <label>发色统一</label>
            <span>{{ formatVal(params.hairColorUniform) }}</span>
          </div>
          <n-slider
            v-model:value="params.hairColorUniform"
            :max="0.35"
            :step="0.01"
            :disabled="!params.refineHair"
          />
        </div>
        <div class="field">
          <div class="field-row">
            <label>发际线修补</label>
            <span>{{ formatVal(params.hairlineRepair) }}</span>
          </div>
          <n-slider
            v-model:value="params.hairlineRepair"
            :max="0.3"
            :step="0.01"
            :disabled="!params.refineHair"
          />
        </div>
      </n-collapse-item>

      <n-collapse-item title="智能正装" name="cloth">
        <div class="field switch-row">
          <label>智能换正装</label>
          <n-switch v-model:value="params.enableCloth" size="small" />
        </div>
        <div class="field">
          <n-select
            v-model:value="params.clothType"
            :options="clothOptions"
            :disabled="!params.enableCloth"
            size="small"
          />
        </div>
        <div class="field">
          <div class="field-row">
            <label>服装贴合度</label>
            <span>{{ formatVal(params.clothFit) }}</span>
          </div>
          <n-slider
            v-model:value="params.clothFit"
            :max="0.4"
            :step="0.01"
            :disabled="!params.enableCloth"
          />
        </div>
      </n-collapse-item>

      <n-collapse-item title="水印" name="watermark">
        <div class="field switch-row">
          <label>开启水印</label>
          <n-switch v-model:value="params.enableWatermark" size="small" />
        </div>
        <template v-if="params.enableWatermark">
          <div class="field">
            <label>水印文字</label>
            <n-input v-model:value="params.watermarkText" size="small" placeholder="请输入水印文字" />
          </div>
          <div class="field">
            <label>水印颜色</label>
            <div class="field-row">
              <n-color-picker v-model:value="params.watermarkColor" :show-alpha="false" size="small" style="width: 100%" />
            </div>
          </div>
          <div class="field">
            <div class="field-row">
              <label>文字大小</label>
              <span>{{ Math.round(params.watermarkFontSize) }}</span>
            </div>
            <n-slider v-model:value="params.watermarkFontSize" :min="8" :max="72" :step="1" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>水印透明度</label>
              <span>{{ formatVal(params.watermarkOpacity) }}</span>
            </div>
            <n-slider v-model:value="params.watermarkOpacity" :min="0.05" :max="1" :step="0.01" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>水印角度</label>
              <span>{{ Math.round(params.watermarkAngle) }}°</span>
            </div>
            <n-slider v-model:value="params.watermarkAngle" :min="-90" :max="90" :step="1" />
          </div>
          <div class="field">
            <div class="field-row">
              <label>水印间距</label>
              <span>{{ Math.round(params.watermarkSpacing) }}</span>
            </div>
            <n-slider v-model:value="params.watermarkSpacing" :min="40" :max="400" :step="4" />
          </div>
        </template>
      </n-collapse-item>

      <n-collapse-item title="高级选项" name="adv">
        <div class="field">
          <label>人脸检测模型</label>
          <n-select
            v-model:value="params.faceDetectModel"
            :options="faceDetectModels.map((m) => ({ label: `${m.title}（${m.desc}）`, value: m.value }))"
            size="small"
            placeholder="选择人脸检测模型"
          />
        </div>
        <div class="field">
          <label>抠图模型</label>
          <n-select
            v-model:value="params.mattingModel"
            :options="mattingModels.map((m) => ({ label: `${m.title}（${m.desc}）`, value: m.value }))"
            size="small"
            placeholder="选择抠图模型"
          />
        </div>
        <div class="field">
          <div class="field-row">
            <label>面部比例</label>
            <span>{{ formatVal(params.faceRatio) }}</span>
          </div>
          <n-slider v-model:value="params.faceRatio" :min="0.45" :max="0.85" :step="0.01" />
        </div>
        <div class="field">
          <div class="field-row">
            <label>头距顶距离</label>
            <span>{{ formatVal(params.headTopDistance) }}</span>
          </div>
          <n-slider v-model:value="params.headTopDistance" :min="0.02" :max="0.2" :step="0.01" />
        </div>
        <div class="field switch-row">
          <label>6 寸打印排版</label>
          <n-switch v-model:value="params.genPrintLayout" size="small" />
        </div>
        <div v-if="params.genPrintLayout" class="field">
          <label>当前纸张：{{ currentPaperLabel }}</label>
          <n-select
            v-model:value="params.paperSize"
            :options="paperSizes.map((p) => ({ label: `${p.title}（${p.desc}）`, value: p.value }))"
            size="small"
          />
        </div>
        <div class="field switch-row">
          <label>目标文件大小</label>
          <n-switch v-model:value="params.enableTargetFileSize" size="small" />
        </div>
        <div v-if="params.enableTargetFileSize" class="field">
          <label>目标大小 KB</label>
          <n-input-number
            v-model:value="params.targetFileSize"
            :min="50"
            :max="2000"
            size="small"
            style="width: 100%"
          />
        </div>
        <div class="field">
          <div class="field-row">
            <label>边缘羽化</label>
            <span>{{ formatVal(params.maskFeather) }}</span>
          </div>
          <n-slider v-model:value="params.maskFeather" :max="0.6" :step="0.01" />
        </div>
      </n-collapse-item>
    </n-collapse>
  </aside>
</template>
