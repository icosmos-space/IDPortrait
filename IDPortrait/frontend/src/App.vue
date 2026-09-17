<script setup>
import { useStudio } from './composables/useStudio'
import StudioTopbar from './components/studio/StudioTopbar.vue'
import StudioLeftRail from './components/studio/StudioLeftRail.vue'
import StudioPreview from './components/studio/StudioPreview.vue'
import StudioRightRail from './components/studio/StudioRightRail.vue'
import CameraModal from './components/studio/CameraModal.vue'
import ExportModal from './components/studio/ExportModal.vue'
import SettingModal from './components/studio/SettingModal.vue'
import AboutModal from './components/studio/AboutModal.vue'
import UpgradeModal from './components/studio/UpgradeModal.vue'
import './styles/studio.css'

const {
  expandedPanels,
  processing,
  progressPercent,
  statusText,
  hasOrigin,
  hasResult,
  activeResultTab,
  resultTabs,
  originView,
  resultView,
  currentResultHint,
  params,
  bgModes,
  clothOptions,
  paperSizes,
  faceDetectModels,
  mattingModels,
  currentPaperLabel,
  specCategories,
  bgPresets,
  specKeyword,
  activeSpecCategory,
  filteredSpecs,
  showCustomSpec,
  showExportModal,
  showSettingModal,
  showAboutModal,
  showUpgradeModal,
  showCameraModal,
  appVersion,
  checkingUpgrade,
  exportDir,
  exportOpt,
  setting,
  remoteStatus,
  remoteBusy,
  aboutInfo,
  statusTone,
  diagnostics,
  previewModeStyle,
  openImage,
  openCameraModal,
  applyCapture,
  handleDrop,
  resetAll,
  runGenerate,
  onTemplateChange,
  applyCustomSpec,
  selectSpecCategory,
  switchResultTab,
  openExportModal,
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
} = useStudio()
</script>

<template>
  <div class="studio">
    <StudioTopbar
      :processing="processing"
      :status-text="statusText"
      :status-tone="statusTone"
      :progress-percent="progressPercent"
      @open-image="openImage"
      @open-camera="openCameraModal"
      @reset="resetAll"
      @generate="runGenerate"
      @export="openExportModal"
    />

    <StudioLeftRail
      :params="params"
      :bg-presets="bgPresets"
      :bg-modes="bgModes"
      :spec-categories="specCategories"
      v-model:spec-keyword="specKeyword"
      :active-spec-category="activeSpecCategory"
      :filtered-specs="filteredSpecs"
      :show-custom-spec="showCustomSpec"
      :app-version="appVersion"
      :preview-mode-style="previewModeStyle"
      @select-spec-category="selectSpecCategory"
      @template-change="onTemplateChange"
      @apply-custom-spec="applyCustomSpec"
      @open-about="openAboutModal"
      @open-setting="openSettingModal"
      @open-upgrade="openUpgradeModal"
    />

    <StudioPreview
      :origin-view="originView"
      :result-view="resultView"
      :has-origin="hasOrigin"
      :has-result="hasResult"
      :active-result-tab="activeResultTab"
      :result-tabs="resultTabs"
      :current-result-hint="currentResultHint"
      :diagnostics="diagnostics"
      :beauty-strength="params.beautyStrength"
      @drop="handleDrop"
      @switch-result-tab="switchResultTab"
    />

    <StudioRightRail
      v-model:expanded-panels="expandedPanels"
      :params="params"
      :cloth-options="clothOptions"
      :paper-sizes="paperSizes"
      :face-detect-models="faceDetectModels"
      :matting-models="mattingModels"
      :current-paper-label="currentPaperLabel"
    />

    <CameraModal v-model:show="showCameraModal" @capture="applyCapture" />

    <ExportModal
      v-model:show="showExportModal"
      :export-opt="exportOpt"
      :export-dir="exportDir"
      @select-dir="selectExportDir"
      @export="doExport"
    />

    <SettingModal
      v-model:show="showSettingModal"
      :setting="setting"
      :remote-status="remoteStatus"
      :remote-busy="remoteBusy"
      @save="saveSetting"
      @start-remote="startRemote"
      @stop-remote="stopRemote"
      @refresh-remote="refreshRemoteStatus"
    />

    <AboutModal v-model:show="showAboutModal" :about-info="aboutInfo" />

    <UpgradeModal
      v-model:show="showUpgradeModal"
      :app-version="appVersion"
      :checking-upgrade="checkingUpgrade"
      @check="checkAndUpgrade"
    />
  </div>
</template>
