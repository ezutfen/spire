<template>
  <li class="nav-item" v-show="navId !== '' && config && config.label">

    <!-- top level nav -->
    <router-link
      v-if="config.to"
      :to="config.to"
      :exact="config.exact ? config.exact : false"
      class="nav-link"
    >
      <i :class="config.labelIcon" v-if="config.labelIcon"></i> {{ config.label }}
    </router-link>

    <!-- nested nav dropdown -->
    <a
      :class="'nav-link collapse ' + (isSectionOpen ? 'active' : 'collapsed')"
      :href="'#sidebar-' + navId"
      @click.prevent="toggleSection"
      role="button"
      :aria-expanded="isSectionOpen ? 'true' : 'false'"
      :aria-controls="'sidebar-' + navId"
      v-if="!config.to"
    >
      <i :class="config.labelIcon" v-if="config.labelIcon"></i>
      {{ config.label }}
    </a>
    <div
      :class="'collapse ' + (isSectionOpen ? 'show' : '')"
      :id="'sidebar-' + navId"
      v-if="!config.to"
    >
      <ul class="nav nav-sm flex-column">
        <li v-for="nav in config.navs">

          <!-- internal link -->
          <router-link
            :class="'nav-link collapse ' + (hasRoute(nav.to) || hasRouteInArray(nav.routes) ? 'active' : 'collapsed')"
            :to="nav.to"
            v-if="!nav.to.includes('http')"
            :exact="nav.exact ? nav.exact : false"
          >
            <i :class="nav.icon" v-if="nav.icon"></i>{{ nav.title }}
            <b-badge class="ml-3" variant="primary" v-if="nav.isAlpha">ALPHA</b-badge>
            <b-badge class="ml-3" variant="primary" v-if="nav.isNew">NEW!</b-badge>

          </router-link>

          <!-- external link -->
          <a
            :class="'nav-link collapse ' + (hasRoute(nav.to) || hasRouteInArray(nav.routes) ? 'active' : 'collapsed')"
            :href="nav.to"
            :target="nav.to"
            v-if="nav.to.includes('http')"
          >
            <i :class="nav.icon" v-if="nav.icon"></i>{{ nav.title }}
            <b-badge class="ml-3" variant="primary" v-if="nav.isAlpha">ALPHA</b-badge>
            <b-badge class="ml-3" variant="primary" v-if="nav.isNew">NEW!</b-badge>
          </a>
        </li>
      </ul>
    </div>
  </li>
</template>

<script>
import {generateUuid} from "@/app/utility/uuid";

export default {
  name: "NavSectionComponent",
  computed: {
    isRouteActive() {
      return this.hasRoute(this.config.routePrefixMatch) || this.hasRouteInArray(this.config.routePrefixMatches)
    },
    isSectionOpen() {
      return this.expanded || this.isRouteActive
    },
  },
  methods: {
    toggleSection() {
      this.expanded = !this.isSectionOpen
    },
    hasRouteInArray(matches) {
      let matched = false
      if (matches && matches.length > 0) {
        for (const m of matches) {
          if (this.$route.path.includes(m)) {
            matched = true
          }
        }
      }

      return matched
    },
    hasRoute: function (partial) {
      return (this.$route.path.indexOf(partial) > -1)
    } // config.topLevelIcon
  },
  data() {
    return {
      navId: "",
      expanded: false,
    }
  },
  props: {
    config: {
      type: Object
    }
  },
  created() {
    this.navId = generateUuid()
    this.expanded = this.isRouteActive
  }
}
</script>

<style scoped>

</style>
