<template>
  <component
    :is="tag"
    :class="className"
    v-html="sanitizedHtml"
  />
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify'
import { computed } from 'vue'

const props = defineProps({
  template: {
    type: String,
    default: '',
  },
  tag: {
    type: String,
    default: 'div',
  },
  class: {
    type: String,
    default: '',
  },
})

function normalizeLegacyMarkup(html: string) {
  if (!html.includes('<b-tabs')) {
    return html
  }

  const container = document.createElement('div')
  container.innerHTML = html

  container.querySelectorAll('b-tabs').forEach((tabsElement, tabsIndex) => {
    const tabNodes = Array.from(tabsElement.querySelectorAll(':scope > b-tab'))
    if (tabNodes.length === 0) {
      return
    }

    const nav = document.createElement('ul')
    nav.className = 'nav nav-tabs'

    const content = document.createElement('div')
    content.className = tabsElement.getAttribute('content-class') || 'mt-2'

    tabNodes.forEach((tabNode, tabIndex) => {
      const tabId = `trusted-tab-${tabsIndex}-${tabIndex}`
      const title = tabNode.getAttribute('title') || `Tab ${tabIndex + 1}`

      const navItem = document.createElement('li')
      navItem.className = 'nav-item'

      const button = document.createElement('button')
      button.type = 'button'
      button.className = `nav-link${tabIndex === 0 ? ' active' : ''}`
      button.textContent = title
      button.setAttribute('data-bs-toggle', 'tab')
      button.setAttribute('data-bs-target', `#${tabId}`)

      navItem.appendChild(button)
      nav.appendChild(navItem)

      const pane = document.createElement('div')
      pane.id = tabId
      pane.className = `tab-pane fade${tabIndex === 0 ? ' show active' : ''}`
      pane.innerHTML = tabNode.innerHTML
      content.appendChild(pane)
    })

    const wrapper = document.createElement('div')
    wrapper.appendChild(nav)
    wrapper.appendChild(content)

    tabsElement.replaceWith(wrapper)
  })

  return container.innerHTML
}

const sanitizedHtml = computed(() => {
  return DOMPurify.sanitize(normalizeLegacyMarkup(props.template ?? ''), {
    ADD_ATTR: ['target', 'rel', 'aria-controls', 'aria-selected', 'data-bs-toggle', 'data-bs-target'],
  })
})

const className = computed(() => props.class)
</script>
