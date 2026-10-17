import { focusManager, useQueryClient } from '@tanstack/vue-query'
import { inject, onBeforeUnmount, onMounted, provide, ref, type InjectionKey, type Ref } from 'vue'

const activityKey: InjectionKey<Ref<boolean>> = Symbol('presentation-activity')
type NativeWindow = Window & { __envariumNativeActive?: boolean }

export function usePresentationActivity(): Ref<boolean> {
  return inject(activityKey, ref(true))
}

export function providePresentationActivity(): Ref<boolean> {
  const client = useQueryClient()
  const active = ref(
    document.visibilityState !== 'hidden' &&
      (window as NativeWindow).__envariumNativeActive !== false,
  )
  provide(activityKey, active)
  function update() {
    const next =
      document.visibilityState !== 'hidden' &&
      (window as NativeWindow).__envariumNativeActive !== false
    if (next === active.value) return
    active.value = next
    focusManager.setFocused(next)
    if (!next) void client.cancelQueries()
    else void client.refetchQueries({ type: 'active', stale: true })
  }
  onMounted(() => {
    focusManager.setFocused(active.value)
    document.addEventListener('visibilitychange', update)
    window.addEventListener('envarium:activity', update)
  })
  onBeforeUnmount(() => {
    document.removeEventListener('visibilitychange', update)
    window.removeEventListener('envarium:activity', update)
    focusManager.setFocused(undefined)
  })
  return active
}
