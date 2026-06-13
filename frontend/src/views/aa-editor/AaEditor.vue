<template>
  <div class="row">
    <div class="col-xl-12">

      <eq-window-simple title="Alternate Advancement Abilities" class="p-3">

        <!-- Toolbar -->
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
              <option v-for="(label, id) in metadata.categories" :value="parseInt(id)">
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

        <!-- Loader -->
        <loader-component v-if="loading"/>

        <!-- Table -->
        <div v-if="!loading">
          <table class="table table-sm table-dark" style="font-size: 13px">
            <thead>
            <tr>
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
              <td colspan="9" class="text-center text-muted">No abilities found</td>
            </tr>
            </tbody>
          </table>

          <!-- Pagination -->
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

    <!-- Delete confirmation modal -->
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
import {AaEditorApi} from '@/app/api/aa-editor-api'
import {ROUTE} from '@/routes'
import Toastify from 'toastify-js'

export default {
  name: 'AaEditor',
  components: {EqWindowSimple, LoaderComponent},
  computed: {
    ROUTE() {
      return ROUTE
    },
  },
  data() {
    return {
      loading: false,
      duplicating: 0,
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
