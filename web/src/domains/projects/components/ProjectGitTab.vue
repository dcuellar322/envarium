<script setup lang="ts">
import { computed, ref } from 'vue'
import CheckoutControls from './CheckoutControls.vue'
import GitChangesPanel from './GitChangesPanel.vue'
import GitWorktreeList from './GitWorktreeList.vue'
import type { GitState, ProjectEnvironment, RuntimeAction } from '../../../api/generated/types.gen'

const props = defineProps<{
  git?: GitState
  pending?: boolean
  error?: boolean
  driver?: string
  lifecyclePending?: boolean
  operationError?: string
  environments: Array<ProjectEnvironment>
  environmentsPending: boolean
  environmentsError: boolean
  registrationPending: boolean
  registrationError?: string
}>()
const emit = defineEmits<{
  register: []
  refresh: []
  environmentAction: [id: string, action: RuntimeAction]
}>()
const controls = ref<InstanceType<typeof CheckoutControls>>()
const selectedGit = ref<GitState>()
const selectedId = ref('')
const git = computed(() =>
  selectedId.value && selectedId.value !== props.git?.projectId ? selectedGit.value : props.git,
)
function refresh() {
  void controls.value?.refresh()
  emit('refresh')
}
</script>

<template>
  <CheckoutControls
    v-if="props.git?.projectId"
    ref="controls"
    :project-id="props.git.projectId"
    :git="props.git"
    :environments="environments"
    @select="
      (id, snapshot) => {
        selectedId = id
        selectedGit = snapshot
      }
    "
    @refresh="emit('refresh')"
  />
  <p v-if="pending" class="panel-state" role="status">
    Reading repository status, files, and branches…
  </p>
  <div v-else-if="error" class="panel-state message--error" role="alert">
    Repository observation failed. Any previous snapshot may be stale.
    <button type="button" @click="refresh">Retry repository</button>
  </div>
  <p v-else-if="git && !git.repository" class="panel-state">
    This project is not a Git working tree.
  </p>
  <article v-if="git?.repository" class="panel">
    <header class="panel-head">
      <div>
        <p>Read-only snapshot</p>
        <h2>Repository state</h2>
      </div>
      <button type="button" :disabled="pending" @click="refresh">Refresh repository</button>
    </header>
    <dl class="fact-grid">
      <div>
        <dt>Branch</dt>
        <dd>{{ git?.detached ? 'Detached HEAD' : git?.branch || 'Unborn branch' }}</dd>
      </div>
      <div>
        <dt>HEAD</dt>
        <dd>
          <code>{{ git?.head?.slice(0, 12) ?? '—' }}</code>
        </dd>
      </div>
      <div>
        <dt>Ahead / behind</dt>
        <dd>{{ git?.ahead ?? 0 }} / {{ git?.behind ?? 0 }}</dd>
      </div>
      <div>
        <dt>Stashes</dt>
        <dd>{{ git?.stashes ?? 0 }}</dd>
      </div>
      <div>
        <dt>Modified</dt>
        <dd>{{ git?.changes.modified ?? 0 }}</dd>
      </div>
      <div>
        <dt>Untracked</dt>
        <dd>{{ git?.changes.untracked ?? 0 }}</dd>
      </div>
    </dl>
    <p v-if="git?.lastCommit" class="commit">
      <code>{{ git.lastCommit.shortHash }}</code> {{ git.lastCommit.subject }}
      <span>by {{ git.lastCommit.author }}</span>
    </p>
  </article>
  <GitChangesPanel v-if="git?.repository" :details="git.details" />
  <article v-if="git?.repository" class="panel">
    <header class="panel-head">
      <div>
        <p>Local refs</p>
        <h2>Branches</h2>
      </div>
    </header>
    <p v-if="git.operationState" class="message message--error" role="status">
      A {{ git.operationState }} operation is in progress. Resolve it before switching.
    </p>
    <p v-if="git.details?.branchesTruncated" class="panel-state">
      Showing the first 512 local branches.
    </p>
    <ul class="git-branches">
      <li v-for="branch in git.details?.branches ?? []" :key="branch.name">
        <strong>{{ branch.name }}</strong
        ><span v-if="branch.current" class="current-branch">Current</span>
        <code>{{ branch.head.slice(0, 12) }}</code
        ><small v-if="branch.upstream">Tracks {{ branch.upstream }}</small>
        <button
          v-if="!branch.current"
          class="btn"
          type="button"
          :disabled="controls?.pending === true"
          @click="controls?.branch(branch.name)"
        >
          Review switch
        </button>
      </li>
    </ul>
    <p v-if="!git.details?.branches.length" class="panel-state">No committed local branches yet.</p>
  </article>
  <article class="panel">
    <header class="panel-head">
      <div>
        <p>Parallel feature environments</p>
        <h2>Worktrees &amp; environments</h2>
      </div>
      <button
        type="button"
        :disabled="registrationPending || controls?.pending === true"
        @click="$emit('register')"
      >
        {{ registrationPending ? 'Registering…' : '↻ Reconcile worktrees' }}
      </button>
    </header>
    <p v-if="operationError" class="panel-state message--error" role="alert">
      {{ operationError }}
    </p>
    <p v-if="registrationError" class="panel-state message--error" role="alert">
      {{ registrationError }}
    </p>
    <p v-else-if="environmentsPending || pending" class="panel-state">
      Reading durable environment registrations…
    </p>
    <p v-else-if="environmentsError" class="panel-state message--error" role="alert">
      Worktree environments are unavailable.
    </p>
    <p v-if="driver === 'process'" class="panel-state">
      Native worktrees use independent ports. Commands or environment values must reference
      ${ports.id}; reconcile after changing port declarations.
    </p>
    <GitWorktreeList
      v-if="!registrationError && !environmentsPending && !pending && !environmentsError"
      :worktrees="props.git?.worktrees ?? []"
      :selected="selectedId || props.git?.projectId"
      :driver="driver"
      :environments="environments"
      :pending="controls?.pending === true || lifecyclePending === true"
      @move="(id) => controls?.move(id)"
      @action="(id, action) => $emit('environmentAction', id, action)"
    />
  </article>
</template>

<style scoped>
.git-branches {
  list-style: none;
  margin: 0;
  padding: 0;
}
.git-branches li {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border);
}
.git-branches strong {
  overflow-wrap: anywhere;
}
.git-branches code,
.git-branches small {
  color: var(--muted);
}
.current-branch {
  border-radius: 5px;
  background: color-mix(in srgb, var(--green) 12%, transparent);
  color: var(--green);
  padding: 3px 7px;
  font-size: 11px;
}
.git-branches li button {
  margin-left: auto;
}
</style>
