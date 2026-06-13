<template>
  <div class="card mb-3" style="background-color: rgba(0,0,0,0.3)">
    <div class="card-header d-flex align-items-center" style="cursor: pointer" @click="collapsed = !collapsed">
      <div class="d-flex align-items-center" style="flex: 1">
        <b-button size="sm" variant="dark" class="mr-2">
          <i :class="collapsed ? 'fa fa-chevron-right' : 'fa fa-chevron-down'"></i>
        </b-button>
        <strong class="mr-3">Rank #{{ index + 1 }}</strong>
        <span class="text-muted mr-3" style="font-size: 13px">
          id: {{ rank.id || '—' }} |
          cost: {{ rank.cost }} |
          lvl: {{ rank.level_req }} |
          exp: {{ rank.expansion }}
        </span>
        <span v-if="rank.spell" class="text-muted" style="font-size: 12px">
          <i class="ra ra-book mr-1"></i>{{ rank.spell ? rank.spell.name : '' }}
        </span>
      </div>
      <div>
        <b-button
          size="sm"
          variant="outline-secondary"
          class="mr-1"
          :disabled="index === 0"
          @click.stop="$emit('move-up')"
        >
          <i class="fa fa-arrow-up"></i>
        </b-button>
        <b-button
          size="sm"
          variant="outline-secondary"
          class="mr-1"
          :disabled="isLast"
          @click.stop="$emit('move-down')"
        >
          <i class="fa fa-arrow-down"></i>
        </b-button>
        <b-button size="sm" variant="outline-danger" @click.stop="$emit('remove')">
          <i class="fa fa-trash"></i>
        </b-button>
      </div>
    </div>

    <div class="card-body" v-show="!collapsed">
      <div class="row mb-3">
        <div class="col-lg-2 col-sm-6">
          <label>Cost</label>
          <input type="number" class="form-control form-control-sm" v-model.number="rank.cost" @change="emitChange">
        </div>
        <div class="col-lg-2 col-sm-6">
          <label>Level Req</label>
          <input type="number" class="form-control form-control-sm" v-model.number="rank.level_req" @change="emitChange">
        </div>
        <div class="col-lg-2 col-sm-6">
          <label>Expansion</label>
          <input type="number" class="form-control form-control-sm" v-model.number="rank.expansion" @change="emitChange">
        </div>
        <div class="col-lg-2 col-sm-6">
          <label>Recast Time</label>
          <input type="number" class="form-control form-control-sm" v-model.number="rank.recast_time" @change="emitChange">
        </div>
        <div class="col-lg-2 col-sm-6">
          <label>Spell Type</label>
          <select class="form-control form-control-sm" v-model.number="rank.spell_type" @change="emitChange">
            <option v-for="(label, id) in spellTypes" :value="parseInt(id)">{{ id }}) {{ label }}</option>
          </select>
        </div>
      </div>

      <div class="row mb-3">
        <div class="col-lg-12">
          <label>Linked Spell</label>
          <aa-spell-picker v-model="rank.spell_id" :spell-id="rank.spell_id" @input="emitChange"/>
        </div>
      </div>

      <hr>

      <h6 class="mb-2"><i class="ra ra-burst-blob mr-1"></i> Effects</h6>
      <aa-effects-table :effects="rank.effects" @change="emitChange"/>

      <hr>

      <h6 class="mb-2"><i class="ra ra-interdiction mr-1"></i> Prerequisites</h6>
      <aa-prereq-table :prereqs="rank.prereqs" @change="emitChange"/>

      <hr>

      <h6 class="mb-2"><i class="ra ra-scroll-unfurled mr-1"></i> Strings (db_str, type 1)</h6>
      <aa-string-editor
        :rank="rank"
        :loaded-strings="rank.strings || {}"
        ref="stringEditor"
        @change="emitChange"
      />
    </div>
  </div>
</template>

<script>
import AaEffectsTable from './AaEffectsTable'
import AaPrereqTable from './AaPrereqTable'
import AaStringEditor from './AaStringEditor'
import AaSpellPicker from './AaSpellPicker'

export default {
  name: 'AaRankCard',
  components: {AaEffectsTable, AaPrereqTable, AaStringEditor, AaSpellPicker},
  props: {
    rank: {type: Object, required: true},
    index: {type: Number, required: true},
    isLast: {type: Boolean, default: false},
    spellTypes: {type: Object, default: () => ({})},
  },
  data() {
    return {
      collapsed: true,
    }
  },
  methods: {
    emitChange() {
      this.$emit('change')
    },
    // Called by parent before save to harvest inline string entries
    harvestStrings() {
      if (this.$refs.stringEditor) {
        this.rank.strings = this.$refs.stringEditor.buildStringsMap()
      }
      return this.rank.strings
    },
  },
}
</script>

<style scoped></style>
