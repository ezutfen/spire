<template>
  <div>
    <div class="input-group input-group-sm">
      <input
        type="number"
        class="form-control form-control-sm"
        style="max-width: 120px"
        v-model.number="localId"
        @change="onIdChange"
        placeholder="Spell ID"
      >
      <input
        type="text"
        class="form-control form-control-sm"
        v-model="search"
        @keyup.enter="doSearch"
        placeholder="Search spell name..."
      >
      <div class="input-group-append">
        <b-button size="sm" variant="outline-primary" @click="doSearch">
          <i class="fa fa-search"></i>
        </b-button>
      </div>
    </div>

    <div v-if="resolvedName" class="mt-1" style="font-size: 13px">
      <i class="ra ra-book mr-1"></i>
      <strong>{{ resolvedName }}</strong>
      <router-link
        v-if="localId > 0"
        :to="spellEditRoute(localId)"
        class="ml-2"
        style="font-size: 12px"
      >
        Open in Spell Editor <i class="fa fa-external-link-alt"></i>
      </router-link>
    </div>

    <div v-if="results.length > 0" class="mt-2" style="max-height: 200px; overflow-y: auto; border: 1px solid #444">
      <div
        v-for="spell in results"
        :key="spell.id"
        @click="pick(spell)"
        class="p-1"
        style="cursor: pointer; font-size: 13px; border-bottom: 1px solid #333"
      >
        <strong>{{ spell.id }}</strong> — {{ spell.name }}
      </div>
    </div>
  </div>
</template>

<script>
import {SpireApi} from '@/app/api/spire-api'
import {ROUTE} from '@/routes'

export default {
  name: 'AaSpellPicker',
  props: {
    spellId: {type: Number, default: 0},
  },
  data() {
    return {
      localId: this.spellId,
      search: '',
      results: [],
      resolvedName: '',
    }
  },
  watch: {
    spellId(val) {
      this.localId = val
      this.resolveName()
    },
  },
  mounted() {
    this.resolveName()
  },
  methods: {
    spellEditRoute(id) {
      return ROUTE.SPELL_EDIT.replace('%s', id)
    },
    onIdChange() {
      this.$emit('input', parseInt(this.localId) || 0)
      this.resolveName()
    },
    pick(spell) {
      this.localId = spell.id
      this.resolvedName = spell.name
      this.results = []
      this.$emit('input', spell.id)
    },
    async doSearch() {
      if (!this.search.trim()) {
        this.results = []
        return
      }
      try {
        const r = await SpireApi.v1().get('/spells_new', {
          params: {where: 'name_like_' + this.search.trim(), limit: 25, orderBy: 'name'},
        })
        this.results = r.data || []
      } catch (e) {
        this.results = []
      }
    },
    async resolveName() {
      if (!this.localId || this.localId <= 0) {
        this.resolvedName = ''
        return
      }
      try {
        const r = await SpireApi.v1().get('/spells_new/' + this.localId, {
          params: {select: 'id.name'},
        })
        this.resolvedName = r.data && r.data.name ? r.data.name : ''
      } catch (e) {
        this.resolvedName = '(not found)'
      }
    },
  },
}
</script>

<style scoped></style>
