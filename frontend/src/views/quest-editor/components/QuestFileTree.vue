<template>
  <div class="quest-file-tree">
    <div class="d-flex align-items-center mb-2 px-2">
      <b-input
        v-model="searchQuery"
        size="sm"
        placeholder="Search files..."
        @input="onSearch"
        class="flex-grow-1"
      />
      <b-dropdown
        size="sm"
        variant="outline-warning"
        class="ml-1"
        right
        no-caret
      >
        <template #button-content>
          <i class="fa fa-plus"></i>
        </template>
        <b-dropdown-item @click="$emit('create-file', '')">Blank Lua File</b-dropdown-item>
        <b-dropdown-item @click="$emit('create-file', 'zone_controller')">Zone Controller</b-dropdown-item>
        <b-dropdown-item @click="$emit('create-file', 'npc')">NPC Script</b-dropdown-item>
        <b-dropdown-item @click="$emit('create-file', 'item')">Item Script</b-dropdown-item>
        <b-dropdown-item @click="$emit('create-file', 'global')">Global Script</b-dropdown-item>
        <b-dropdown-item @click="$emit('create-file', 'plugin')">Plugin Module</b-dropdown-item>
        <b-dropdown-divider/>
        <b-dropdown-item @click="$emit('create-folder')">New Folder</b-dropdown-item>
      </b-dropdown>
      <b-button
        size="sm"
        variant="outline-primary"
        class="ml-1"
        @click="refreshTree"
        title="Refresh"
      >
        <i class="fa fa-refresh"></i>
      </b-button>
    </div>

    <div class="tree-container" v-if="filteredTree.length > 0">
      <quest-tree-node
        v-for="node in filteredTree"
        :key="node.path"
        :node="node"
        :active-file-path="activeFilePath"
        :expanded-dirs="expandedDirs"
        :level="0"
        @file-select="$emit('file-select', $event)"
        @toggle-dir="toggleDir"
        @context-menu="onContextMenu"
      />
    </div>
    <div v-else class="text-center text-muted p-3">
      <app-loader :is-loading="loading" padding="2"/>
      <span v-if="!loading">No files found</span>
    </div>
  </div>
</template>

<script type="ts">
import {QuestEditorApi} from '@/app/api/quest-editor-api'
import {debounce} from '@/app/utility/debounce'
import QuestTreeNode from './QuestTreeNode.vue'

export default {
  components: {
    QuestTreeNode,
  },
  props: {
    activeFilePath: { type: String, default: '' },
  },
  data() {
    return {
      treeData: [],
      filteredTree: [],
      expandedDirs: {},
      searchQuery: '',
      loading: false,
    }
  },
  async mounted() {
    await this.refreshTree()
  },
  methods: {
    async refreshTree() {
      this.loading = true
      try {
        this.treeData = this.buildTree(await QuestEditorApi.getTree())
        this.applyFilter()
        this.$emit('tree-loaded', this.treeData)
      } catch (e) {
        console.error('Failed to load tree', e)
      }
      this.loading = false
    },

    onSearch: debounce(function () {
      this.applyFilter()
    }, 200),

    applyFilter() {
      if (!this.searchQuery || this.searchQuery.trim() === '') {
        this.filteredTree = this.treeData
        return
      }

      const q = this.searchQuery.toLowerCase()
      this.filteredTree = this.filterNodes(this.treeData, q)
    },

    buildTree(flatNodes) {
      const nodeMap = {}
      const roots = []

      flatNodes.forEach(node => {
        nodeMap[node.path] = {
          ...node,
          children: [],
        }
      })

      Object.values(nodeMap).forEach((node) => {
        const parentPath = node.path.includes('/') ? node.path.split('/').slice(0, -1).join('/') : ''
        if (parentPath && nodeMap[parentPath]) {
          nodeMap[parentPath].children.push(node)
        } else {
          roots.push(node)
        }
      })

      const sortNodes = (nodes) => {
        nodes.sort((a, b) => {
          if (a.is_directory !== b.is_directory) {
            return a.is_directory ? -1 : 1
          }
          return a.name.localeCompare(b.name)
        })
        nodes.forEach(child => sortNodes(child.children))
      }
      sortNodes(roots)

      return roots
    },

    filterNodes(nodes, query) {
      const result = []
      for (const node of nodes) {
        const matches = node.path.toLowerCase().includes(query) || node.name.toLowerCase().includes(query)
        if (node.is_directory) {
          const children = this.filterNodes(node.children || [], query)
          if (matches || children.length > 0) {
            result.push({
              ...node,
              children,
            })
          }
        } else if (matches) {
          result.push(node)
        }
      }
      return result
    },

    toggleDir(path) {
      this.$set(this.expandedDirs, path, !this.expandedDirs[path])
    },

    onContextMenu(node) {
      this.$emit('context-menu', { node })
    },
  },
}
</script>

<style scoped>
.quest-file-tree {
  height: 100%;
  overflow-y: auto;
}

.tree-container {
  font-size: 13px;
}

.tree-item {
  padding: 3px 8px;
  cursor: pointer;
  white-space: nowrap;
  border-radius: 3px;
}

.tree-item:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

.tree-item-active {
  background-color: rgba(255, 193, 7, 0.2) !important;
}

.tree-icon {
  font-size: 12px;
  min-width: 16px;
  text-align: center;
}

.tree-label {
  font-size: 13px;
}

</style>
