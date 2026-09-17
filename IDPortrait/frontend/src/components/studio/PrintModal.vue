<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  layoutImg: { type: String, default: '' },
  printers: { type: Array, default: () => [] },
  paperSizes: { type: Array, default: () => [] },
  printerName: { type: String, default: '' },
  paperSize: { type: String, default: '6inch' },
  copies: { type: Number, default: 1 },
  landscape: { type: Boolean, default: false },
  printing: { type: Boolean, default: false },
  nativePrint: { type: Boolean, default: true },
})

const emit = defineEmits([
  'update:show',
  'update:printerName',
  'update:paperSize',
  'update:copies',
  'update:landscape',
  'print',
  'refresh-printers',
])

const hasLayout = computed(() => Boolean(props.layoutImg))

const printerOptions = computed(() =>
  (props.printers || []).map((p) => ({
    label: p.isDefault ? `${p.name}（默认）` : p.name,
    value: p.name,
  })),
)

const paperOptions = computed(() =>
  (props.paperSizes || []).map((p) => ({
    label: p.desc ? `${p.title} · ${p.desc}` : p.title,
    value: p.value,
  })),
)
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    title="打印排版照"
    style="width: 560px"
    :bordered="false"
    :segmented="{ content: true, footer: 'soft' }"
    :mask-closable="false"
    @update:show="(v) => emit('update:show', v)"
  >
    <div class="modal-stack">
      <p class="modal-hint">
        {{
          nativePrint
            ? '打印机、纸张与份数将直接应用于本机打印任务。'
            : '当前为浏览器模式，将打开系统打印对话框；纸张与份数可能需在对话框中再确认。'
        }}
      </p>
      <div class="print-preview">
        <img v-if="hasLayout" :src="layoutImg" alt="排版照预览" />
        <p v-else class="print-empty">暂无排版照，请先生成后再打印</p>
      </div>

      <div class="print-meta">
        <label class="modal-label">打印机</label>
        <div class="print-row">
          <n-select
            :value="printerName"
            :options="printerOptions"
            :disabled="!nativePrint || !printerOptions.length"
            placeholder="选择打印机"
            style="flex: 1"
            @update:value="(v) => emit('update:printerName', v)"
          />
          <button
            v-if="nativePrint"
            class="btn ghost"
            type="button"
            :disabled="printing"
            @click="emit('refresh-printers')"
          >
            刷新
          </button>
        </div>
      </div>

      <div class="print-meta">
        <label class="modal-label">纸张规格</label>
        <n-select
          :value="paperSize"
          :options="paperOptions"
          placeholder="选择纸张"
          @update:value="(v) => emit('update:paperSize', v)"
        />
      </div>

      <div class="print-meta print-inline">
        <div>
          <label class="modal-label">打印份数</label>
          <n-input-number
            :value="copies"
            :min="1"
            :max="99"
            size="small"
            style="width: 120px"
            @update:value="(v) => emit('update:copies', v || 1)"
          />
        </div>
        <div>
          <label class="modal-label">横向打印</label>
          <n-switch
            :value="landscape"
            @update:value="(v) => emit('update:landscape', v)"
          />
        </div>
      </div>
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn ghost" type="button" :disabled="printing" @click="emit('update:show', false)">
          取消
        </button>
        <button class="btn primary" type="button" :disabled="!hasLayout || printing" @click="emit('print')">
          <span v-if="printing" class="spin" />
          {{ printing ? '打印中…' : '开始打印' }}
        </button>
      </div>
    </template>
  </n-modal>
</template>
