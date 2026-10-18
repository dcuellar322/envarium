<script setup lang="ts">
import { computed } from 'vue'
import type {
  GitWorktree,
  ProjectEnvironment,
  RuntimeAction,
} from '../../../api/generated/types.gen'
const props = defineProps<{
  worktrees: GitWorktree[]
  environments: ProjectEnvironment[]
  pending: boolean
  driver?: string
}>()
defineEmits<{ action: [id: string, action: RuntimeAction] }>()
const pathKey = (path: string) => path.replaceAll('\\', '/').replace(/\/$/, '')
const checkouts = computed(() => {
  const entries = props.worktrees.map((tree) => ({
    tree,
    observed: true,
    environment: props.environments.find((item) => pathKey(item.path) === pathKey(tree.path)),
  }))
  for (const environment of props.environments) {
    if (!entries.some((entry) => pathKey(entry.tree.path) === pathKey(environment.path)))
      entries.push({
        observed: false,
        tree: {
          path: environment.path,
          branch: environment.branch,
          head: environment.head || '',
          detached: environment.detached,
          bare: environment.bare,
          locked: environment.locked,
        },
        environment,
      })
  }
  return entries
})
function canStart() {
  return props.driver === 'compose'
}
</script>

<template>
  <div v-if="checkouts.length" class="worktree-list">
    <article
      v-for="entry in checkouts"
      :key="entry.tree.path"
      :aria-label="`Worktree ${entry.tree.branch || 'detached'}`"
    >
      <header>
        <strong>{{
          entry.tree.detached ? 'Detached HEAD' : entry.tree.branch || 'Bare repository'
        }}</strong
        ><span v-if="entry.environment?.primary">Primary</span
        ><span v-if="entry.tree.locked">Locked</span><code>{{ entry.tree.head.slice(0, 12) }}</code>
      </header>
      <code class="checkout-path">{{ entry.tree.path }}</code>
      <template v-if="entry.environment">
        <p>
          {{ entry.environment.hostname }} ·
          {{ entry.environment.allocation.portLeases.length }} allocated ports ·
          {{ entry.environment.availability }}
        </p>
        <p v-if="!entry.observed" class="unavailable" role="status">
          This checkout was not returned by Git. Reconcile worktrees before starting it.
        </p>
        <p v-if="entry.environment.unavailableReason" class="unavailable" role="status">
          {{ entry.environment.unavailableReason }}
        </p>
        <p v-if="!entry.environment.primary && !canStart()" class="unavailable">
          {{
            driver === 'process'
              ? 'Isolated worktree starts currently support Compose. For native processes, switch branches in the main checkout and use its project controls.'
              : 'Runtime information is unavailable. Refresh the project before starting this worktree.'
          }}
        </p>
        <div v-if="entry.environment.availability === 'available'" class="worktree-actions">
          <button
            v-if="!entry.environment.primary"
            type="button"
            class="btn"
            :disabled="pending || !entry.observed || !canStart()"
            @click="$emit('action', entry.environment.id, 'start')"
          >
            Review start
          </button>
          <button
            v-if="!entry.environment.primary"
            type="button"
            class="btn"
            :disabled="pending || !entry.observed || !canStart()"
            @click="$emit('action', entry.environment.id, 'restart')"
          >
            Review restart
          </button>
          <button
            type="button"
            class="btn"
            :disabled="pending"
            @click="$emit('action', entry.environment.id, 'stop')"
          >
            Review stop
          </button>
        </div>
      </template>
      <p v-else>
        Not registered as a runtime environment. Reconcile worktrees to allocate an isolated
        runtime.
      </p>
    </article>
  </div>
  <p v-else class="panel-state">No Git worktrees were returned for this checkout.</p>
</template>

<style scoped>
.worktree-list {
  display: grid;
  gap: 12px;
}
article {
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg);
  min-width: 0;
}
header {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
header strong {
  overflow-wrap: anywhere;
}
header span {
  color: var(--accent);
  font-size: 10px;
  border: 1px solid var(--border);
  border-radius: 5px;
  padding: 3px 6px;
}
header code {
  color: var(--soft);
  font-size: 11px;
  margin-left: auto;
}
.checkout-path {
  overflow-wrap: anywhere;
  color: var(--text);
  font-size: 12px;
}
p {
  color: var(--muted);
  font-size: 12px;
  line-height: 1.5;
  margin: 10px 0 0;
}
p.unavailable {
  color: var(--yellow);
}
.worktree-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}
.worktree-actions button {
  min-height: 36px;
}
</style>
