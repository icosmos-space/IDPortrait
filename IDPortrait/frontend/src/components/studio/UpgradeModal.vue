<script setup>
defineProps({
  show: { type: Boolean, default: false },
  appVersion: { type: String, default: '1.0.0' },
  checkingUpgrade: { type: Boolean, default: false },
})

const emit = defineEmits(['update:show', 'check'])
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    title="版本更新"
    style="width: 420px"
    :bordered="false"
    :segmented="{ content: true, footer: 'soft' }"
    :mask-closable="false"
    @update:show="(v) => emit('update:show', v)"
  >
    <div class="modal-stack">
      <p class="upgrade-text">
        当前版本为 <strong>v{{ appVersion }}</strong>。是否检查并升级到最新版本？
      </p>
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn ghost" type="button" @click="emit('update:show', false)">暂不升级</button>
        <button class="btn primary" type="button" :disabled="checkingUpgrade" @click="emit('check')">
          <span v-if="checkingUpgrade" class="spin" />
          {{ checkingUpgrade ? '检查中…' : '检查升级' }}
        </button>
      </div>
    </template>
  </n-modal>
</template>
