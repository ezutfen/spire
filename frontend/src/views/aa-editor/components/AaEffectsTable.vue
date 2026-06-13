<template>
  <div>
    <table class="table table-sm table-dark" style="font-size: 13px">
      <thead>
      <tr>
        <th style="width: 70px">Slot</th>
        <th style="width: 120px">Effect ID (SPA)</th>
        <th>Description</th>
        <th style="width: 110px">Base1</th>
        <th style="width: 110px">Base2</th>
        <th style="width: 60px"></th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="(effect, index) in effects" :key="index">
        <td>
          <input type="number" class="form-control form-control-sm" v-model.number="effect.slot" @change="emitChange">
        </td>
        <td>
          <select class="form-control form-control-sm" v-model.number="effect.effect_id" @change="emitChange">
            <option :value="0">-- None --</option>
            <option v-for="(def, id) in spaDefinitions" :value="parseInt(id)">
              {{ id }}) {{ def.effectName }}
            </option>
          </select>
        </td>
        <td>
          <span v-if="spaDefinitions[effect.effect_id]" style="color: #8fbaec; font-size: 12px">
            {{ spaDefinitions[effect.effect_id].spa }}
            <small class="text-muted">— {{ spaDefinitions[effect.effect_id].description }}</small>
          </span>
          <span v-else class="text-muted">—</span>
        </td>
        <td>
          <input type="number" class="form-control form-control-sm" v-model.number="effect.base_1" @change="emitChange">
        </td>
        <td>
          <input type="number" class="form-control form-control-sm" v-model.number="effect.base_2" @change="emitChange">
        </td>
        <td>
          <b-button size="sm" variant="outline-danger" @click="remove(index)">
            <i class="fa fa-trash"></i>
          </b-button>
        </td>
      </tr>
      <tr v-if="effects.length === 0">
        <td colspan="6" class="text-center text-muted">No effects</td>
      </tr>
      </tbody>
    </table>
    <b-button size="sm" variant="success" @click="add">
      <i class="fa fa-plus mr-1"></i> Add Effect
    </b-button>
  </div>
</template>

<script>
import {SPELL_SPA_DEFINITIONS} from '@/app/constants/eq-spell-spa-definitions'

export default {
  name: 'AaEffectsTable',
  props: {
    effects: {type: Array, default: () => []},
  },
  data() {
    return {
      spaDefinitions: SPELL_SPA_DEFINITIONS,
    }
  },
  methods: {
    add() {
      const nextSlot = this.effects.length
      this.effects.push({slot: nextSlot, effect_id: 0, base_1: 0, base_2: 0})
      this.emitChange()
    },
    remove(index) {
      this.effects.splice(index, 1)
      this.emitChange()
    },
    emitChange() {
      this.$emit('change')
    },
  },
}
</script>

<style scoped></style>
