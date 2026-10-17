<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, ref } from 'vue'

import type { ProjectEnvironment } from '../../../api/generated/types.gen'
import { loadAgentSessions, startAgentSession, stopTerminalSession } from '../api'

const props = defineProps<{ projectId: string; environments: Array<ProjectEnvironment> }>()
const emit = defineEmits<{ terminal: [] }>()
const queryClient = useQueryClient()
const provider = ref<'codex' | 'claude'>('codex')
const environmentId = ref('')
const sessions = useQuery({
  queryKey: computed(() => ['agent-sessions', props.projectId]),
  queryFn: () => loadAgentSessions(props.projectId),
  refetchInterval: 5_000,
})
const create = useMutation({
  mutationFn: () =>
    startAgentSession({
      projectId: props.projectId,
      provider: provider.value,
      environmentId: environmentId.value || undefined,
      columns: 120,
      rows: 36,
    }),
  onSuccess: async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['agent-sessions', props.projectId] }),
      queryClient.invalidateQueries({ queryKey: ['terminal-sessions', props.projectId] }),
    ])
  },
})
const terminate = useMutation({
  mutationFn: stopTerminalSession,
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['agent-sessions', props.projectId] }),
})
</script>

<template>
  <article class="agent-panel">
    <header>
      <div>
        <p>Provider-neutral session metadata</p>
        <h2>Coding agents</h2>
      </div>
      <span>User-visible terminal output only</span>
    </header>
    <div class="agent-disclosure" role="note">
      <strong>What Envarium records</strong>
      <p>
        Provider, project, checkout, working directory, lifecycle status, byte counts, and
        user-visible PTY output while connected. Envarium neither requests nor claims access to
        hidden reasoning.
      </p>
    </div>
    <form @submit.prevent="create.mutate()">
      <label
        >Provider<select v-model="provider">
          <option value="codex">Codex</option>
          <option value="claude">Claude Code</option>
        </select></label
      >
      <label v-if="environments.length"
        >Checkout<select v-model="environmentId">
          <option value="">Primary checkout</option>
          <option v-for="environment in environments" :key="environment.id" :value="environment.id">
            {{ environment.name }}
          </option>
        </select></label
      >
      <button type="submit" :disabled="create.isPending.value">
        {{ create.isPending.value ? 'Starting…' : 'Start agent' }}
      </button>
      <button type="button" @click="emit('terminal')">Open terminal tab</button>
    </form>
    <p v-if="create.isError.value" class="error" role="alert">{{ create.error.value?.message }}</p>
    <div v-if="sessions.data.value?.length" class="agent-list">
      <article v-for="session in sessions.data.value" :key="session.id">
        <div>
          <strong>{{ session.displayName }}</strong
          ><span>{{ session.workingDirectory }}</span>
        </div>
        <span class="status" :class="`status--${session.status}`">{{ session.status }}</span>
        <dl>
          <div>
            <dt>Provider</dt>
            <dd>{{ session.provider }}</dd>
          </div>
          <div>
            <dt>Checkout</dt>
            <dd>{{ session.environmentId ? 'worktree' : 'primary' }}</dd>
          </div>
          <div>
            <dt>Visible output</dt>
            <dd>
              {{ session.outputBytes.toLocaleString() }} bytes<span v-if="session.outputTruncated">
                · reconnect buffer truncated</span
              >
            </dd>
          </div>
          <div>
            <dt>Created</dt>
            <dd>{{ new Date(session.createdAt).toLocaleString() }}</dd>
          </div>
        </dl>
        <button
          v-if="['starting', 'active'].includes(session.status)"
          type="button"
          :disabled="terminate.isPending.value"
          @click="terminate.mutate(session.id)"
        >
          Terminate
        </button>
      </article>
    </div>
    <p v-else-if="sessions.isPending.value" class="empty">Loading agent sessions…</p>
    <p v-else-if="sessions.isError.value" class="error" role="alert">
      Agent sessions are unavailable.
    </p>
    <p v-else class="empty">No agent sessions have been started for this project.</p>
  </article>
</template>

<style scoped>
.agent-panel {
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 13px;
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--panel) 97%, transparent),
    color-mix(in srgb, var(--bg) 97%, transparent)
  );
}
header,
form,
.agent-list > article {
  display: flex;
  align-items: center;
  gap: 10px;
}
header {
  justify-content: space-between;
}
header p {
  margin: 0;
  color: var(--accent);
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}
header h2 {
  margin: 4px 0 0;
  font-size: 16px;
}
header > span {
  color: var(--green);
  font-size: 10px;
}
.agent-disclosure {
  margin: 14px 0;
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--accent) 20%, transparent);
  border-radius: 9px;
  background: color-mix(in srgb, var(--accent) 6%, transparent);
}
.agent-disclosure p {
  margin: 5px 0 0;
  max-width: 900px;
  color: var(--muted);
  line-height: 1.5;
}
form {
  flex-wrap: wrap;
  margin-bottom: 14px;
}
label {
  display: grid;
  gap: 3px;
  color: var(--soft);
  font-size: 9px;
  text-transform: uppercase;
}
select,
button {
  min-height: 32px;
  padding: 6px 9px;
  border: 1px solid var(--border);
  border-radius: 7px;
  background: var(--panel-2);
  color: var(--text);
}
form button {
  align-self: end;
}
form button[type='submit'] {
  border-color: color-mix(in srgb, var(--accent) 50%, transparent);
  background: var(--accent);
  color: var(--on-accent);
  font-weight: 800;
}
.agent-list {
  display: grid;
  gap: 8px;
}
.agent-list > article {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) auto;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 9px;
  background: var(--panel);
}
.agent-list article > div {
  display: grid;
  gap: 3px;
}
.agent-list article > div span {
  overflow: hidden;
  color: var(--muted);
  font-size: 10px;
  text-overflow: ellipsis;
}
.status {
  padding: 4px 7px;
  border-radius: 99px;
  background: color-mix(in srgb, var(--muted) 10%, transparent);
  color: var(--muted);
  font-size: 9px;
  text-transform: uppercase;
}
.status--active {
  background: color-mix(in srgb, var(--green) 10%, transparent);
  color: var(--green);
}
dl {
  display: grid;
  grid-column: 1/-1;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin: 5px 0 0;
}
dl > div {
  padding: 8px;
  border-radius: 7px;
  background: var(--panel-2);
}
dt {
  color: var(--soft);
  font-size: 9px;
  text-transform: uppercase;
}
dd {
  margin: 4px 0 0;
  overflow-wrap: anywhere;
}
.agent-list button {
  grid-column: 2;
  color: var(--red);
}
.empty,
.error {
  padding: 28px;
  color: var(--muted);
  text-align: center;
}
.error {
  color: var(--red);
}
@media (max-width: 760px) {
  header {
    align-items: flex-start;
    display: grid;
  }
  dl {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
