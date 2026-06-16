<template>
  <div class="row">
    <div class="col-xl-12">

      <!-- Header card -->
      <eq-window-simple class="mb-3">
        <div class="d-flex align-items-center" style="flex-wrap: wrap">
          <div style="flex: 1; min-width: 250px">
            <h4 class="mb-0" style="color: #fff">
              <i class="ra ra-star mr-1"></i>
              {{ isNew ? 'New AA Ability' : ('AA Ability #' + abilityId) }}
              <b-badge v-if="dirty" variant="warning" class="ml-2">Unsaved</b-badge>
            </h4>
            <div class="text-muted" style="font-size: 13px">
              <span v-if="!isNew">Editing "{{ form.name }}"</span>
              <span v-else>Define a new ability below</span>
            </div>
          </div>
          <div class="mt-2">
            <b-button variant="success" :disabled="saving" @click="save">
              <i class="fa fa-save mr-1"></i> {{ isNew ? 'Create' : 'Save' }}
            </b-button>
            <b-button variant="outline-primary" class="ml-1" :disabled="isNew" @click="duplicate">
              <i class="fa fa-copy mr-1"></i> Duplicate
            </b-button>
            <b-button variant="outline-danger" class="ml-1" :disabled="isNew" @click="confirmDelete">
              <i class="fa fa-trash mr-1"></i> Delete
            </b-button>
            <b-button variant="outline-secondary" class="ml-1" @click="$router.push(ROUTE.AA_LIST)">
              <i class="fa fa-arrow-left"></i> Back
            </b-button>
          </div>
        </div>
      </eq-window-simple>

      <loader-component v-if="loading"/>

      <div v-if="!loading">
        <eq-tabs selected="Ability">

          <!-- Ability tab -->
          <eq-tab name="Ability" :selected="true">
            <eq-window-simple title="Ability Properties" class="p-3">
              <div class="row">
                <div class="col-lg-4 col-sm-12 mb-2">
                  <label>Name *</label>
                  <input type="text" class="form-control" v-model="form.name" @input="markDirty">
                </div>
                <div class="col-lg-2 col-sm-6 mb-2">
                  <label>Category</label>
                  <select class="form-control" v-model.number="form.category" @change="markDirty">
                    <option v-for="(label, id) in metadata.categories" :value="parseInt(id)">
                      {{ id }}) {{ label }}
                    </option>
                  </select>
                </div>
                <div class="col-lg-2 col-sm-6 mb-2">
                  <label>Type</label>
                  <select class="form-control" v-model.number="form.type" @change="markDirty">
                    <option v-for="(label, id) in metadata.types" :value="parseInt(id)">
                      {{ id }}) {{ label }}
                    </option>
                  </select>
                </div>
                <div class="col-lg-2 col-sm-6 mb-2">
                  <label>Status</label>
                  <select class="form-control" v-model.number="form.status" @change="markDirty">
                    <option v-for="(label, id) in metadata.statuses" :value="parseInt(id)">
                      {{ id }}) {{ label }}
                    </option>
                  </select>
                </div>
                <div class="col-lg-2 col-sm-6 mb-2">
                  <label>Charges</label>
                  <input type="number" class="form-control" v-model.number="form.charges" @input="markDirty">
                </div>
              </div>

              <div class="row">
                <div class="col-lg-3 col-sm-6 mb-2">
                  <div class="form-check">
                    <input type="checkbox" class="form-check-input" id="enabled" :checked="form.enabled === 1" @change="form.enabled = $event.target.checked ? 1 : 0; markDirty()">
                    <label class="form-check-label" for="enabled">Enabled</label>
                  </div>
                </div>
                <div class="col-lg-3 col-sm-6 mb-2">
                  <div class="form-check">
                    <input type="checkbox" class="form-check-input" id="grant_only" :checked="form.grant_only === 1" @change="form.grant_only = $event.target.checked ? 1 : 0; markDirty()">
                    <label class="form-check-label" for="grant_only">Grant Only</label>
                  </div>
                </div>
                <div class="col-lg-3 col-sm-6 mb-2">
                  <div class="form-check">
                    <input type="checkbox" class="form-check-input" id="reset_on_death" :checked="form.reset_on_death === 1" @change="form.reset_on_death = $event.target.checked ? 1 : 0; markDirty()">
                    <label class="form-check-label" for="reset_on_death">Reset On Death</label>
                  </div>
                </div>
                <div class="col-lg-3 col-sm-6 mb-2">
                  <div class="form-check">
                    <input type="checkbox" class="form-check-input" id="auto_grant" :checked="form.auto_grant_enabled === 1" @change="form.auto_grant_enabled = $event.target.checked ? 1 : 0; markDirty()">
                    <label class="form-check-label" for="auto_grant">Auto-Grant Enabled</label>
                  </div>
                </div>
              </div>

              <hr>

              <div class="mb-3">
                <label><strong>Classes</strong></label>
                <aa-class-race-bitmask
                  v-model="form.classes"
                  :entries="classEntries"
                  @input="markDirty"
                />
              </div>

              <div class="mb-3">
                <label><strong>Races</strong></label>
                <aa-class-race-bitmask
                  v-model="form.races"
                  :entries="raceEntries"
                  @input="markDirty"
                />
              </div>

              <div class="mb-3">
                <label><strong>Deities</strong></label>
                <aa-class-race-bitmask
                  v-model="form.deities"
                  :entries="deityEntries"
                  @input="markDirty"
                />
              </div>

              <div class="text-muted" style="font-size: 12px">
                <i class="fa fa-info-circle"></i>
                First Rank ID (derived): <strong>{{ form.first_rank_id || '—' }}</strong>
              </div>
            </eq-window-simple>
          </eq-tab>

          <!-- Ranks tab -->
          <eq-tab name="Ranks">
            <eq-window-simple class="p-3">
              <div class="d-flex justify-content-between align-items-center mb-3">
                <h5 class="mb-0">Rank Chain ({{ form.ranks.length }})</h5>
                <div>
                  <b-button size="sm" variant="success" @click="addRank">
                    <i class="fa fa-plus mr-1"></i> Add Rank
                  </b-button>
                </div>
              </div>

              <div v-if="form.ranks.length === 0" class="text-center text-muted p-4">
                No ranks defined. Click "Add Rank" to begin.
              </div>

              <aa-rank-card
                v-for="(rank, index) in form.ranks"
                :key="rank.temp_id || ('rank-' + index)"
                :ref="'rankCard_' + index"
                :rank="rank"
                :index="index"
                :is-last="index === form.ranks.length - 1"
                :spell-types="metadata.spell_types"
                @change="markDirty"
                @move-up="moveRank(index, -1)"
                @move-down="moveRank(index, 1)"
                @remove="removeRank(index)"
              />

              <div v-if="form.ranks.length > 1" class="mt-2 text-muted" style="font-size: 12px">
                <i class="fa fa-info-circle"></i>
                Ranks are ordered top-to-bottom. The chain (prev_id / next_id / first_rank_id) is
                rebuilt automatically on save.
              </div>
            </eq-window-simple>
          </eq-tab>

          <!-- Preview / Validation tab -->
          <eq-tab name="Preview">
            <eq-window-simple title="Validation & Preview" class="p-3">
              <div class="mb-3">
                <b-button variant="primary" @click="runValidation">
                  <i class="fa fa-check-circle mr-1"></i> Run Validation
                </b-button>
              </div>

              <div v-if="validation">
                <div v-if="validation.errors.length" class="mb-3">
                  <h6 class="text-danger">Errors</h6>
                  <ul>
                    <li v-for="e in validation.errors" class="text-danger">{{ e }}</li>
                  </ul>
                </div>
                <div v-if="validation.warnings.length" class="mb-3">
                  <h6 class="text-warning">Warnings</h6>
                  <ul>
                    <li v-for="w in validation.warnings" class="text-warning">{{ w }}</li>
                  </ul>
                </div>
                <div v-if="validation.valid && validation.warnings.length === 0">
                  <b-alert variant="success" show>No issues detected.</b-alert>
                </div>
              </div>

              <hr>

              <h6>Cost / Level Progression</h6>
              <table class="table table-sm table-dark" style="font-size: 13px">
                <thead>
                <tr>
                  <th>Rank</th>
                  <th>Cost</th>
                  <th>Level Req</th>
                  <th>Expansion</th>
                  <th>Effects</th>
                  <th>Prereqs</th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="(rank, index) in form.ranks">
                  <td>#{{ index + 1 }}</td>
                  <td>{{ rank.cost }}</td>
                  <td>{{ rank.level_req }}</td>
                  <td>{{ (metadata.expansions || {})[rank.expansion] || rank.expansion }}</td>
                  <td>{{ (rank.effects || []).length }}</td>
                  <td>{{ (rank.prereqs || []).length }}</td>
                </tr>
                </tbody>
              </table>
            </eq-window-simple>
          </eq-tab>

        </eq-tabs>
      </div>

    </div>

    <!-- Duplicate options modal -->
    <b-modal id="aa-dup-modal" title="Duplicate AA Ability" ok-title="Duplicate" @ok="performDuplicate">
      <div class="form-group">
        <label>New Name</label>
        <input type="text" class="form-control" v-model="dupName">
      </div>
      <div class="form-check">
        <input type="checkbox" class="form-check-input" id="remap_self" v-model="dupRemapSelf">
        <label class="form-check-label" for="remap_self">
          Remap prereqs that reference this ability to the new copy
        </label>
      </div>
    </b-modal>

    <!-- Delete confirmation modal -->
    <b-modal
      id="aa-delete-modal-edit"
      title="Delete AA Ability"
      ok-title="Delete"
      ok-variant="danger"
      @ok="performDelete"
    >
      <p v-if="form">
        Delete <strong>{{ form.name }}</strong> (id {{ abilityId }}) and all
        <strong>{{ form.ranks.length }}</strong> rank(s)?
      </p>
      <p class="text-muted" style="font-size: 12px">
        <i class="fa fa-info-circle"></i>
        Referenced db_str strings and linked spells are preserved.
      </p>
    </b-modal>

  </div>
</template>

<script>
import EqWindowSimple from '@/components/eq-ui/EQWindowSimple'
import EqTabs from '@/components/eq-ui/EQTabs'
import EqTab from '@/components/eq-ui/EQTab'
import LoaderComponent from '@/components/LoaderComponent'
import AaRankCard from './components/AaRankCard'
import AaClassRaceBitmask from './components/AaClassRaceBitmask'
import {AaEditorApi} from '@/app/api/aa-editor-api'
import {ROUTE} from '@/routes'
import {DB_PLAYER_CLASSES_ALL} from '@/app/constants/eq-classes-constants'
import {DB_PLAYER_RACES} from '@/app/constants/eq-races-constants'
import {DB_DIETIES_FULL} from '@/app/constants/eq-deities-constants'
import Toastify from 'toastify-js'

export default {
  name: 'AaAbilityEditor',
  components: {EqWindowSimple, EqTabs, EqTab, LoaderComponent, AaRankCard, AaClassRaceBitmask},
  computed: {
    ROUTE() {
      return ROUTE
    },
  },
  data() {
    return {
      loading: false,
      saving: false,
      dirty: false,
      isNew: false,
      abilityId: 0,
      form: this.emptyForm(),
      metadata: {
        categories: {}, types: {}, spell_types: {}, statuses: {}, expansions: {},
      },
      classEntries: [],
      raceEntries: [],
      deityEntries: [],
      validation: null,
      dupName: '',
      dupRemapSelf: false,
    }
  },
  beforeRouteLeave(to, from, next) {
    if (this.dirty) {
      const answer = window.confirm('You have unsaved changes. Are you sure you want to leave?')
      if (!answer) {
        return next(false)
      }
    }
    next()
  },
  async mounted() {
    this.classEntries = this.buildClassEntries()
    this.raceEntries = this.buildRaceEntries()
    this.deityEntries = this.buildDeityEntries()

    await this.loadMetadata()

    const idParam = this.$route.params.id
    if (!idParam || idParam === 'new') {
      this.isNew = true
      this.form = this.emptyForm()
      // seed one default rank
      this.form.ranks.push(this.newRank(0))
      this.snapshot()
    } else {
      this.abilityId = parseInt(idParam)
      await this.load()
    }
  },
  methods: {
    emptyForm() {
      return {
        name: '',
        category: 1,
        classes: 0,
        races: 0,
        drakkin_heritage: 0,
        deities: 0,
        status: 0,
        type: 0,
        charges: 0,
        grant_only: 0,
        enabled: 1,
        reset_on_death: 0,
        auto_grant_enabled: 0,
        first_rank_id: 0,
        ranks: [],
      }
    },
    newRank(index) {
      return {
        temp_id: 'r-' + Date.now() + '-' + index,
        id: 0,
        upper_hotkey_sid: 0,
        lower_hotkey_sid: 0,
        title_sid: 0,
        desc_sid: 0,
        cost: 2,
        level_req: 51,
        spell_id: 0,
        spell: null,
        spell_type: 0,
        recast_time: 0,
        expansion: 7,
        prev_id: 0,
        next_id: 0,
        effects: [],
        prereqs: [],
        strings: {},
      }
    },
    buildClassEntries() {
      const out = []
      Object.keys(DB_PLAYER_CLASSES_ALL).forEach(k => {
        const c = DB_PLAYER_CLASSES_ALL[k]
        out.push({value: parseInt(k), mask: c.mask, label: c.class, short: c.short, icon: c.icon})
      })
      return out
    },
    buildRaceEntries() {
      const out = []
      Object.keys(DB_PLAYER_RACES).forEach(k => {
        const r = DB_PLAYER_RACES[k]
        out.push({value: parseInt(k), mask: parseInt(r.mask), label: r.race, short: r.short, icon: r.icon})
      })
      return out
    },
    buildDeityEntries() {
      const out = []
      Object.keys(DB_DIETIES_FULL).forEach(k => {
        const d = DB_DIETIES_FULL[k]
        out.push({value: parseInt(k), mask: d.mask, label: d.name, short: d.short})
      })
      return out
    },
    async loadMetadata() {
      try {
        this.metadata = await AaEditorApi.getMetadata()
      } catch (e) {
        // optional
      }
    },
    async load() {
      this.loading = true
      try {
        const full = await AaEditorApi.getAbility(this.abilityId)
        if (!full) {
          this.showToast('Ability not found')
          return
        }
        this.form = this.fullToForm(full)
        this.snapshot()
      } catch (e) {
        this.showToast('Failed to load ability: ' + (e.response?.data?.error || e.message))
      } finally {
        this.loading = false
      }
    },
    // Transform the server's AaAbilityFull response into the editable form.
    fullToForm(full) {
      const a = full.aa_ability
      const form = {
        name: a.name,
        category: a.category,
        classes: a.classes,
        races: a.races,
        drakkin_heritage: a.drakkin_heritage,
        deities: a.deities,
        status: a.status,
        type: a.type,
        charges: a.charges,
        grant_only: a.grant_only,
        enabled: a.enabled,
        reset_on_death: a.reset_on_death,
        auto_grant_enabled: a.auto_grant_enabled,
        first_rank_id: a.first_rank_id,
        ranks: [],
      }
      ;(full.ranks || []).forEach((r, i) => {
        const rank = r.aa_rank
        form.ranks.push({
          temp_id: 'r-' + rank.id + '-' + i,
          id: rank.id,
          upper_hotkey_sid: rank.upper_hotkey_sid,
          lower_hotkey_sid: rank.lower_hotkey_sid,
          title_sid: rank.title_sid,
          desc_sid: rank.desc_sid,
          cost: rank.cost,
          level_req: rank.level_req,
          spell_id: rank.spell,
          spell: r.spell,
          spell_type: rank.spell_type,
          recast_time: rank.recast_time,
          expansion: rank.expansion,
          prev_id: rank.prev_id,
          next_id: rank.next_id,
          effects: (r.effects || []).map(e => ({slot: e.slot, effect_id: e.effect_id, base_1: e.base_1, base_2: e.base_2})),
          prereqs: (r.prereqs || []).map(p => ({aa_id: p.aa_id, points: p.points})),
          strings: r.strings || {},
        })
      })
      return form
    },
    // Build the AaAbilityInput payload to send to the backend.
    buildPayload() {
      const payload = {
        name: this.form.name,
        category: parseInt(this.form.category) || 0,
        classes: parseInt(this.form.classes) || 0,
        races: parseInt(this.form.races) || 0,
        drakkin_heritage: parseInt(this.form.drakkin_heritage) || 0,
        deities: parseInt(this.form.deities) || 0,
        status: parseInt(this.form.status) || 0,
        type: parseInt(this.form.type) || 0,
        charges: parseInt(this.form.charges) || 0,
        grant_only: parseInt(this.form.grant_only) || 0,
        enabled: parseInt(this.form.enabled) || 0,
        reset_on_death: parseInt(this.form.reset_on_death) || 0,
        auto_grant_enabled: parseInt(this.form.auto_grant_enabled) || 0,
        ranks: [],
      }

      this.form.ranks.forEach((rank, index) => {
        // harvest inline strings from the string editor sub-component
        let strings = {}
        const ref = this.$refs['rankCard_' + index]
        if (ref && ref[0] && ref[0].harvestStrings) {
          strings = ref[0].harvestStrings()
        }

        payload.ranks.push({
          id: parseInt(rank.id) || 0,
          upper_hotkey_sid: parseInt(rank.upper_hotkey_sid) || 0,
          lower_hotkey_sid: parseInt(rank.lower_hotkey_sid) || 0,
          title_sid: parseInt(rank.title_sid) || 0,
          desc_sid: parseInt(rank.desc_sid) || 0,
          cost: parseInt(rank.cost) || 0,
          level_req: parseInt(rank.level_req) || 0,
          spell: parseInt(rank.spell_id) || 0,
          spell_type: parseInt(rank.spell_type) || 0,
          recast_time: parseInt(rank.recast_time) || 0,
          expansion: parseInt(rank.expansion) || 0,
          effects: (rank.effects || []).map(e => ({
            slot: parseInt(e.slot) || 0,
            effect_id: parseInt(e.effect_id) || 0,
            base_1: parseInt(e.base_1) || 0,
            base_2: parseInt(e.base_2) || 0,
          })),
          prereqs: (rank.prereqs || []).map(p => ({
            aa_id: parseInt(p.aa_id) || 0,
            points: parseInt(p.points) || 0,
          })),
          strings: strings,
        })
      })

      return payload
    },
    async save() {
      if (!this.form.name || !this.form.name.trim()) {
        this.showToast('Name is required')
        return
      }
      this.saving = true
      try {
        const payload = this.buildPayload()
        let full
        if (this.isNew) {
          full = await AaEditorApi.createAbility(payload)
          this.showToast('Created "' + full.aa_ability.name + '"', '#4ae84a')
          this.isNew = false
          this.abilityId = full.aa_ability.id
          this.$router.replace(ROUTE.AA_ABILITY_EDIT.replace(':id', full.aa_ability.id))
        } else {
          full = await AaEditorApi.saveAbility(this.abilityId, payload)
          this.showToast('Saved', '#4ae84a')
        }
        this.form = this.fullToForm(full)
        this.snapshot()
      } catch (e) {
        this.showToast('Save failed: ' + (e.response?.data?.error || e.message))
      } finally {
        this.saving = false
      }
    },
    duplicate() {
      this.dupName = (this.form.name || '') + ' (Copy)'
      this.dupRemapSelf = false
      this.$bvModal.show('aa-dup-modal')
    },
    async performDuplicate() {
      try {
        const created = await AaEditorApi.duplicateAbility(this.abilityId, {
          name: this.dupName,
          remap_self: this.dupRemapSelf,
        })
        this.showToast('Duplicated to "' + created.aa_ability.name + '"', '#4ae84a')
        this.$router.push(ROUTE.AA_ABILITY_EDIT.replace(':id', created.aa_ability.id))
      } catch (e) {
        this.showToast('Duplicate failed: ' + (e.response?.data?.error || e.message))
      }
    },
    confirmDelete() {
      this.$bvModal.show('aa-delete-modal-edit')
    },
    async performDelete() {
      try {
        await AaEditorApi.deleteAbility(this.abilityId)
        this.showToast('Deleted "' + this.form.name + '"', '#4ae84a')
        this.dirty = false
        this.$router.push(ROUTE.AA_LIST)
      } catch (e) {
        this.showToast('Delete failed: ' + (e.response?.data?.error || e.message))
      }
    },
    addRank() {
      this.form.ranks.push(this.newRank(this.form.ranks.length))
      this.markDirty()
    },
    removeRank(index) {
      if (!window.confirm('Remove rank #' + (index + 1) + '?')) {
        return
      }
      this.form.ranks.splice(index, 1)
      this.markDirty()
    },
    moveRank(index, dir) {
      const newIndex = index + dir
      if (newIndex < 0 || newIndex >= this.form.ranks.length) {
        return
      }
      const tmp = this.form.ranks[index]
      this.$set(this.form.ranks, index, this.form.ranks[newIndex])
      this.$set(this.form.ranks, newIndex, tmp)
      this.markDirty()
    },
    async runValidation() {
      try {
        this.validation = await AaEditorApi.previewRank(this.buildPayload())
      } catch (e) {
        this.showToast('Validation failed: ' + (e.response?.data?.error || e.message))
      }
    },
    markDirty() {
      this.dirty = true
    },
    snapshot() {
      this.dirty = false
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

<style scoped></style>
