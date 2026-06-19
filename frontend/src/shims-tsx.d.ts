import type { ComponentPublicInstance, VNode } from 'vue'

declare global {
  namespace JSX {
    interface Element extends VNode {}
    interface ElementClass extends ComponentPublicInstance {}
    interface IntrinsicElements {
      [elem: string]: any
    }
  }
}
