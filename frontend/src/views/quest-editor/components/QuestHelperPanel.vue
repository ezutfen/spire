<template>
  <div class="quest-helper-panel" v-if="loaded">
    <div class="px-2 mb-2">
      <b-input
        v-model="search"
        size="sm"
        placeholder="Search API..."
        @input="onSearch"
      />
    </div>

    <div class="helper-section" v-if="searchResults.length > 0">
      <div class="helper-section-title">Search Results</div>
      <div
        v-for="item in searchResults"
        :key="item.name"
        class="helper-item"
        @click="insertSnippet(item.snippet)"
      >
        <code class="helper-code">{{ item.name }}</code>
        <div class="helper-params text-muted" v-if="item.params">{{ item.params }}</div>
      </div>
    </div>

    <div class="helper-section" v-if="searchResults.length === 0">
      <div class="helper-section-title">Methods (Lua)</div>
      <div
        v-for="type in Object.keys(methodTypes).slice(0, 10)"
        :key="type"
      >
        <div class="helper-type-header" @click="toggleType(type)">
          <i :class="['fa', expandedTypes[type] ? 'fa-chevron-down' : 'fa-chevron-right']" style="font-size:10px"></i>
          {{ type }} ({{ methodTypes[type].length }})
        </div>
        <div v-if="expandedTypes[type]" class="helper-type-methods">
          <div
            v-for="method in methodTypes[type].slice(0, 20)"
            :key="method.method"
            class="helper-item"
            @click="insertMethod(method)"
          >
            <code class="helper-code">{{ methodPrefix(method) }}{{ method.method }}</code>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script type="ts">
import {SpireApi} from '@/app/api/spire-api'
import {debounce} from '@/app/utility/debounce'

export default {
  data() {
    return {
      api: {},
      loaded: false,
      search: '',
      searchResults: [],
      methodTypes: {},
      expandedTypes: {},
    }
  },
  async mounted() {
    await this.loadDefinitions()
  },
  methods: {
    async loadDefinitions() {
      try {
        const r = await SpireApi.v1().get('/quest-api/definitions')
        if (r.data && r.data.data) {
          this.api = r.data.data
          if (this.api.lua && this.api.lua.methods) {
            this.methodTypes = this.api.lua.methods
          }
          this.loaded = true
        }
      } catch (e) {
        console.error('Failed to load API definitions', e)
      }
    },

    onSearch: debounce(function () {
      this.doSearch()
    }, 300),

    doSearch() {
      if (!this.search || this.search.trim() === '' || !this.api.lua) {
        this.searchResults = []
        return
      }

      const q = this.search.toLowerCase()
      const results = []

      if (this.api.lua.methods) {
        for (const [type, methods] of Object.entries(this.api.lua.methods)) {
          for (const method of methods) {
            if (method.method.toLowerCase().includes(q)) {
              results.push({
                name: method.method,
                params: method.params ? method.params.join(', ') : '',
                snippet: this.buildSnippet(type, method),
              })
            }
          }
          if (results.length > 50) break
        }
      }

      if (this.api.lua.events) {
        for (const event of this.api.lua.events) {
          if (event.event_identifier.toLowerCase().includes(q)) {
            results.push({
              name: event.event_identifier,
              params: '',
              snippet: this.buildEventSnippet(event),
            })
          }
          if (results.length > 50) break
        }
      }

      this.searchResults = results.slice(0, 50)
    },

    toggleType(type) {
      this.$set(this.expandedTypes, type, !this.expandedTypes[type])
    },

    methodPrefix(method) {
      return ''
    },

    buildSnippet(type, method) {
      const snakeCase = (s) => s.replace(/\W+/g, ' ').split(/ |\B(?=[A-Z])/).map(w => w.toLowerCase()).join('_')
      let prefix = type
      if (type === 'eq') {
        prefix = 'eq.'
      } else {
        prefix = snakeCase(type).toLowerCase() + ':'
      }
      const params = method.params ? method.params.join(', ') : ''
      return `${prefix}${method.method}(${params})`
    },

    buildEventSnippet(event) {
      return `-- ${event.event_identifier} event\nfunction event_${event.event_identifier.replace(/^event_/, '')}(e)\n  -- handler code\nend`
    },

    insertMethod(method) {
      const type = Object.keys(this.methodTypes).find(t =>
        this.methodTypes[t].some(m => m.method === method.method)
      )
      if (type) {
        this.insertSnippet(this.buildSnippet(type, method))
      }
    },

    insertSnippet(snippet) {
      this.$emit('insert-snippet', snippet)
    },
  },
}
</script>

<style scoped>
.quest-helper-panel {
  flex: 1 1 auto;
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  font-size: 12px;
}

.helper-section-title {
  font-weight: bold;
  padding: 4px 8px;
  font-size: 11px;
  text-transform: uppercase;
  color: #888;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.helper-type-header {
  padding: 3px 8px;
  cursor: pointer;
  font-weight: bold;
  font-size: 12px;
}

.helper-type-header:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.helper-type-methods {
  padding-left: 12px;
}

.helper-item {
  padding: 2px 8px;
  cursor: pointer;
  border-radius: 2px;
}

.helper-item:hover {
  background-color: rgba(255, 193, 7, 0.1);
}

.helper-code {
  font-size: 11px;
  color: #4a9de8;
}

.helper-params {
  font-size: 10px;
  padding-left: 8px;
}
</style>
