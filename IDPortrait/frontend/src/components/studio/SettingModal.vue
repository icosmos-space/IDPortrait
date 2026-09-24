<script setup>
import { computed, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'

const props = defineProps({
  show: { type: Boolean, default: false },
  setting: { type: Object, required: true },
  remoteStatus: {
    type: Object,
    default: () => ({ running: false, port: 8787, url: '', addr: '' }),
  },
  remoteBusy: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:show',
  'save',
  'start-remote',
  'stop-remote',
  'refresh-remote',
])

const message = useMessage()
const localPort = computed({
  get: () => props.setting.remotePort || 8787,
  set: (v) => {
    props.setting.remotePort = Number(v) || 8787
  },
})

const remoteUrl = computed(() => props.remoteStatus?.url || props.remoteStatus?.addr || '')
const isRemoteRunning = computed(() => !!props.remoteStatus?.running)
const precheckExpanded = ref([])

watch(
  () => props.show,
  (v) => {
    if (v) emit('refresh-remote')
  },
)

async function copyUrl() {
  if (!remoteUrl.value) return
  try {
    await navigator.clipboard.writeText(remoteUrl.value)
    message.success('已复制访问地址')
  } catch {
    message.warning('复制失败，请手动选择地址')
  }
}

function onRemoteSwitch(v) {
  props.setting.remoteEnabled = v
  if (v) {
    emit('start-remote')
  } else {
    emit('stop-remote')
  }
}
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    title="软件设置"
    style="width: 520px"
    :bordered="false"
    :segmented="{ content: true, footer: 'soft' }"
    :mask-closable="false"
    @update:show="(v) => emit('update:show', v)"
  >
    <div class="modal-stack">
      <label class="modal-label">模型目录</label>
      <n-input v-model:value="setting.modelDir" placeholder="选择或填写模型目录" />
      <label class="modal-label">缓存目录</label>
      <n-input v-model:value="setting.cacheDir" placeholder="选择或填写缓存目录" />
      <div class="field switch-row">
        <label>屏幕翻拍检测</label>
        <n-switch v-model:value="setting.enableRephotoDetect" size="small" />
      </div>
      <div class="field switch-row">
        <label>AI 图片预检</label>
        <n-switch v-model:value="setting.enableAiDetect" size="small" />
      </div>

      <n-collapse
        class="precheck-collapse"
        display-directive="show"
        :expanded-names="precheckExpanded"
        @update:expanded-names="(names) => { precheckExpanded = names }"
      >
        <n-collapse-item title="导入人脸预检" name="precheck">
          <p class="precheck-desc">关闭后仍会检测「是否有且仅有一张人脸」；下列开关可分别关闭对应拒绝项。</p>
          <div class="field switch-row">
            <label>侧脸 / 俯仰</label>
            <n-switch v-model:value="setting.checkPose" size="small" />
          </div>
          <div class="field switch-row">
            <label>头肩倾斜 / 扭转</label>
            <n-switch v-model:value="setting.checkTilt" size="small" />
          </div>
          <div class="field switch-row">
            <label>人脸模糊</label>
            <n-switch v-model:value="setting.checkBlur" size="small" />
          </div>
          <div class="field switch-row">
            <label>人脸马赛克</label>
            <n-switch v-model:value="setting.checkMosaic" size="small" />
          </div>
          <div class="field switch-row">
            <label>闭眼 / 遮挡</label>
            <n-switch v-model:value="setting.checkParse" size="small" />
          </div>
        </n-collapse-item>
      </n-collapse>

      <div class="field switch-row progress-style-row">
        <label>进度样式</label>
        <n-radio-group v-model:value="setting.progressStyle" name="progress-style">
          <n-space :size="16" :wrap="false">
            <n-radio value="card">白底卡片</n-radio>
            <n-radio value="ghost">透明浮层</n-radio>
          </n-space>
        </n-radio-group>
      </div>

      <div class="remote-block">
        <div class="remote-title">远程服务</div>
        <p class="remote-desc">
          开启后以 HTTPS 提供服务（摄像头需要安全连接）。其他设备用下方地址访问；首次会提示证书不受信任，选择继续访问即可。
        </p>
        <div class="field switch-row">
          <label>启动远程服务</label>
          <n-switch
            :value="setting.remoteEnabled"
            :disabled="remoteBusy"
            size="small"
            @update:value="onRemoteSwitch"
          />
        </div>
        <label class="modal-label">监听端口</label>
        <n-input-number
          v-model:value="localPort"
          :min="1024"
          :max="65535"
          :disabled="isRemoteRunning || remoteBusy"
          style="width: 100%"
        />
        <div v-if="isRemoteRunning && remoteUrl" class="remote-url-box">
          <div class="remote-url-label">访问地址</div>
          <div class="remote-url-row">
            <code class="remote-url">{{ remoteUrl }}</code>
            <button class="btn ghost" type="button" @click="copyUrl">复制</button>
          </div>
        </div>
        <div class="remote-status" :class="{ on: isRemoteRunning }">
          {{ isRemoteRunning ? '服务运行中' : '服务未启动' }}
        </div>
      </div>
    </div>
    <template #footer>
      <div class="modal-actions">
        <button class="btn primary" type="button" @click="emit('save')">保存设置</button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.precheck-collapse {
  margin-top: 4px;
  padding-top: 8px;
  border-top: 1px solid rgba(148, 163, 184, 0.35);
}
.precheck-desc {
  margin: 0 0 10px;
  font-size: 12px;
  color: #64748b;
  line-height: 1.5;
}
.progress-style-row :deep(.n-radio-group) {
  display: inline-flex;
  flex-wrap: nowrap;
  align-items: center;
}
.remote-block {
  margin-top: 8px;
  padding-top: 14px;
  border-top: 1px solid rgba(148, 163, 184, 0.35);
}
.remote-title {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 4px;
}
.remote-desc {
  margin: 0 0 12px;
  font-size: 12px;
  color: #64748b;
  line-height: 1.5;
}
.remote-url-box {
  margin-top: 10px;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 8px;
}
.remote-url-label {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 6px;
}
.remote-url-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.remote-url {
  flex: 1;
  font-size: 13px;
  color: #0f766e;
  word-break: break-all;
}
.remote-status {
  margin-top: 10px;
  font-size: 12px;
  color: #94a3b8;
}
.remote-status.on {
  color: #0f766e;
}
</style>
