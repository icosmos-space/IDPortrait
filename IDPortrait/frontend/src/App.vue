<script setup>
import { useStudio } from './composables/useStudio'
import StudioTopbar from './components/studio/StudioTopbar.vue'
import StudioLeftRail from './components/studio/StudioLeftRail.vue'
import StudioPreview from './components/studio/StudioPreview.vue'
import StudioRightRail from './components/studio/StudioRightRail.vue'
import CameraModal from './components/studio/CameraModal.vue'
import ExportModal from './components/studio/ExportModal.vue'
import PrintModal from './components/studio/PrintModal.vue'
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
  hasMatting,
  activeResultTab,
  activeOriginTab,
  resultTabs,
  originTabs,
  originView,
  mattingView,
  resultView,
  resultSet,
  currentResultHint,
  currentOriginHint,
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
  showPrintModal,
  showSettingModal,
  showAboutModal,
  showUpgradeModal,
  showCameraModal,
  appVersion,
  checkingUpgrade,
  exportDir,
  exportOpt,
  printCopies,
  printPrinterName,
  printPaperSize,
  printLandscape,
  printBusy,
  printers,
  nativePrint,
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
  switchOriginTab,
  openExportModal,
  openPrintModal,
  refreshPrinters,
  doPrint,
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
      @open-image="openImage"
      @open-camera="openCameraModal"
      @reset="resetAll"
      @generate="runGenerate"
      @export="openExportModal"
      @print="openPrintModal"
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
      :status-text="statusText"
      :status-tone="statusTone"
      :progress-percent="progressPercent"
      @select-spec-category="selectSpecCategory"
      @template-change="onTemplateChange"
      @apply-custom-spec="applyCustomSpec"
      @open-about="openAboutModal"
      @open-setting="openSettingModal"
      @open-upgrade="openUpgradeModal"
    />

    <StudioPreview
      :origin-view="originView"
      :matting-view="mattingView"
      :result-view="resultView"
      :has-origin="hasOrigin"
      :has-matting="hasMatting"
      :has-result="hasResult"
      :active-origin-tab="activeOriginTab"
      :origin-tabs="originTabs"
      :active-result-tab="activeResultTab"
      :result-tabs="resultTabs"
      :current-origin-hint="currentOriginHint"
      :current-result-hint="currentResultHint"
      :diagnostics="diagnostics"
      :beauty-strength="params.beautyStrength"
      @drop="handleDrop"
      @switch-origin-tab="switchOriginTab"
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

    <PrintModal
      v-model:show="showPrintModal"
      v-model:copies="printCopies"
      v-model:printer-name="printPrinterName"
      v-model:paper-size="printPaperSize"
      v-model:landscape="printLandscape"
      :layout-img="resultSet.layout"
      :printers="printers"
      :paper-sizes="paperSizes"
      :printing="printBusy"
      :native-print="nativePrint"
      @print="doPrint"
      @refresh-printers="refreshPrinters"
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
