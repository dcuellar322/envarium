import { browserSocketProtocols } from '../../session/bootstrap'
import { onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'

type ConnectionState = 'connecting' | 'connected' | 'disconnected'

type EnvariumEvent = {
  type?: string
  projectId?: string
  operationId?: string
  sequence?: number
}

export function useEventConnection(
  onEvent?: (event: EnvariumEvent) => void,
  active?: Ref<boolean>,
) {
  const state = ref<ConnectionState>('connecting')
  let socket: WebSocket | undefined
  let reconnectTimer: number | undefined
  let lastSequence = 0
  let attempts = 0
  let disposed = false

  function connect() {
    if (typeof WebSocket === 'undefined') {
      state.value = 'disconnected'
      return
    }
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const connection = new WebSocket(
      `${protocol}//${window.location.host}/ws/v1/events?after=${lastSequence}`,
      browserSocketProtocols(),
    )
    socket = connection
    connection.addEventListener('message', (message) => {
      if (socket !== connection || active?.value === false) return
      try {
        const event = JSON.parse(String(message.data)) as EnvariumEvent
        if (typeof event.sequence === 'number')
          lastSequence = Math.max(lastSequence, event.sequence)
        onEvent?.(event)
      } catch {
        socket?.close(1002, 'invalid event envelope')
        return
      }
      attempts = 0
      state.value = 'connected'
    })
    connection.addEventListener('close', () => {
      if (socket !== connection) return
      state.value = 'disconnected'
      if (!disposed && active?.value !== false) {
        const delay = Math.min(5_000, 250 * 2 ** attempts)
        attempts += 1
        reconnectTimer = window.setTimeout(connect, delay)
      }
    })
    connection.addEventListener('error', () => {
      if (socket !== connection) return
      state.value = 'disconnected'
      socket?.close()
    })
  }

  function pause() {
    if (reconnectTimer !== undefined) window.clearTimeout(reconnectTimer)
    reconnectTimer = undefined
    const previous = socket
    socket = undefined
    previous?.close(1000, 'presentation paused')
    state.value = 'connecting'
  }
  if (active)
    watch(
      active,
      (visible) => {
        if (visible) connect()
        else pause()
      },
      { flush: 'sync' },
    )
  onMounted(() => {
    if (active?.value !== false) connect()
  })

  onBeforeUnmount(() => {
    disposed = true
    if (reconnectTimer !== undefined) window.clearTimeout(reconnectTimer)
    socket?.close()
  })
  return state
}
