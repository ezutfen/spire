<template>
  <div v-show="isActive">
    <slot></slot>
  </div>
</template>

<script>
export default {
  name: 'EqTab',
  props: {
    name: { required: true },
    selected: { default: false }
  },

  data() {
    return {
      isActive: false
    };
  },
  computed: {
    href() {
      return '#' + this.name.toLowerCase().replace(/ /g, '-');
    }
  },

  inject: {
    eqTabsApi: { default: null }
  },

  mounted() {
    this.isActive = this.selected;
    if (this.eqTabsApi) {
      this.eqTabsApi.register(this);
    }
  },

  beforeUnmount() {
    if (this.eqTabsApi) {
      this.eqTabsApi.unregister(this);
    }
  }
}
</script>

<style scoped>

</style>
