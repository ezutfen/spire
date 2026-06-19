import { createPinia } from 'pinia'
import { createApp, configureCompat, defineAsyncComponent } from 'vue'
import { Icon } from 'leaflet'

import App from './App.vue'
import router from './router'
import { LegacyBootstrapPlugin } from './plugins/legacy-bootstrap'

import 'bootstrap/dist/css/bootstrap.min.css'
import '@/assets/css/custom.css'
import './assets/css/theme.min.css'
import './assets/fonts/feather/feather.min.css'
import '@/components/eq-ui/styles/eq-ui.css'
import '@/components/eq-ui/styles/eq-ui-buttons.css'
import './assets/css/global.css'
import 'fontawesome-4.7'
import 'rpg-awesome/css/rpg-awesome.min.css'
import 'highlight.js/styles/tomorrow-night-bright.css'
import 'toastify-js/src/toastify.css'
import '@exuanbo/file-icons-js/dist/css/file-icons.min.css'
import 'leaflet/dist/leaflet.css'

import hljs from 'highlight.js/lib/core'
import json from 'highlight.js/lib/languages/json'

hljs.registerLanguage('json', json)

configureCompat({
  MODE: 2,
})

delete (Icon.Default.prototype as any)._getIconUrl
Icon.Default.mergeOptions({
  iconRetinaUrl: new URL('../node_modules/leaflet/dist/images/marker-icon-2x.png', import.meta.url).href,
  iconUrl: new URL('../node_modules/leaflet/dist/images/marker-icon.png', import.meta.url).href,
  shadowUrl: new URL('../node_modules/leaflet/dist/images/marker-shadow.png', import.meta.url).href,
})

const app = createApp(App)
const pinia = createPinia()

app.component('app-loader', defineAsyncComponent(() => import('@/components/LoaderComponent.vue')))

app.config.errorHandler = (err, _vm, info) => {
  console.error(`Error in ${info}:`, err)
}

app.use(pinia)
app.use(router)
app.use(LegacyBootstrapPlugin)

app.mount('#app')
