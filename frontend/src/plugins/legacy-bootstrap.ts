import {
  Teleport,
  computed,
  createVNode,
  defineComponent,
  h,
  inject,
  onBeforeUnmount,
  onMounted,
  provide,
  reactive,
  ref,
  renderSlot,
  watch,
  type App as VueApp,
  type Directive,
  type PropType,
} from 'vue'

type ModalController = {
  show: () => void
  hide: () => void
}

const modalRegistry = new Map<string, ModalController>()

const modalService = {
  register(id: string, controller: ModalController) {
    modalRegistry.set(id, controller)
  },
  unregister(id: string) {
    modalRegistry.delete(id)
  },
  show(id: string) {
    modalRegistry.get(id)?.show()
  },
  hide(id: string) {
    modalRegistry.get(id)?.hide()
  },
}

function withModelModifiers(value: string, modifiers?: Record<string, boolean>) {
  if (modifiers?.trim) {
    return value.trim()
  }

  return value
}

function normalizeNumber(value: string, modifiers?: Record<string, boolean>) {
  const normalized = withModelModifiers(value, modifiers)
  if (!modifiers?.number) {
    return normalized
  }

  if (normalized === '') {
    return null
  }

  const parsed = Number(normalized)
  return Number.isNaN(parsed) ? normalized : parsed
}

const BButton = defineComponent({
  name: 'BButton',
  inheritAttrs: false,
  props: {
    variant: { type: String, default: 'secondary' },
    size: { type: String, default: '' },
    type: { type: String, default: 'button' },
    disabled: { type: Boolean, default: false },
  },
  setup(props, { attrs, slots }) {
    return () =>
      h(
        'button',
        {
          ...attrs,
          type: props.type,
          disabled: props.disabled,
          class: ['btn', `btn-${props.variant}`, props.size ? `btn-${props.size}` : '', attrs.class],
        },
        slots.default?.(),
      )
  },
})

const BFormInput = defineComponent({
  name: 'BFormInput',
  inheritAttrs: false,
  props: {
    modelValue: {
      type: [String, Number] as PropType<string | number | null>,
      default: '',
    },
    modelModifiers: {
      type: Object as PropType<Record<string, boolean>>,
      default: () => ({}),
    },
  },
  emits: ['update:modelValue', 'input', 'change'],
  setup(props, { attrs, emit }) {
    const update = (event: Event) => {
      const value = normalizeNumber((event.target as HTMLInputElement).value, props.modelModifiers)
      emit('update:modelValue', value)
      emit('input', value)
      emit('change', value)
    }

    return () =>
      h('input', {
        ...attrs,
        value: props.modelValue ?? '',
        class: ['form-control', attrs.class],
        onInput: update,
        onChange: update,
      })
  },
})

const BFormTextarea = defineComponent({
  name: 'BFormTextarea',
  inheritAttrs: false,
  props: {
    modelValue: {
      type: String,
      default: '',
    },
    modelModifiers: {
      type: Object as PropType<Record<string, boolean>>,
      default: () => ({}),
    },
  },
  emits: ['update:modelValue', 'input', 'change'],
  setup(props, { attrs, emit }) {
    const update = (event: Event) => {
      const value = withModelModifiers((event.target as HTMLTextAreaElement).value, props.modelModifiers)
      emit('update:modelValue', value)
      emit('input', value)
      emit('change', value)
    }

    return () =>
      h(
        'textarea',
        {
          ...attrs,
          value: props.modelValue ?? '',
          class: ['form-control', attrs.class],
          onInput: update,
          onChange: update,
        },
        attrs.value,
      )
  },
})

const BFormSelectOption = defineComponent({
  name: 'BFormSelectOption',
  inheritAttrs: false,
  props: {
    value: {
      type: [String, Number, Boolean] as PropType<string | number | boolean | null>,
      default: null,
    },
  },
  setup(props, { attrs, slots }) {
    return () =>
      h(
        'option',
        {
          ...attrs,
          value: props.value ?? '',
        },
        slots.default?.(),
      )
  },
})

const BFormSelect = defineComponent({
  name: 'BFormSelect',
  inheritAttrs: false,
  props: {
    modelValue: {
      type: [String, Number] as PropType<string | number | null>,
      default: '',
    },
    modelModifiers: {
      type: Object as PropType<Record<string, boolean>>,
      default: () => ({}),
    },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { attrs, slots, emit }) {
    const update = (event: Event) => {
      const value = normalizeNumber((event.target as HTMLSelectElement).value, props.modelModifiers)
      emit('update:modelValue', value)
      emit('change', value)
    }

    return () =>
      h(
        'select',
        {
          ...attrs,
          value: props.modelValue ?? '',
          class: ['form-select', attrs.class],
          onInput: update,
          onChange: update,
        },
        slots.default?.(),
      )
  },
})

const BFormCheckbox = defineComponent({
  name: 'BFormCheckbox',
  inheritAttrs: false,
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    switch: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { attrs, slots, emit }) {
    const update = (event: Event) => {
      const checked = (event.target as HTMLInputElement).checked
      emit('update:modelValue', checked)
      emit('change', checked)
    }

    return () =>
      h('div', { class: ['form-check', props.switch ? 'form-switch' : '', attrs.class] }, [
        h('input', {
          ...attrs,
          checked: props.modelValue,
          class: ['form-check-input'],
          type: 'checkbox',
          onChange: update,
        }),
        slots.default ? h('label', { class: 'form-check-label' }, slots.default()) : null,
      ])
  },
})

const BInputGroup = defineComponent({
  name: 'BInputGroup',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['input-group', attrs.class] }, slots.default?.())
  },
})

const BInputGroupAppend = defineComponent({
  name: 'BInputGroupAppend',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['input-group-append', attrs.class] }, slots.default?.())
  },
})

const BInputGroupPrepend = defineComponent({
  name: 'BInputGroupPrepend',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['input-group-prepend', attrs.class] }, slots.default?.())
  },
})

const BInputGroupText = defineComponent({
  name: 'BInputGroupText',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('span', { ...attrs, class: ['input-group-text', attrs.class] }, slots.default?.())
  },
})

const BAlert = defineComponent({
  name: 'BAlert',
  inheritAttrs: false,
  props: {
    variant: { type: String, default: 'primary' },
    show: { type: [Boolean, Number], default: true },
    dismissible: { type: Boolean, default: false },
  },
  emits: ['dismissed'],
  setup(props, { attrs, slots, emit }) {
    const visible = ref(Boolean(props.show))

    watch(
      () => props.show,
      (value) => {
        visible.value = Boolean(value)
      },
    )

    return () =>
      visible.value
        ? h('div', { ...attrs, class: ['alert', `alert-${props.variant}`, attrs.class] }, [
            props.dismissible
              ? h(
                  'button',
                  {
                    type: 'button',
                    class: 'btn-close float-end',
                    onClick: () => {
                      visible.value = false
                      emit('dismissed')
                    },
                  },
                  [],
                )
              : null,
            slots.default?.(),
          ])
        : null
  },
})

const BBadge = defineComponent({
  name: 'BBadge',
  inheritAttrs: false,
  props: {
    variant: { type: String, default: 'secondary' },
    pill: { type: Boolean, default: false },
  },
  setup(props, { attrs, slots }) {
    return () =>
      h(
        'span',
        {
          ...attrs,
          class: ['badge', `bg-${props.variant}`, props.pill ? 'rounded-pill' : '', attrs.class],
        },
        slots.default?.(),
      )
  },
})

const BPagination = defineComponent({
  name: 'BPagination',
  props: {
    modelValue: { type: Number, default: 1 },
    totalRows: { type: Number, default: 0 },
    perPage: { type: Number, default: 10 },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
    const pageCount = computed(() => Math.max(1, Math.ceil(props.totalRows / props.perPage)))

    const setPage = (page: number) => {
      const nextPage = Math.min(pageCount.value, Math.max(1, page))
      emit('update:modelValue', nextPage)
      emit('change', nextPage)
    }

    return () =>
      h(
        'ul',
        { class: 'pagination' },
        Array.from({ length: pageCount.value }, (_, index) => {
          const page = index + 1
          return h(
            'li',
            {
              class: ['page-item', page === props.modelValue ? 'active' : ''],
            },
            [
              h(
                'button',
                {
                  type: 'button',
                  class: 'page-link',
                  onClick: () => setPage(page),
                },
                String(page),
              ),
            ],
          )
        }),
      )
  },
})

const BTabsContextKey = Symbol('BTabsContextKey')

const BTabs = defineComponent({
  name: 'BTabs',
  props: {
    contentClass: { type: String, default: '' },
    fill: { type: Boolean, default: false },
  },
  setup(props, { slots }) {
    const tabs = reactive<{ id: string; title: string }[]>([])
    const activeTab = ref<string>('')

    const registerTab = (tab: { id: string; title: string; selected?: boolean }) => {
      if (!tabs.some((entry) => entry.id === tab.id)) {
        tabs.push({ id: tab.id, title: tab.title })
      }

      if (!activeTab.value || tab.selected) {
        activeTab.value = tab.id
      }
    }

    const unregisterTab = (tabId: string) => {
      const index = tabs.findIndex((entry) => entry.id === tabId)
      if (index >= 0) {
        tabs.splice(index, 1)
      }

      if (activeTab.value === tabId) {
        activeTab.value = tabs[0]?.id || ''
      }
    }

    provide(BTabsContextKey, {
      registerTab,
      unregisterTab,
      activeTab,
      setActiveTab: (tabId: string) => {
        activeTab.value = tabId
      },
    })

    return () =>
      h('div', {}, [
        h(
          'ul',
          { class: ['nav', 'nav-tabs', props.fill ? 'nav-fill' : ''] },
          tabs.map((tab) =>
            h('li', { class: 'nav-item' }, [
              h(
                'button',
                {
                  type: 'button',
                  class: ['nav-link', activeTab.value === tab.id ? 'active' : ''],
                  onClick: () => {
                    activeTab.value = tab.id
                  },
                },
                tab.title,
              ),
            ]),
          ),
        ),
        h('div', { class: props.contentClass }, slots.default?.()),
      ])
  },
})

const BTab = defineComponent({
  name: 'BTab',
  props: {
    title: { type: String, required: true },
    active: { type: Boolean, default: false },
  },
  setup(props, { slots }) {
    const tabs = inject<{
      registerTab: (tab: { id: string; title: string; selected?: boolean }) => void
      unregisterTab: (id: string) => void
      activeTab: { value: string }
    }>(BTabsContextKey)

    const tabId = `b-tab-${Math.random().toString(36).slice(2)}`

    onMounted(() => {
      tabs?.registerTab({ id: tabId, title: props.title, selected: props.active })
    })

    onBeforeUnmount(() => {
      tabs?.unregisterTab(tabId)
    })

    return () => (tabs?.activeTab.value === tabId ? h('div', {}, slots.default?.()) : null)
  },
})

const BDropdown = defineComponent({
  name: 'BDropdown',
  props: {
    text: { type: String, default: '' },
  },
  setup(props, { slots, attrs }) {
    const open = ref(false)
    return () =>
      h('div', { class: ['dropdown', attrs.class] }, [
        h(
          'button',
          {
            type: 'button',
            class: 'btn btn-secondary dropdown-toggle',
            onClick: () => {
              open.value = !open.value
            },
          },
          slots.buttonContent ? slots.buttonContent() : props.text,
        ),
        h(
          'div',
          {
            class: ['dropdown-menu', open.value ? 'show' : ''],
            style: open.value ? '' : 'display: none;',
          },
          slots.default?.(),
        ),
      ])
  },
})

const BDropdownItem = defineComponent({
  name: 'BDropdownItem',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () =>
      h(
        'button',
        {
          ...attrs,
          type: 'button',
          class: ['dropdown-item', attrs.class],
        },
        slots.default?.(),
      )
  },
})

const BDropdownDivider = defineComponent({
  name: 'BDropdownDivider',
  setup() {
    return () => h('hr', { class: 'dropdown-divider' })
  },
})

const BAvatar = defineComponent({
  name: 'BAvatar',
  props: {
    src: { type: String, default: '' },
    text: { type: String, default: '' },
  },
  setup(props, { attrs }) {
    return () =>
      h(
        'span',
        {
          class: ['avatar', attrs.class],
        },
        props.src
          ? [h('img', { src: props.src, alt: props.text || 'avatar', class: 'rounded-circle' })]
          : props.text,
      )
  },
})

const BAvatarGroup = defineComponent({
  name: 'BAvatarGroup',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['d-inline-flex align-items-center gap-2', attrs.class] }, slots.default?.())
  },
})

const BFormGroup = defineComponent({
  name: 'BFormGroup',
  props: {
    label: { type: String, default: '' },
  },
  setup(props, { slots, attrs }) {
    return () =>
      h('div', { class: ['mb-3', attrs.class] }, [
        props.label ? h('label', { class: 'form-label' }, props.label) : null,
        slots.default?.(),
      ])
  },
})

const BPopover = defineComponent({
  name: 'BPopover',
  setup() {
    return () => null
  },
})

const BContainer = defineComponent({
  name: 'BContainer',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['container', attrs.class] }, slots.default?.())
  },
})

const BRow = defineComponent({
  name: 'BRow',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['row', attrs.class] }, slots.default?.())
  },
})

const BListGroup = defineComponent({
  name: 'BListGroup',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['list-group', attrs.class] }, slots.default?.())
  },
})

const BListGroupItem = defineComponent({
  name: 'BListGroupItem',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['list-group-item', attrs.class] }, slots.default?.())
  },
})

const BFormTag = defineComponent({
  name: 'BFormTag',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('span', { ...attrs, class: ['badge bg-secondary', attrs.class] }, slots.default?.())
  },
})

const BFormTags = defineComponent({
  name: 'BFormTags',
  inheritAttrs: false,
  setup(_, { attrs, slots }) {
    return () => h('div', { ...attrs, class: ['d-flex flex-wrap gap-2', attrs.class] }, slots.default?.())
  },
})

const BSpinner = defineComponent({
  name: 'BSpinner',
  inheritAttrs: false,
  setup(_, { attrs }) {
    return () => h('div', { ...attrs, class: ['spinner-border spinner-border-sm', attrs.class], role: 'status' })
  },
})

const BModal = defineComponent({
  name: 'BModal',
  props: {
    id: { type: String, required: true },
    title: { type: String, default: '' },
    centered: { type: Boolean, default: false },
    size: { type: String, default: '' },
    okTitle: { type: String, default: 'OK' },
    cancelTitle: { type: String, default: 'Cancel' },
    okOnly: { type: Boolean, default: false },
    hideFooter: { type: Boolean, default: false },
  },
  emits: ['show', 'shown', 'hide', 'hidden', 'ok', 'cancel'],
  setup(props, { slots, emit }) {
    const visible = ref(false)

    const show = () => {
      visible.value = true
      emit('show')
      emit('shown')
    }

    const hide = () => {
      visible.value = false
      emit('hide')
      emit('hidden')
    }

    onMounted(() => {
      modalService.register(props.id, { show, hide })
    })

    onBeforeUnmount(() => {
      modalService.unregister(props.id)
    })

    return () =>
      visible.value
        ? h(Teleport, { to: 'body' }, [
            h('div', { class: 'modal-backdrop fade show' }),
            h(
              'div',
              {
                class: ['modal fade show'],
                style: 'display: block;',
                tabindex: -1,
                onClick: (event: MouseEvent) => {
                  if (event.target === event.currentTarget) {
                    hide()
                  }
                },
              },
              [
                h('div', { class: ['modal-dialog', props.centered ? 'modal-dialog-centered' : '', props.size ? `modal-${props.size}` : ''] }, [
                  h('div', { class: 'modal-content' }, [
                    h('div', { class: 'modal-header' }, [
                      slots['modal-title'] ? renderSlot(slots, 'modal-title') : h('h5', { class: 'modal-title' }, props.title),
                      h('button', {
                        type: 'button',
                        class: 'btn-close',
                        'aria-label': 'Close',
                        onClick: hide,
                      }),
                    ]),
                    h('div', { class: 'modal-body' }, slots.default?.()),
                    !props.hideFooter
                      ? h(
                          'div',
                          { class: 'modal-footer' },
                          slots['modal-footer']
                            ? renderSlot(slots, 'modal-footer')
                            : [
                                !props.okOnly
                                  ? h(
                                      'button',
                                      {
                                        type: 'button',
                                        class: 'btn btn-secondary',
                                        onClick: () => {
                                          emit('cancel')
                                          hide()
                                        },
                                      },
                                      props.cancelTitle,
                                    )
                                  : null,
                                h(
                                  'button',
                                  {
                                    type: 'button',
                                    class: 'btn btn-primary',
                                    onClick: () => {
                                      emit('ok')
                                      hide()
                                    },
                                  },
                                  props.okTitle,
                                ),
                              ],
                        )
                      : null,
                  ]),
                ]),
              ],
            ),
          ])
        : null
  },
})

const bModalDirective: Directive = {
  mounted(el, binding) {
    const modalId = binding.value || Object.keys(binding.modifiers)[0]
    if (!modalId) {
      return
    }

    const handler = () => modalService.show(modalId)
    ;(el as HTMLElement).__bModalHandler = handler
    el.addEventListener('click', handler)
  },
  unmounted(el) {
    const handler = (el as HTMLElement).__bModalHandler
    if (handler) {
      el.removeEventListener('click', handler)
    }
  },
}

const registeredComponents = [
  BAlert,
  BBadge,
  BAvatar,
  BAvatarGroup,
  BButton,
  BContainer,
  BDropdown,
  BDropdownDivider,
  BDropdownItem,
  BFormCheckbox,
  BFormGroup,
  BFormInput,
  BFormSelect,
  BFormSelectOption,
  BFormTag,
  BFormTags,
  BFormTextarea,
  BInputGroup,
  BInputGroupAppend,
  BInputGroupPrepend,
  BInputGroupText,
  BListGroup,
  BListGroupItem,
  BModal,
  BPagination,
  BPopover,
  BRow,
  BSpinner,
  BTab,
  BTabs,
]

export const LegacyBootstrapPlugin = {
  install(app: VueApp) {
    for (const component of registeredComponents) {
      app.component(component.name, component)
    }

    app.directive('b-modal', bModalDirective)
    app.config.globalProperties.$bvModal = {
      show: (id: string) => modalService.show(id),
      hide: (id: string) => modalService.hide(id),
    }
  },
}
