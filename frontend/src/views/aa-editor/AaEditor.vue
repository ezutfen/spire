<template>
  <div class="row">
    <div class="col-xl-12">

      <db-connection-status-pill/>

      <eq-window-simple title="Alternate Advancement Abilities" class="p-3">

        <div class="row mb-3">
          <div class="col-lg-3 col-sm-12">
            <label>Name or ID</label>
            <input
              type="text"
              class="form-control"
              autofocus
              v-model="search"
              v-on:keyup.enter="loadList(0)"
              placeholder="Name or ID"
            >
          </div>
          <div class="col-lg-2 col-sm-12">
            <label>Category</label>
            <select class="form-control" v-model.number="category" @change="loadList(0)">
              <option value="0">-- Any --</option>
              <option v-for="(label, id) in metadata.categories" :key="'category-' + id" :value="parseInt(id)">
                {{ id }}) {{ label }}
              </option>
            </select>
          </div>
          <div class="col-lg-2 col-sm-12">
            <label>Enabled</label>
            <select class="form-control" v-model.number="enabled" @change="loadList(0)">
              <option value="-1">-- Any --</option>
              <option value="1">Enabled</option>
              <option value="0">Disabled</option>
            </select>
          </div>
          <div class="col-lg-2 col-sm-12">
            <label>Per Page</label>
            <select class="form-control" v-model.number="limit" @change="loadList(0)">
              <option value="25">25</option>
              <option value="50">50</option>
              <option value="100">100</option>
            </select>
          </div>
          <div class="col-lg-3 col-sm-12 mt-3">
            <b-button variant="primary" @click="loadList(0)">
              <i class="fa fa-search mr-1"></i> Search
            </b-button>
            <b-button variant="success" class="ml-1" :to="ROUTE.AA_ABILITY_NEW">
              <i class="fa fa-plus mr-1"></i> New Ability
            </b-button>
            <b-button variant="outline-secondary" class="ml-1" @click="loadList()">
              <i class="fa fa-refresh"></i>
            </b-button>
          </div>
        </div>

        <div class="d-flex align-items-center justify-content-between flex-wrap mb-3">
          <div class="text-muted" style="font-size: 13px">
            Selected {{ selectedCount }} ability<span v-if="selectedCount !== 1">ies</span> for export
          </div>
          <div class="mt-2 mt-md-0">
            <b-button
              size="sm"
              variant="outline-warning"
              :disabled="selectedCount === 0 || exporting"
              @click="exportSelected"
            >
              <i class="fa fa-download mr-1"></i>
              {{ exporting ? 'Exporting...' : ('Export Selected (' + selectedCount + ')') }}
            </b-button>
            <b-button
              size="sm"
              variant="outline-info"
              class="ml-1"
              :disabled="importPreviewing || importApplying"
              @click="openImportModal"
            >
              <i class="fa fa-upload mr-1"></i> Import Bundle
            </b-button>
          </div>
        </div>

        <loader-component v-if="loading"/>

        <div v-if="!loading">
          <table class="table table-sm table-dark" style="font-size: 13px">
            <thead>
            <tr>
              <th style="width: 45px" class="text-center">
                <input
                  type="checkbox"
                  :checked="allVisibleSelected"
                  :disabled="items.length === 0"
                  @change="toggleSelectAllVisible"
                >
              </th>
              <th style="width: 60px" @click="setOrder('id')" class="cursor-pointer">ID</th>
              <th @click="setOrder('name')" class="cursor-pointer">Name</th>
              <th style="width: 120px">Category</th>
              <th style="width: 80px">Ranks</th>
              <th style="width: 100px">First Cost</th>
              <th style="width: 90px">First Lvl</th>
              <th style="width: 80px">Type</th>
              <th style="width: 80px">Enabled</th>
              <th style="width: 220px">Actions</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="item in items" :key="item.id">
              <td class="text-center align-middle">
                <input
                  type="checkbox"
                  :checked="isSelected(item.id)"
                  @change="toggleSelection(item.id)"
                >
              </td>
              <td>{{ item.id }}</td>
              <td>
                <router-link :to="abilityEditRoute(item.id)" style="color: #8fbaec">
                  {{ item.name }}
                </router-link>
              </td>
              <td>{{ metadata.categories[item.category] || item.category }}</td>
              <td>
                <b-badge variant="info">{{ item.rank_count }}</b-badge>
              </td>
              <td>{{ item.first_rank_cost || '—' }}</td>
              <td>{{ item.first_rank_level || '—' }}</td>
              <td>{{ metadata.types[item.type] || item.type }}</td>
              <td>
                <b-badge :variant="item.enabled ? 'success' : 'secondary'">
                  {{ item.enabled ? 'Yes' : 'No' }}
                </b-badge>
              </td>
              <td>
                <b-button
                  size="sm"
                  variant="outline-primary"
                  @click="$router.push(abilityEditRoute(item.id))"
                >
                  <i class="fa fa-edit"></i> Edit
                </b-button>
                <b-button
                  size="sm"
                  variant="outline-success"
                  class="ml-1"
                  :disabled="duplicating === item.id"
                  @click="duplicate(item)"
                >
                  <i class="fa fa-copy"></i>
                </b-button>
                <b-button
                  size="sm"
                  variant="outline-danger"
                  class="ml-1"
                  @click="confirmDelete(item)"
                >
                  <i class="fa fa-trash"></i>
                </b-button>
              </td>
            </tr>
            <tr v-if="items.length === 0">
              <td colspan="10" class="text-center text-muted">No abilities found</td>
            </tr>
            </tbody>
          </table>

          <div class="d-flex justify-content-between align-items-center">
            <div class="text-muted" style="font-size: 13px">
              {{ total }} abilities
            </div>
            <div>
              <b-button
                size="sm"
                variant="outline-secondary"
                :disabled="page === 0"
                @click="loadList(page - 1)"
              >
                <i class="fa fa-chevron-left"></i> Prev
              </b-button>
              <span class="mx-2" style="font-size: 13px">Page {{ page + 1 }}</span>
              <b-button
                size="sm"
                variant="outline-secondary"
                :disabled="items.length < limit"
                @click="loadList(page + 1)"
              >
                Next <i class="fa fa-chevron-right"></i>
              </b-button>
            </div>
          </div>
        </div>
      </eq-window-simple>
    </div>

    <b-modal
      id="aa-import-modal"
      title="Import AA Bundle"
      hide-footer
      size="xl"
      @hidden="resetImportState"
    >
      <div class="row">
        <div class="col-lg-6">
          <h6>Target Connection</h6>
          <db-connection-status-pill/>
        </div>
        <div class="col-lg-6">
          <h6>Bundle File</h6>
          <input
            ref="importFileInput"
            type="file"
            class="form-control"
            accept="application/json,.json"
            @change="handleImportFile"
          >
          <div v-if="importFileName" class="text-muted mt-2" style="font-size: 12px">
            Loaded: {{ importFileName }}
          </div>
          <div v-if="importBundle && importBundle.source_connection" class="text-muted mt-2" style="font-size: 12px">
            Source: {{ formatSourceConnection(importBundle.source_connection) }}
          </div>
        </div>
      </div>

      <loader-component v-if="importPreviewing"/>

      <div v-if="importPreview" class="mt-3">
        <div class="d-flex align-items-center flex-wrap mb-3">
          <b-badge variant="success" class="mr-2">Create {{ importPreview.creates }}</b-badge>
          <b-badge variant="primary" class="mr-2">Update {{ importPreview.updates }}</b-badge>
          <b-badge :variant="importPreview.blocked > 0 ? 'danger' : 'secondary'">
            Blocked {{ importPreview.blocked }}
          </b-badge>
        </div>

        <b-alert v-if="importPreview.blocked > 0" variant="danger" show>
          Import is blocked until all missing prerequisite AAs and spells exist on the target connection.
        </b-alert>

        <b-alert v-if="importPreview.missing_prereq_aa_ids && importPreview.missing_prereq_aa_ids.length" variant="warning" show>
          Missing prerequisite AA ids: {{ importPreview.missing_prereq_aa_ids.join(', ') }}
        </b-alert>

        <b-alert v-if="importPreview.missing_spell_ids && importPreview.missing_spell_ids.length" variant="warning" show>
          Missing spell ids: {{ importPreview.missing_spell_ids.join(', ') }}
        </b-alert>

        <table class="table table-sm table-dark" style="font-size: 13px">
          <thead>
          <tr>
            <th style="width: 100px">AA ID</th>
            <th>Name</th>
            <th style="width: 100px">Action</th>
            <th>Notes</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="action in importPreview.actions" :key="'import-action-' + action.ability_id">
            <td>{{ action.ability_id }}</td>
            <td>{{ action.name }}</td>
            <td>
              <b-badge :variant="actionBadgeVariant(action.action)">
                {{ action.action }}
              </b-badge>
            </td>
            <td>
              <span v-if="!action.reasons || action.reasons.length === 0" class="text-muted">Ready</span>
              <div v-for="reason in action.reasons || []" :key="action.ability_id + '-' + reason">
                {{ reason }}
              </div>
            </td>
          </tr>
          </tbody>
        </table>
      </div>

      <div class="d-flex justify-content-end mt-3">
        <b-button variant="outline-secondary" @click="$bvModal.hide('aa-import-modal')">
          Close
        </b-button>
        <b-button
          variant="primary"
          class="ml-2"
          :disabled="!canApplyImport"
          @click="applyImport"
        >
          {{ importApplying ? 'Applying...' : 'Apply Import' }}
        </b-button>
      </div>
    </b-modal>

    <b-modal
      id="aa-delete-modal"
      title="Delete AA Ability"
      ok-title="Delete"
      ok-variant="danger"
      @ok="performDelete"
    >
      <p v-if="deleteTarget">
        Are you sure you want to delete <strong>{{ deleteTarget.name }}</strong> (id {{ deleteTarget.id }})?
      </p>
      <p class="text-warning" v-if="deleteTarget">
        This will permanently remove <strong>{{ deleteTarget.rank_count }}</strong> rank(s)
        and all of their effects and prerequisites.
      </p>
      <p class="text-muted" style="font-size: 12px">
        <i class="fa fa-info-circle"></i>
        Referenced db_str strings and linked spells are preserved (they are shared data).
      </p>
    </b-modal>

  </div>
</template>

<script>
import EqWindowSimple from '@/components/eq-ui/EQWindowSimple'
import LoaderComponent from '@/components/LoaderComponent'
import DbConnectionStatusPill from '@/components/DbConnectionStatusPill'
import {AaEditorApi} from '@/app/api/aa-editor-api'
import {ROUTE} from '@/routes'
import Toastify from 'toastify-js'

export default {
  name: 'AaEditor',
  components: {EqWindowSimple, LoaderComponent, DbConnectionStatusPill},
  computed: {
    ROUTE() {
      return ROUTE
    },
    selectedCount() {
      return this.selectedAbilityIds.length
    },
    allVisibleSelected() {
      return this.items.length > 0 && this.items.every(item => this.selectedAbilityIds.includes(item.id))
    },
    canApplyImport() {
      return !!(
        this.importBundle &&
        this.importPreview &&
        this.importPreview.valid &&
        this.importPreview.blocked === 0 &&
        !this.importApplying
      )
    },
  },
  data() {
    return {
      loading: false,
      duplicating: 0,
      exporting: false,
      importPreviewing: false,
      importApplying: false,
      search: '',
      category: 0,
      enabled: -1,
      limit: 50,
      page: 0,
      total: 0,
      orderBy: 'name',
      orderDir: 'asc',
      items: [],
      metadata: {categories: {}, types: {}},
      deleteTarget: null,
      selectedAbilityIds: [],
      importFileName: '',
      importBundle: null,
      importPreview: null,
    }
  },
  async mounted() {
    await this.loadMetadata()
    await this.loadList(0)
  },
  methods: {
    abilityEditRoute(id) {
      return ROUTE.AA_ABILITY_EDIT.replace(':id', id)
    },
    actionBadgeVariant(action) {
      if (action === 'create') {
        return 'success'
      }
      if (action === 'update') {
        return 'primary'
      }
      return 'danger'
    },
    setOrder(col) {
      if (this.orderBy === col) {
        this.orderDir = this.orderDir === 'asc' ? 'desc' : 'asc'
      } else {
        this.orderBy = col
        this.orderDir = 'asc'
      }
      this.loadList(0)
    },
    async loadMetadata() {
      try {
        this.metadata = await AaEditorApi.getMetadata()
      } catch (e) {
        // metadata optional
      }
    },
    async loadList(page) {
      if (typeof page === 'number') {
        this.page = page
      }
      this.loading = true
      try {
        const res = await AaEditorApi.listAbilities({
          search: this.search,
          category: this.category,
          enabled: this.enabled,
          limit: this.limit,
          page: this.page,
          orderBy: this.orderBy,
          orderDirection: this.orderDir,
        })
        this.items = res.items || []
        this.total = res.total || 0
      } catch (e) {
        this.showToast('Failed to load abilities: ' + (e.response?.data?.error || e.message))
      } finally {
        this.loading = false
      }
    },
    isSelected(id) {
      return this.selectedAbilityIds.includes(id)
    },
    toggleSelection(id) {
      if (this.isSelected(id)) {
        this.selectedAbilityIds = this.selectedAbilityIds.filter(selectedId => selectedId !== id)
      } else {
        this.selectedAbilityIds = [...this.selectedAbilityIds, id]
      }
    },
    toggleSelectAllVisible() {
      if (this.allVisibleSelected) {
        const visibleIds = this.items.map(item => item.id)
        this.selectedAbilityIds = this.selectedAbilityIds.filter(id => !visibleIds.includes(id))
        return
      }

      const merged = new Set(this.selectedAbilityIds)
      this.items.forEach(item => merged.add(item.id))
      this.selectedAbilityIds = Array.from(merged)
    },
    async exportSelected() {
      if (this.selectedCount === 0) {
        return
      }

      this.exporting = true
      try {
        const orderedIds = [...this.selectedAbilityIds].sort((a, b) => a - b)
        const bundle = await AaEditorApi.exportAbilities(orderedIds)
        const filename = 'aa-bundle-' + new Date().toISOString().replace(/[:.]/g, '-') + '.json'
        const blob = new Blob([JSON.stringify(bundle, null, 2)], {type: 'application/json'})
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = filename
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
        this.showToast('Exported ' + orderedIds.length + ' AA abilit' + (orderedIds.length === 1 ? 'y' : 'ies'), '#4ae84a')
      } catch (e) {
        this.showToast('Export failed: ' + (e.response?.data?.error || e.message))
      } finally {
        this.exporting = false
      }
    },
    openImportModal() {
      this.resetImportState()
      this.$bvModal.show('aa-import-modal')
    },
    resetImportState() {
      this.importFileName = ''
      this.importBundle = null
      this.importPreview = null
      this.importPreviewing = false
      this.importApplying = false
      if (this.$refs.importFileInput) {
        this.$refs.importFileInput.value = ''
      }
    },
    formatSourceConnection(sourceConnection) {
      if (!sourceConnection) {
        return 'Unknown source connection'
      }

      const parts = [sourceConnection.name || 'Unknown']
      if (sourceConnection.content_db_name) {
        parts.push(sourceConnection.content_db_name)
      }
      if (sourceConnection.is_default) {
        parts.push('default')
      }
      return parts.join(' • ')
    },
    async handleImportFile(event) {
      const file = event.target.files && event.target.files[0]
      if (!file) {
        return
      }

      this.importFileName = file.name
      this.importPreview = null
      this.importBundle = null

      try {
        const text = await file.text()
        this.importBundle = JSON.parse(text)
        await this.runImportPreview()
      } catch (e) {
        this.importBundle = null
        this.showToast('Failed to read bundle: ' + (e.message || 'Invalid JSON bundle'))
      }
    },
    async runImportPreview() {
      if (!this.importBundle) {
        return
      }

      this.importPreviewing = true
      try {
        this.importPreview = await AaEditorApi.previewImport(this.importBundle)
      } catch (e) {
        this.importPreview = null
        this.showToast('Preview failed: ' + (e.response?.data?.error || e.message))
      } finally {
        this.importPreviewing = false
      }
    },
    async applyImport() {
      if (!this.canApplyImport) {
        return
      }

      this.importApplying = true
      try {
        const result = await AaEditorApi.applyImport(this.importBundle)
        this.showToast('Imported ' + (result.applied_ability_ids || []).length + ' AA abilit' + ((result.applied_ability_ids || []).length === 1 ? 'y' : 'ies'), '#4ae84a')
        this.$bvModal.hide('aa-import-modal')
        await this.loadList(0)
      } catch (e) {
        this.showToast('Import failed: ' + (e.response?.data?.error || e.message))
      } finally {
        this.importApplying = false
      }
    },
    async duplicate(item) {
      this.duplicating = item.id
      try {
        const created = await AaEditorApi.duplicateAbility(item.id, {name: item.name + ' (Copy)'})
        this.showToast('Duplicated to "' + created.aa_ability.name + '" (id ' + created.aa_ability.id + ')')
        await this.loadList()
      } catch (e) {
        this.showToast('Duplicate failed: ' + (e.response?.data?.error || e.message))
      } finally {
        this.duplicating = 0
      }
    },
    confirmDelete(item) {
      this.deleteTarget = item
      this.$bvModal.show('aa-delete-modal')
    },
    async performDelete() {
      if (!this.deleteTarget) {
        return
      }
      try {
        await AaEditorApi.deleteAbility(this.deleteTarget.id)
        this.selectedAbilityIds = this.selectedAbilityIds.filter(id => id !== this.deleteTarget.id)
        this.showToast('Deleted "' + this.deleteTarget.name + '"')
        this.deleteTarget = null
        await this.loadList()
      } catch (e) {
        this.showToast('Delete failed: ' + (e.response?.data?.error || e.message))
      }
    },
    showToast(message, color) {
      Toastify({
        text: message,
        duration: 3000,
        gravity: 'bottom',
        position: 'right',
        backgroundColor: color || '#e85a4a',
      }).showToast()
    },
  },
}
</script>

<style scoped>
.cursor-pointer {
  cursor: pointer;
}
</style>
