<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  layoutImg: { type: String, default: '' },
  paperLabel: { type: String, default: '6寸' },
  copies: { type: Number, default: 1 },
})

const emit = defineEmits(['update:show', 'update:copies', 'print'])

const hasLayout = computed(() => Boolean(props.layoutImg))
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
      <p class="modal-hint">将打开系统打印对话框，可选择打印机与纸张设置。</p>
      <div class="print-preview">
        <img v-if="hasLayout" :src="layoutImg" alt="排版照预览" />
        <p v-else class="print-empty">暂无排版照，请先生成后再打印</p>
      </div>
      <div class="print-meta">
        <label class="modal-label">纸张规格</label>
        <n-input :value="paperLabel" readonly />
      </div>
      <div class="print-meta">
        <label class="modal-label">打印份数</label>
        <n-input-number
          :value="copies"
          :min="1"
          :max="20"
          size="small"
          style="width: 120px"
          @update:value="(v) => emit('update:copies', v || 1)"
        />
      </div>
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn ghost" type="button" @click="emit('update:show', false)">取消</button>
        <button class="btn primary" type="button" :disabled="!hasLayout" @click="emit('print')">
          选择打印机并打印
        </button>
      </div>
    </template>
  </n-modal>
</template>
