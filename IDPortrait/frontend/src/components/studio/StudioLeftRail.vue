<script setup>
import { BG_MODES, BG_PRESETS } from '../../constants/studio'

defineProps({
  params: { type: Object, required: true },
  bgPresets: { type: Array, default: () => BG_PRESETS },
  bgModes: { type: Array, default: () => BG_MODES },
  specCategories: { type: Array, default: () => [] },
  specKeyword: { type: String, default: '' },
  activeSpecCategory: { type: String, default: 'common' },
  filteredSpecs: { type: Array, default: () => [] },
  showCustomSpec: { type: Boolean, default: false },
  appVersion: { type: String, default: '1.0.0' },
  previewModeStyle: { type: Function, required: true },
})

const emit = defineEmits([
  'update:specKeyword',
  'select-spec-category',
  'template-change',
  'apply-custom-spec',
  'open-about',
  'open-setting',
  'open-upgrade',
])
</script>

<template>
  <aside class="rail left">
    <section class="rail-section pin-top">
      <div class="section-label">背景颜色</div>
      <div class="swatches">
        <button
          v-for="c in bgPresets"
          :key="c.value"
          type="button"
          class="swatch"
          :class="{ active: params.bgPreset === c.value }"
          :style="{ '--swatch': c.value }"
          @click="params.bgPreset = c.value"
        >
          {{ c.label }}
        </button>
      </div>
      <n-color-picker v-model:value="params.bgColor" :modes="['hex']" size="small" style="margin-top: 10px" />
      <div class="bg-modes" style="margin-top: 10px">
        <button
          v-for="m in bgModes"
          :key="m.value"
          type="button"
          class="bg-mode"
          :class="{ active: params.bgMode === m.value }"
          @click="params.bgMode = m.value"
        >
          <span class="bg-mode-preview" :style="previewModeStyle(m.value)" />
          {{ m.label }}
        </button>
      </div>
    </section>

    <section class="rail-section grow specs-section">
      <div class="section-label">证件规格</div>
      <n-input
        :value="specKeyword"
        size="small"
        clearable
        placeholder="搜索规格，如护照、教资"
        class="spec-search"
        @update:value="(v) => emit('update:specKeyword', v)"
      />
      <div class="spec-cats" role="tablist">
        <button
          v-for="cat in specCategories"
          :key="cat.key"
          type="button"
          class="spec-cat"
          :class="{ active: activeSpecCategory === cat.key && !specKeyword.trim() }"
          @click="emit('select-spec-category', cat.key)"
        >
          {{ cat.label }}
        </button>
      </div>

      <div class="spec-scroll">
        <div v-if="showCustomSpec" class="custom-spec">
          <div class="field">
            <label>宽度 mm</label>
            <n-input-number v-model:value="params.customWidth" :min="10" :max="100" size="small" style="width: 100%" />
          </div>
          <div class="field">
            <label>高度 mm</label>
            <n-input-number v-model:value="params.customHeight" :min="10" :max="140" size="small" style="width: 100%" />
          </div>
          <button class="btn primary custom-apply" type="button" @click="emit('apply-custom-spec')">应用自定义</button>
        </div>

        <div v-else class="template-grid">
          <button
            v-for="t in filteredSpecs"
            :key="t.value"
            type="button"
            class="template-card"
            :class="{ active: params.template === t.value }"
            @click="emit('template-change', t.value)"
          >
            <strong>{{ t.title }}</strong>
            <span>{{ t.desc }}</span>
          </button>
          <p v-if="!filteredSpecs.length" class="spec-empty">未找到匹配规格</p>
        </div>
      </div>
    </section>

    <section class="rail-section pin-footer">
      <button class="footer-link" type="button" @click="emit('open-about')">关于</button>
      <button class="footer-link" type="button" @click="emit('open-setting')">设置</button>
      <button class="footer-version" type="button" @click="emit('open-upgrade')" title="检查更新">
        v{{ appVersion }}
      </button>
    </section>
  </aside>
</template>
