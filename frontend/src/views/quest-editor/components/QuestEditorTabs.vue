<template>
  <div class="quest-editor-tabs">
    <div class="tab-bar d-flex" v-if="openTabs.length > 0">
      <div
        v-for="tab in openTabs"
        :key="tab.path"
        :class="['tab-item d-flex align-items-center', { 'tab-active': tab.path === activeTab }]"
        @click="switchTab(tab.path)"
      >
        <span
          :class="['tab-name text-truncate', { 'tab-dirty': tab.dirty }]"
          :title="tab.path"
        >{{ tabName(tab.path) }}</span>
        <span
          v-if="tab.dirty"
          class="tab-dirty-dot"
          title="Unsaved changes"
        >&bull;</span>
        <span class="tab-close" @click.stop="closeTab(tab.path)">&times;</span>
      </div>
    </div>

    <div class="editor-area" v-if="currentContent !== null">
      <editor
        ref="aceEditor"
        v-model="currentContent"
        @init="editorInit"
        lang="lua"
        theme="terminal"
        width="100%"
        height="100%"
      ></editor>
    </div>
    <div v-else class="no-file-open d-flex align-items-center justify-content-center">
      <div class="text-center text-muted">
        <i class="fa fa-file-code-o" style="font-size: 48px; opacity: 0.3"></i>
        <p class="mt-3">Select a file from the tree to begin editing</p>
      </div>
    </div>
  </div>
</template>

<script type="ts">
export default {
  props: {
    tabs: { type: Array, default: () => [] },
    activeTab: { type: String, default: '' },
    fontSize: { type: Number, default: 14 },
  },
  data() {
    return {
      currentContent: null,
      internalUpdate: false,
    }
  },
  computed: {
    openTabs() {
      return this.tabs
    },
  },
  watch: {
    activeTab(newVal) {
      this.loadActiveContent()
    },
    currentContent() {
      this.emitDirty()
    },
    tabs: {
      handler() {
        this.loadActiveContent()
      },
      deep: true,
    },
  },
  methods: {
    loadActiveContent() {
      const tab = this.tabs.find(t => t.path === this.activeTab)
      if (tab) {
        this.internalUpdate = true
        this.currentContent = tab.content
        this.$nextTick(() => {
          this.internalUpdate = false
        })
      } else {
        this.currentContent = null
      }
    },

    tabName(path) {
      return path.split('/').pop() || path
    },

    switchTab(path) {
      this.emitDirty()
      this.$emit('switch-tab', path)
    },

    closeTab(path) {
      this.$emit('close-tab', path)
    },

    emitDirty() {
      if (this.internalUpdate) return
      const tab = this.tabs.find(t => t.path === this.activeTab)
      if (tab && this.currentContent !== null) {
        this.$emit('content-change', { path: this.activeTab, content: this.currentContent })
      }
    },

    editorInit() {
      require('brace/ext/language_tools')
      require('brace/theme/terminal')
      require('brace/mode/lua')

      this.$nextTick(() => {
        if (this.$refs.aceEditor && this.$refs.aceEditor.editor) {
          this.$refs.aceEditor.editor.setFontSize(this.fontSize)
          this.$refs.aceEditor.editor.setOptions({
            showPrintMargin: false,
            wrap: true,
            tabSize: 2,
          })
        }
      })
    },

    getContent() {
      return this.currentContent
    },

    setContent(content) {
      this.internalUpdate = true
      this.currentContent = content
      this.$nextTick(() => {
        this.internalUpdate = false
      })
    },
  },
}
</script>

<style scoped>
.quest-editor-tabs {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.tab-bar {
  background-color: rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  overflow-x: auto;
  flex-shrink: 0;
}

.tab-item {
  padding: 6px 10px;
  cursor: pointer;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  white-space: nowrap;
  min-width: 0;
  max-width: 180px;
  font-size: 12px;
  align-items: center;
}

.tab-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.tab-active {
  background-color: rgba(255, 255, 255, 0.1);
  border-bottom: 2px solid #e8c34a;
}

.tab-name {
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-dirty {
  font-style: italic;
  color: #e8c34a;
}

.tab-dirty-dot {
  color: #e8c34a;
  margin-left: 4px;
  font-size: 16px;
  line-height: 1;
}

.tab-close {
  margin-left: 6px;
  font-size: 14px;
  opacity: 0.5;
  line-height: 1;
}

.tab-close:hover {
  opacity: 1;
  color: #e85a4a;
}

.editor-area {
  flex: 1;
  overflow: hidden;
}

.no-file-open {
  flex: 1;
  color: #888;
}
</style>
