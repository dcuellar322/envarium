<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

defineProps<{ labelledby: string }>()
const emit = defineEmits<{ close: [] }>()
const dialog = ref<HTMLDialogElement>()
let previousFocus: HTMLElement | null = null
onMounted(() => {
  previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
  if (typeof dialog.value?.showModal === 'function') dialog.value.showModal()
  else if (dialog.value) dialog.value.open = true
})
onBeforeUnmount(() => {
  if (previousFocus?.isConnected) previousFocus.focus()
})
function backdrop(event: MouseEvent) {
  const bounds = dialog.value?.getBoundingClientRect()
  if (
    bounds &&
    (event.clientX < bounds.left ||
      event.clientX > bounds.right ||
      event.clientY < bounds.top ||
      event.clientY > bounds.bottom)
  )
    emit('close')
}
</script>

<template>
  <dialog
    ref="dialog"
    class="app-dialog"
    :aria-labelledby="labelledby"
    aria-modal="true"
    @cancel.prevent="emit('close')"
    @click.self="backdrop"
  >
    <slot />
  </dialog>
</template>

<style scoped>
:where(.app-dialog) {
  max-width: calc(100vw - 32px);
  max-height: calc(100dvh - 32px);
  margin: auto;
  padding: 0;
  border: 0;
  background: var(--panel);
  color: var(--text);
  overflow: auto;
}
.app-dialog::backdrop {
  background: color-mix(in srgb, var(--bg) 78%, transparent);
}
</style>
