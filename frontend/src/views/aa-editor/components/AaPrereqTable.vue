<template>
  <div>
    <table class="table table-sm table-dark" style="font-size: 13px">
      <thead>
      <tr>
        <th style="width: 120px">Required AA</th>
        <th>Name</th>
        <th style="width: 120px">Points</th>
        <th style="width: 60px"></th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="(prereq, index) in prereqs" :key="index">
        <td>
          <input
            type="number"
            class="form-control form-control-sm"
            v-model.number="prereq.aa_id"
            @change="emitChange"
            placeholder="AA id"
          >
        </td>
        <td>
          <span v-if="abilityNames[prereq.aa_id]" style="color: #8fbaec">
            {{ abilityNames[prereq.aa_id] }}
          </span>
          <span v-else class="text-muted">—</span>
        </td>
        <td>
          <input type="number" class="form-control form-control-sm" v-model.number="prereq.points" @change="emitChange">
        </td>
        <td>
          <b-button size="sm" variant="outline-danger" @click="remove(index)">
            <i class="fa fa-trash"></i>
          </b-button>
        </td>
      </tr>
      <tr v-if="prereqs.length === 0">
        <td colspan="4" class="text-center text-muted">No prerequisites</td>
      </tr>
      </tbody>
    </table>
    <b-button size="sm" variant="success" @click="add">
      <i class="fa fa-plus mr-1"></i> Add Prerequisite
    </b-button>
  </div>
</template>

<script>
import {SpireApi} from '@/app/api/spire-api'

export default {
  name: 'AaPrereqTable',
  props: {
    prereqs: {type: Array, default: () => []},
  },
  data() {
    return {
      abilityNames: {},
    }
  },
  watch: {
    prereqs: {
      deep: true,
      handler() {
        this.resolveNames()
      },
    },
  },
  mounted() {
    this.resolveNames()
  },
  methods: {
    add() {
      this.prereqs.push({aa_id: 0, points: 0})
      this.emitChange()
    },
    remove(index) {
      this.prereqs.splice(index, 1)
      this.emitChange()
    },
    emitChange() {
      this.$emit('change')
    },
    async resolveNames() {
      const ids = this.prereqs.map(p => p.aa_id).filter(id => id > 0 && !this.abilityNames[id])
      await Promise.all(ids.map(async (id) => {
        try {
          const r = await SpireApi.v1().get('/aa_ability/' + id)
          if (r && r.data && r.data.name) {
            this.$set(this.abilityNames, id, r.data.name)
          }
        } catch (e) {
          // not found / unreachable
        }
      }))
    },
  },
}
</script>

<style scoped></style>
