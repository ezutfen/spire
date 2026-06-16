<template>
  <div>
    <div class="row">
      <div
        v-for="entry in entries"
        :key="entry.value"
        class="text-center p-0 mr-2 mb-2"
        style="min-width: 60px"
      >
        <div class="text-center" style="font-size: 12px">
          {{ entry.short || entry.label }}
        </div>
        <div class="text-center">
          <b-button
            size="sm"
            :variant="isSelected(entry) ? 'primary' : 'outline-secondary'"
            :class="{ 'aa-bitmask-selected': isSelected(entry) }"
            @click="toggle(entry)"
            class="aa-bitmask-button"
          >
            <span v-if="isSelected(entry)" class="aa-bitmask-check">
              <i class="fa fa-check"></i>
            </span>
            <span v-if="entry.icon" :class="'item-' + entry.icon" style="display: inline-block; width: 32px; height: 32px"></span>
            <span v-else style="font-size: 11px">{{ entry.short || '?' }}</span>
          </b-button>
        </div>
      </div>
    </div>
    <div class="mt-1" style="font-size: 12px; color: #aaa">
      Bitmask value: <strong>{{ value }}</strong>
      <b-button size="sm" variant="outline-success" class="ml-2" @click="selectAll">All</b-button>
      <b-button size="sm" variant="outline-danger" class="ml-1" @click="selectNone">None</b-button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AaClassRaceBitmask',
  props: {
    value: {type: Number, default: 0},
    entries: {type: Array, default: () => []},
  },
  methods: {
    isSelected(entry) {
      return (parseInt(this.value) & parseInt(entry.mask)) > 0
    },
    toggle(entry) {
      const mask = parseInt(entry.mask)
      let next = parseInt(this.value) || 0
      if ((next & mask) > 0) {
        next = next & ~mask
      } else {
        next = next | mask
      }
      this.$emit('input', next)
    },
    selectAll() {
      let all = 0
      this.entries.forEach(e => { all |= parseInt(e.mask) })
      this.$emit('input', all)
    },
    selectNone() {
      this.$emit('input', 0)
    },
  },
}
</script>

<style scoped>
.aa-bitmask-button {
  position: relative;
  width: 44px;
  min-height: 38px;
  padding: 2px 4px;
}

.aa-bitmask-selected {
  border-color: #63b3ff;
  box-shadow: 0 0 0 1px #0d6efd, 0 0 8px rgba(13, 110, 253, .75);
}

.aa-bitmask-check {
  align-items: center;
  background: #0d6efd;
  border: 1px solid #b7dbff;
  border-radius: 3px;
  color: #fff;
  display: flex;
  font-size: 9px;
  height: 15px;
  justify-content: center;
  line-height: 1;
  position: absolute;
  right: -5px;
  top: -5px;
  width: 15px;
  z-index: 2;
}
</style>
