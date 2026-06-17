<template>
  <div class="quest-editor-root">
    <div v-if="!appEnvLocal" class="quest-editor-blocked">
      <div>
        <i class="fa fa-lock" style="font-size: 40px; opacity: 0.35"></i>
        <p class="mt-3 mb-0">Quest Editor is only available for local Spire installs.</p>
      </div>
    </div>
    <eq-window-simple
      v-else
      title="Quest Editor"
      class="p-0 quest-editor-window"
      classes="quest-editor-window-content"
    >
      <div class="quest-editor-toolbar d-flex align-items-center px-2 py-1">
        <b-button
          size="sm"
          variant="outline-warning"
          :disabled="!activeTab"
          @click="saveActiveFile"
        >
          <i class="fa fa-save"></i> Save
        </b-button>
        <b-button
          size="sm"
          variant="outline-warning"
          :disabled="!hasDirtyTabs"
          @click="saveAllFiles"
          class="ml-1"
        >
          <i class="fa fa-save"></i> Save All
        </b-button>
        <b-button
          size="sm"
          variant="outline-info"
          :disabled="!activeTab || !capabilities.format_available"
          @click="formatActiveFile"
          class="ml-1"
          :title="capabilities.format_available ? 'Format with stylua' : 'stylua not found on PATH'"
        >
          <i class="fa fa-magic"></i> Format
        </b-button>
        <b-button
          size="sm"
          variant="outline-info"
          :disabled="!activeTab || !capabilities.validate_available"
          @click="validateActiveFile"
          class="ml-1"
          :title="capabilities.validate_available ? 'Validate with luac -p' : 'luac not found on PATH'"
        >
          <i class="fa fa-check-circle"></i> Validate
        </b-button>
        <b-button
          size="sm"
          variant="outline-success"
          :disabled="!activeTab"
          @click="reloadActiveFile"
          class="ml-1"
          title="Reload quests on server"
        >
          <i class="fa fa-refresh"></i> Reload
        </b-button>

        <div class="ml-auto d-flex align-items-center">
          <b-form-checkbox
            size="sm"
            v-model="formatOnSave"
            class="mr-3"
          >
            <small>Format on Save</small>
          </b-form-checkbox>
          <span v-if="activeTab" class="text-muted" style="font-size:11px">
            {{ activeTab }}
          </span>
        </div>
      </div>

      <div class="quest-editor-body">
        <div class="quest-pane quest-pane-left">
          <quest-file-tree
            ref="fileTree"
            :active-file-path="activeTab"
            @file-select="openFile"
            @create-file="promptCreateFile"
            @create-folder="promptCreateFolder"
            @context-menu="handleContextMenu"
            @tree-loaded="onTreeLoaded"
          />
        </div>

        <div class="quest-pane quest-pane-center">
          <quest-editor-tabs
            ref="editorTabs"
            :tabs="openTabs"
            :active-tab="activeTab"
            :font-size="fontSize"
            @switch-tab="switchTab"
            @close-tab="closeTab"
            @content-change="onContentChange"
            @insert-snippet="insertSnippet"
          />
        </div>

        <div class="quest-pane quest-pane-right">
          <quest-helper-panel
            ref="helperPanel"
            @insert-snippet="insertSnippet"
          />
        </div>
      </div>
    </eq-window-simple>

    <b-modal
      id="create-file-modal"
      :title="createModalTitle"
      @ok="doCreateFile"
      :ok-disabled="!createModalPath"
    >
      <b-form-input
        v-model="createModalPath"
        placeholder="path/filename.lua"
      />
      <div class="mt-2 text-muted" style="font-size:12px" v-if="createModalHint">
        {{ createModalHint }}
      </div>
    </b-modal>

    <b-modal
      id="create-folder-modal"
      title="Create Folder"
      @ok="doCreateFolder"
      :ok-disabled="!createModalPath"
    >
      <b-form-input
        v-model="createModalPath"
        placeholder="path/to/folder"
      />
    </b-modal>

    <b-modal
      id="rename-modal"
      title="Rename / Move"
      @ok="doRename"
      :ok-disabled="!renameNewPath"
    >
      <b-form-group label="New path">
        <b-form-input v-model="renameNewPath" />
      </b-form-group>
    </b-modal>

    <b-modal
      id="delete-confirm-modal"
      title="Confirm Delete"
      @ok="doDelete"
      ok-variant="danger"
    >
      <p>Are you sure you want to delete <strong>{{ deleteTargetPath }}</strong>?</p>
      <p class="text-danger" v-if="deleteTargetIsDir">This will delete the folder and all its contents recursively.</p>
    </b-modal>
  </div>
</template>

<script type="ts">
import EqWindowSimple from '@/components/eq-ui/EQWindowSimple.vue'
import QuestFileTree from './components/QuestFileTree.vue'
import QuestEditorTabs from './components/QuestEditorTabs.vue'
import QuestHelperPanel from './components/QuestHelperPanel.vue'
import {QuestEditorApi} from '@/app/api/quest-editor-api'
import {QuestTemplates} from './quest-templates'
import {LocalSettings} from '@/app/local-settings/localsettings'
import {AppEnv} from '@/app/env/app-env'
import Toastify from 'toastify-js'

export default {
  components: {
    EqWindowSimple,
    QuestFileTree,
    QuestEditorTabs,
    QuestHelperPanel,
  },
  data() {
    return {
      openTabs: [],
      activeTab: '',
      capabilities: {
        format_available: false,
        validate_available: false,
      },
      appEnvLocal: true,
      formatOnSave: false,
      fontSize: 14,
      contextNode: null,

      createModalPath: '',
      createModalType: 'blank',
      createModalTitle: 'Create File',
      createModalHint: '',

      renameOldPath: '',
      renameNewPath: '',

      deleteTargetPath: '',
      deleteTargetIsDir: false,
    }
  },
  computed: {
    hasDirtyTabs() {
      return this.openTabs.some(t => t.dirty)
    },
  },
  async mounted() {
    this.appEnvLocal = AppEnv.isAppLocal()
    this.formatOnSave = LocalSettings.isQuestEditorFormatOnSave()
    this.fontSize = LocalSettings.getQuestEditorFontSize()

    if (!this.appEnvLocal) {
      return
    }

    try {
      this.capabilities = await QuestEditorApi.getCapabilities()
    } catch (e) {
      // not available in non-local mode
    }

    window.addEventListener('beforeunload', this.handleBeforeUnload)
  },
  beforeUnmount() {
    window.removeEventListener('beforeunload', this.handleBeforeUnload)
  },
  beforeRouteLeave(to, from, next) {
    if (this.hasDirtyTabs) {
      const answer = window.confirm('You have unsaved changes. Are you sure you want to leave?')
      if (!answer) return next(false)
    }
    next()
  },
  methods: {
    handleBeforeUnload(e) {
      if (this.hasDirtyTabs) {
        e.preventDefault()
        e.returnValue = ''
      }
    },

    onTreeLoaded(treeData) {
      return treeData
    },

    async openFile(path) {
      const existing = this.openTabs.find(t => t.path === path)
      if (existing) {
        this.syncDirtyToTab()
        this.activeTab = path
        return
      }

      try {
        const data = await QuestEditorApi.getFile(path)
        if (data) {
          this.syncDirtyToTab()
          this.openTabs.push({
            path: data.path,
            content: data.contents,
            originalContent: data.contents,
            dirty: false,
          })
          this.activeTab = path
        }
      } catch (e) {
        this.showToast('Failed to open file: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    switchTab(path) {
      this.syncDirtyToTab()
      this.activeTab = path
    },

    closeTab(path) {
      const tab = this.openTabs.find(t => t.path === path)
      if (tab && tab.dirty) {
        if (!window.confirm(`Close ${path} without saving?`)) return
      }

      const idx = this.openTabs.findIndex(t => t.path === path)
      if (idx >= 0) this.openTabs.splice(idx, 1)

      if (this.activeTab === path) {
        if (this.openTabs.length > 0) {
          const newIdx = Math.min(idx, this.openTabs.length - 1)
          this.activeTab = this.openTabs[newIdx].path
        } else {
          this.activeTab = ''
        }
      }
    },

    onContentChange({ path, content }) {
      const tab = this.openTabs.find(t => t.path === path)
      if (tab) {
        tab.content = content
        tab.dirty = content !== tab.originalContent
      }
    },

    syncDirtyToTab() {
      if (this.$refs.editorTabs && this.activeTab) {
        const editorContent = this.$refs.editorTabs.getContent()
        const tab = this.openTabs.find(t => t.path === this.activeTab)
        if (tab && editorContent !== null) {
          tab.content = editorContent
          tab.dirty = editorContent !== tab.originalContent
        }
      }
    },

    async saveActiveFile() {
      if (!this.activeTab) return
      this.syncDirtyToTab()

      const tab = this.openTabs.find(t => t.path === this.activeTab)
      if (!tab) return

      try {
        if (this.formatOnSave && this.capabilities.format_available) {
          await this.formatFileContent(tab)
        }

        await QuestEditorApi.saveFile(tab.path, tab.content)
        tab.originalContent = tab.content
        tab.dirty = false
        this.showToast('Saved ' + tab.path, '#4aa8e8')
      } catch (e) {
        this.showToast('Save failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    async saveAllFiles() {
      this.syncDirtyToTab()
      for (const tab of this.openTabs) {
        if (tab.dirty) {
          try {
            if (this.formatOnSave && this.capabilities.format_available) {
              await this.formatFileContent(tab)
            }
            await QuestEditorApi.saveFile(tab.path, tab.content)
            tab.originalContent = tab.content
            tab.dirty = false
          } catch (e) {
            this.showToast('Save failed: ' + tab.path, '#e85a4a')
          }
        }
      }
      this.showToast('All files saved', '#4aa8e8')
    },

    async formatActiveFile() {
      if (!this.activeTab) return
      this.syncDirtyToTab()

      const tab = this.openTabs.find(t => t.path === this.activeTab)
      if (!tab) return

      try {
        await this.formatFileContent(tab)
        this.showToast('Formatted ' + tab.path, '#4aa8e8')
      } catch (e) {
        this.showToast('Format failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    async formatFileContent(tab) {
      const result = await QuestEditorApi.formatFile(tab.path, tab.content)
      if (result && result.error) {
        this.showToast('Format error: ' + result.error, '#e8a34a')
        return
      }
      if (result && result.formatted) {
        tab.content = result.formatted
        tab.dirty = tab.content !== tab.originalContent
        if (this.$refs.editorTabs && this.activeTab === tab.path) {
          this.$refs.editorTabs.setContent(result.formatted)
        }
      }
    },

    async validateActiveFile() {
      if (!this.activeTab) return
      this.syncDirtyToTab()

      const tab = this.openTabs.find(t => t.path === this.activeTab)
      if (!tab) return

      try {
        const result = await QuestEditorApi.validateFile(tab.path, tab.content)
        if (result) {
          if (result.valid) {
            this.showToast('Validation passed', '#4ae84a')
          } else {
            const errors = result.errors ? result.errors.join('\n') : 'Unknown error'
            this.showToast('Validation failed: ' + errors, '#e85a4a')
          }
        }
      } catch (e) {
        this.showToast('Validate failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    async reloadActiveFile() {
      if (!this.activeTab) return

      try {
        await QuestEditorApi.reloadFile(this.activeTab)
        this.showToast('Quests reloaded', '#4ae84a')
      } catch (e) {
        this.showToast('Reload failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    promptCreateFile(templateType) {
      this.createModalType = templateType || 'blank'
      this.createModalPath = ''
      this.createModalTitle = 'Create File'
      this.createModalHint = ''
      this.$bvModal.show('create-file-modal')
    },

    promptCreateFolder() {
      this.createModalPath = ''
      this.$bvModal.show('create-folder-modal')
    },

    async doCreateFile() {
      if (!this.createModalPath) return

      let content = ''
      const name = this.createModalPath.split('/').pop().replace('.lua', '')
      switch (this.createModalType) {
        case 'zone_controller':
          content = QuestTemplates.zoneController(name)
          break
        case 'npc':
          content = QuestTemplates.npc(name)
          break
        case 'item':
          content = QuestTemplates.item(name)
          break
        case 'global':
          content = QuestTemplates.global()
          break
        case 'plugin':
          content = QuestTemplates.plugin(name)
          break
        default:
          content = QuestTemplates.blank(this.createModalPath)
      }

      if (!this.createModalPath.endsWith('.lua')) {
        this.createModalPath += '.lua'
      }

      try {
        await QuestEditorApi.createFile(this.createModalPath, content)
        this.showToast('Created ' + this.createModalPath, '#4ae84a')
        if (this.$refs.fileTree) await this.$refs.fileTree.refreshTree()
        await this.openFile(this.createModalPath)
      } catch (e) {
        this.showToast('Create failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    async doCreateFolder() {
      if (!this.createModalPath) return

      try {
        await QuestEditorApi.createFolder(this.createModalPath)
        this.showToast('Created folder ' + this.createModalPath, '#4ae84a')
        if (this.$refs.fileTree) await this.$refs.fileTree.refreshTree()
      } catch (e) {
        this.showToast('Create failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    handleContextMenu({ node }) {
      this.contextNode = node

      const choice = prompt(
        `Actions for ${node.path}:\n1 = Rename / Move\n2 = Delete\n\nEnter choice:`
      )

      if (choice === '1') {
        this.renameOldPath = node.path
        this.renameNewPath = node.path
        this.$bvModal.show('rename-modal')
      } else if (choice === '2') {
        this.deleteTargetPath = node.path
        this.deleteTargetIsDir = node.is_directory
        this.$bvModal.show('delete-confirm-modal')
      }
    },

    async doRename() {
      if (!this.renameNewPath || !this.renameOldPath) return

      try {
        await QuestEditorApi.movePath(this.renameOldPath, this.renameNewPath)
        this.remapOpenTabsForMove(this.renameOldPath, this.renameNewPath)

        this.showToast('Moved to ' + this.renameNewPath, '#4ae84a')
        if (this.$refs.fileTree) await this.$refs.fileTree.refreshTree()
      } catch (e) {
        this.showToast('Move failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    async doDelete() {
      if (!this.deleteTargetPath) return

      try {
        await QuestEditorApi.deletePath(this.deleteTargetPath)
        this.removeOpenTabsForPath(this.deleteTargetPath)

        this.showToast('Deleted ' + this.deleteTargetPath, '#4ae84a')
        if (this.$refs.fileTree) await this.$refs.fileTree.refreshTree()
      } catch (e) {
        this.showToast('Delete failed: ' + (e.response?.data?.error || e.message), '#e85a4a')
      }
    },

    insertSnippet(snippet) {
      if (this.$refs.editorTabs && this.$refs.editorTabs.$refs.aceEditor) {
        const editor = this.$refs.editorTabs.$refs.aceEditor.editor
        if (editor) {
          editor.insert(snippet)
          editor.focus()
        }
      }
    },

    showToast(text, color) {
      Toastify({
        text: text,
        duration: 3000,
        gravity: 'bottom',
        position: 'right',
        backgroundColor: color || '#333',
      }).showToast()
    },

    remapOpenTabsForMove(oldPath, newPath) {
      const prefix = oldPath + '/'
      this.openTabs.forEach(tab => {
        if (tab.path === oldPath) {
          tab.path = newPath
          return
        }
        if (tab.path.startsWith(prefix)) {
          tab.path = newPath + tab.path.slice(oldPath.length)
        }
      })
      if (this.activeTab === oldPath) {
        this.activeTab = newPath
      } else if (this.activeTab.startsWith(prefix)) {
        this.activeTab = newPath + this.activeTab.slice(oldPath.length)
      }
    },

    removeOpenTabsForPath(targetPath) {
      const prefix = targetPath + '/'
      this.openTabs = this.openTabs.filter(tab => tab.path !== targetPath && !tab.path.startsWith(prefix))
      if (this.activeTab === targetPath || this.activeTab.startsWith(prefix)) {
        this.activeTab = this.openTabs.length > 0 ? this.openTabs[0].path : ''
      }
    },
  },
  watch: {
    formatOnSave(val) {
      LocalSettings.setQuestEditorFormatOnSave(val)
    },
  },
}
</script>

<style scoped>
.quest-editor-root {
  height: calc(100dvh - 120px);
  max-height: calc(100dvh - 120px);
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.quest-editor-blocked {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: #aaa;
  padding: 32px;
}

.quest-editor-window {
  height: 100%;
  max-height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.quest-editor-window >>> .quest-editor-window-content {
  flex: 1 1 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
  padding: 0 !important;
}

.quest-editor-toolbar {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
}

.quest-editor-body {
  display: flex;
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.quest-pane {
  min-height: 0;
}

.quest-pane-left {
  width: 250px;
  min-width: 200px;
  flex: 0 0 250px;
  max-height: 100%;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.quest-pane-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.quest-pane-right {
  width: 250px;
  min-width: 200px;
  flex: 0 0 250px;
  display: flex;
  flex-direction: column;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
  min-height: 0;
  overflow: hidden;
}

.quest-pane-left > .quest-file-tree,
.quest-pane-center > .quest-editor-tabs,
.quest-pane-right > .quest-helper-panel {
  flex: 1 1 auto;
  min-height: 0;
}
</style>
