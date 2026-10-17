/* eslint-disable vue/one-component-per-file */
import { cleanup, render } from '@testing-library/vue'
import { defineComponent, h, nextTick, ref } from 'vue'
import { afterEach, expect, test, vi } from 'vitest'

import { useEventConnection } from '../../src/domains/system/composables/useEventConnection'

class Socket extends EventTarget {
  static instances: Socket[] = []
  closed = false
  constructor(readonly url: string) {
    super()
    Socket.instances.push(this)
  }
  close() {
    this.closed = true
    this.dispatchEvent(new Event('close'))
  }
}
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
  Socket.instances = []
})
test('pauses UI events while hidden and resumes with the last observed sequence', async () => {
  vi.stubGlobal('WebSocket', Socket)
  const active = ref(true),
    observed = vi.fn()
  render(
    defineComponent({
      setup() {
        useEventConnection(observed, active)
        return () => h('div')
      },
    }),
  )
  const first = Socket.instances[0]!
  first.dispatchEvent(
    new MessageEvent('message', {
      data: JSON.stringify({ sequence: 42, type: 'runtime.observed' }),
    }),
  )
  expect(observed).toHaveBeenCalledTimes(1)
  active.value = false
  await nextTick()
  expect(first.closed).toBe(true)
  first.dispatchEvent(new MessageEvent('message', { data: '{}' }))
  expect(observed).toHaveBeenCalledTimes(1)
  active.value = true
  await nextTick()
  expect(Socket.instances).toHaveLength(2)
  expect(Socket.instances[1]!.url).toContain('after=42')
  cleanup()
  expect(Socket.instances[1]!.closed).toBe(true)
})
test('an initially hidden presentation opens no event connection', () => {
  vi.stubGlobal('WebSocket', Socket)
  render(
    defineComponent({
      setup() {
        useEventConnection(undefined, ref(false))
        return () => h('div')
      },
    }),
  )
  expect(Socket.instances).toHaveLength(0)
})
