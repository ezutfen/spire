<template>
  <div>
    <div v-for="field in fields" :key="field.key" class="row mb-2">
      <div class="col-lg-3 col-sm-12" style="font-size: 13px; padding-top: 6px">
        <strong>{{ field.label }}</strong>
        <small class="text-muted d-block">sid: {{ rank[field.sidField] || 0 }}</small>
      </div>
      <div class="col-lg-9 col-sm-12">
        <div class="input-group input-group-sm">
          <input
            type="text"
            class="form-control form-control-sm"
            v-model="stringValues[field.key]"
            :placeholder="'(no string)'"
            @input="emitChange"
          >
          <div class="input-group-append">
            <b-button
              size="sm"
              variant="outline-success"
              :disabled="!stringValues[field.key]"
              @click="markNew(field.key)"
            >
              {{ newFlags[field.key] ? 'New (allocate)' : 'Save as new' }}
            </b-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AaStringEditor',
  props: {
    rank: {type: Object, required: true},
    // strings loaded from the server, keyed by sid-field key (1=title,2=desc,3=upper_hotkey,4=lower_hotkey)
    loadedStrings: {type: Object, default: () => ({})},
  },
  data() {
    return {
      fields: [
        {key: 1, label: 'Title', sidField: 'title_sid'},
        {key: 2, label: 'Description', sidField: 'desc_sid'},
        {key: 3, label: 'Upper Hotkey', sidField: 'upper_hotkey_sid'},
        {key: 4, label: 'Lower Hotkey', sidField: 'lower_hotkey_sid'},
      ],
      stringValues: {},
      newFlags: {},
    }
  },
  watch: {
    loadedStrings: {
      immediate: true,
      handler() {
        this.syncFromLoaded()
      },
    },
  },
  mounted() {
    this.syncFromLoaded()
  },
  methods: {
    syncFromLoaded() {
      const out = {}
      for (const f of this.fields) {
        const loaded = this.loadedStrings[f.key]
        out[f.key] = loaded ? loaded.value : ''
      }
      this.stringValues = out
    },
    markNew(key) {
      this.$set(this.newFlags, key, !this.newFlags[key])
      this.emitChange()
    },
    // Produce the Strings map expected by the backend API. When a value is set
    // and flagged as new (or the field had no sid), the backend allocates a free
    // db_str id; otherwise it updates the existing referenced sid.
    buildStringsMap() {
      const out = {}
      for (const f of this.fields) {
        const val = (this.stringValues[f.key] || '').trim()
        if (val === '') {
          continue
        }
        const existingSid = parseInt(this.rank[f.sidField]) || 0
        const isNew = this.newFlags[f.key] || existingSid <= 0
        out[f.key] = {
          id: isNew ? 0 : existingSid,
          type: 1,
          value: val,
        }
      }
      return out
    },
    emitChange() {
      this.$emit('change')
    },
  },
}
</script>

<style scoped></style>
