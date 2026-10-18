<script setup lang="ts">
import type { Project, AiProviderDescriptor } from '../../../api/generated/types.gen'
defineProps<{
  projects: Pick<Project, 'id' | 'displayName'>[]
  providers: Pick<AiProviderDescriptor, 'id' | 'name'>[]
  loading: boolean
  pending: boolean
}>()
defineEmits<{ run: [] }>()
const project = defineModel<string>('project', { required: true })
const provider = defineModel<string>('provider', { required: true })
</script>
<template>
  <div class="controls">
    <label
      >Project<select v-model="project" :disabled="!projects.length || pending">
        <option v-if="!projects.length" value="">
          {{ loading ? 'Loading projects…' : 'No trusted projects' }}
        </option>
        <option v-for="item in projects" :key="item.id" :value="item.id">
          {{ item.displayName }}
        </option>
      </select></label
    >
    <label
      >Optional AI<select v-model="provider" :disabled="!project || pending">
        <option value="">Deterministic only</option>
        <option v-for="item in providers" :key="item.id" :value="item.id">
          {{ item.name }}
        </option>
      </select></label
    >
    <button class="primary" :disabled="!project || pending" @click="$emit('run')">
      {{ pending ? 'Collecting…' : 'Run diagnosis' }}
    </button>
  </div>
</template>
<style scoped>
.controls {
  display: flex;
  align-items: end;
  gap: 8px;
}
.controls label {
  min-width: 170px;
  display: grid;
  gap: 5px;
  color: var(--soft);
  font-size: 10px;
  text-transform: uppercase;
}
select,
button {
  min-height: 38px;
  padding: 0 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel-strong);
  color: var(--text);
}
button {
  cursor: pointer;
  font-weight: 800;
  background: var(--accent);
  color: var(--on-accent);
}
:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
@media (max-width: 720px) {
  .controls {
    flex-wrap: wrap;
    width: 100%;
  }
  .controls label {
    flex: 1 1 170px;
  }
}
</style>
