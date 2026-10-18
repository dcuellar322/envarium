<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import type { GitState, ProjectEnvironment } from '../../../api/generated/types.gen'
import { loadProjectGitInventory } from '../api'
import { useCheckoutTransition } from '../composables/useCheckoutTransition'
import BranchSwitchDialog from './BranchSwitchDialog.vue'
import RuntimePlanDialog from './RuntimePlanDialog.vue'

const props = defineProps<{
  projectId: string
  environments: ProjectEnvironment[]
  git?: GitState
}>()
const emit = defineEmits<{ select: [id: string, git?: GitState]; refresh: [] }>()
const selected = ref(props.projectId)
const id = computed(() => selected.value)
const switching = useCheckoutTransition(
  computed(() => props.projectId),
  (target) => {
    if (target) selected.value = target
    emit('refresh')
  },
)
const inventory = useQuery({
  queryKey: computed(() => ['project-git', selected.value, 'full']),
  queryFn: ({ signal }) => loadProjectGitInventory(selected.value, signal),
  enabled: computed(() => selected.value !== props.projectId),
  shallow: true,
  refetchInterval: 10_000,
})
const git = computed(() => (selected.value === props.projectId ? props.git : inventory.data.value))
watch([selected, git], () => emit('select', selected.value, git.value), { immediate: true })
watch(
  () => props.projectId,
  (value) => {
    selected.value = value
  },
)
function sources() {
  const environment = props.environments.find((item) => item.id === id.value)
  if (!environment && id.value !== props.projectId) return [id.value]
  if (!environment || environment.primary) {
    return [
      props.projectId,
      ...props.environments.filter((item) => item.primary).map((item) => item.id),
    ]
  }
  return [environment.id]
}
function branch(name: string) {
  void switching.prepare(sources(), id.value, name)
}
function move(target: string) {
  void switching.prepare(sources(), target)
}
defineExpose({ branch, move, refresh: () => inventory.refetch(), pending: switching.active })
</script>

<template>
  <article class="panel checkout-controls">
    <label for="git-checkout">Inspect checkout</label>
    <select id="git-checkout" v-model="selected" :disabled="switching.active.value">
      <option :value="projectId">Main project checkout</option>
      <option
        v-for="environment in environments.filter((item) => !item.primary && !item.bare)"
        :key="environment.id"
        :value="environment.id"
        :disabled="environment.availability !== 'available'"
      >
        {{ environment.name }}
      </option>
    </select>
    <p>
      Branch changes guide you through stopping, switching, and reviewing the selected checkout’s
      fresh start plan. Other worktrees keep running.
    </p>
    <p v-if="inventory.isError.value && selected !== projectId" role="alert" class="message--error">
      Selected checkout could not be read. Any previous snapshot may be stale.
      <button type="button" @click="inventory.refetch()">Retry</button>
    </p>
    <p v-if="switching.error.value" class="message--error" role="alert">
      {{ switching.error.value }}
    </p>
    <p v-else-if="switching.message.value" role="status">{{ switching.message.value }}</p>
  </article>
  <BranchSwitchDialog
    v-if="switching.branchPlan.value"
    :plan="switching.branchPlan.value"
    :pending="switching.pending.value"
    @cancel="switching.cancel"
    @confirm="switching.confirm"
  />
  <RuntimePlanDialog
    v-if="switching.runtimePlan.value"
    :plan="switching.runtimePlan.value"
    :pending="switching.pending.value"
    @cancel="switching.cancel"
    @confirm="switching.confirm"
  />
</template>

<style scoped>
.checkout-controls {
  display: grid;
  gap: 12px;
}
label {
  font-weight: 700;
}
select {
  width: 100%;
  min-height: 40px;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg);
  color: var(--text);
}
p {
  margin: 0;
  color: var(--muted);
  line-height: 1.6;
}
</style>
