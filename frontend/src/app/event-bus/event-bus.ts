import mitt from 'mitt'

type EventPayload = unknown

const emitter = mitt<Record<string, EventPayload>>()

export const EventBus = {
  $emit(event: string, payload?: EventPayload) {
    emitter.emit(event, payload)
  },
  $on(event: string, handler: (payload?: EventPayload) => void) {
    emitter.on(event, handler)
  },
  $off(event: string, handler?: (payload?: EventPayload) => void) {
    if (!handler) {
      emitter.all.delete(event)
      return
    }

    emitter.off(event, handler)
  },
}
