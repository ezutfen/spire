<template>
  <div class="tree-node">
    <div
      :class="[
        'tree-item d-flex align-items-center',
        { 'tree-item-active': activeFilePath === node.path },
        { 'tree-item-dir': node.is_directory }
      ]"
      :style="{ paddingLeft: `${level * 16 + 8}px` }"
      @click="onNodeClick"
      @contextmenu.prevent="$emit('context-menu', node)"
    >
      <i
        v-if="node.is_directory"
        :class="['mr-1 tree-icon fa', expanded ? 'fa-folder-open' : 'fa-folder']"
        style="color: #e8c34a"
      ></i>
      <i
        v-else
        class="mr-1 tree-icon fa fa-file-code-o"
        style="color: #4a9de8"
      ></i>
      <span class="tree-label text-truncate" :title="node.name">{{ node.name }}</span>
    </div>

    <div v-if="node.is_directory && expanded">
      <quest-tree-node
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :active-file-path="activeFilePath"
        :expanded-dirs="expandedDirs"
        :level="level + 1"
        @file-select="$emit('file-select', $event)"
        @toggle-dir="$emit('toggle-dir', $event)"
        @context-menu="$emit('context-menu', $event)"
      />
    </div>
  </div>
</template>

<script type="ts">
export default {
  name: 'QuestTreeNode',
  props: {
    node: { type: Object, required: true },
    activeFilePath: { type: String, default: '' },
    expandedDirs: { type: Object, default: () => ({}) },
    level: { type: Number, default: 0 },
  },
  computed: {
    expanded() {
      return !!this.expandedDirs[this.node.path]
    },
  },
  methods: {
    onNodeClick() {
      if (this.node.is_directory) {
        this.$emit('toggle-dir', this.node.path)
        return
      }
      this.$emit('file-select', this.node.path)
    },
  },
}
</script>
