<script setup lang="ts">
import { RouterLink } from 'vue-router'

import type {
  ActionDefinition,
  GitState,
  HealthResult,
  RuntimeLogEntry,
  RuntimeMetricSample,
  RuntimeObservation,
  RuntimeAction,
} from '../../../api/generated/types.gen'
import { formatBytes } from '../../../lib/format'
import type { RuntimeSelection } from '../api'
import ProjectEndpointsCard from './ProjectEndpointsCard.vue'
import ProjectLogPreview from './ProjectLogPreview.vue'
import ProjectServicesPanel from './ProjectServicesPanel.vue'

defineProps<{
  projectId: string
  runtime?: RuntimeObservation
  runtimePending: boolean
  metrics: Array<RuntimeMetricSample>
  recentLogs: Array<RuntimeLogEntry>
  logConnection: string
  cpu: number
  cpuAvailable: boolean
  memory: number
  memoryAvailable: boolean
  memoryLimit: number
  git?: GitState
  changes: number
  healthState: string
  requiredHealth: Array<HealthResult>
  endpoints: Array<ActionDefinition>
  quickActions: Array<ActionDefinition>
  actionPending: boolean
  lifecyclePending: boolean
}>()
const emit = defineEmits<{
  selectTab: [tab: 'logs' | 'git']
  runAction: [action: ActionDefinition]
  lifecycle: [action: RuntimeAction, selection: RuntimeSelection]
}>()

function forwardLifecycle(action: RuntimeAction, selection: RuntimeSelection) {
  emit('lifecycle', action, selection)
}
</script>

<template>
  <div class="overview-grid">
    <div class="overview-main">
      <ProjectServicesPanel
        :runtime="runtime"
        :pending="runtimePending"
        :metrics="metrics"
        :action-pending="lifecyclePending"
        @lifecycle="forwardLifecycle"
      />

      <ProjectEndpointsCard
        v-if="endpoints.length"
        :endpoints="endpoints"
        :action-pending="actionPending"
        @run-action="emit('runAction', $event)"
      />

      <ProjectLogPreview
        :recent-logs="recentLogs"
        :log-connection="logConnection"
        @view-all="emit('selectTab', 'logs')"
      />
    </div>

    <aside class="overview-side">
      <article class="panel resource-card">
        <header class="panel-head">
          <div>
            <p>Current sample</p>
            <h2>Resources</h2>
          </div>
        </header>
        <dl>
          <div>
            <dt>CPU</dt>
            <dd>{{ cpuAvailable ? `${cpu.toFixed(1)}%` : '—' }}</dd>
            <span><i :style="{ width: `${cpuAvailable ? Math.min(cpu, 100) : 0}%` }"></i></span>
          </div>
          <div>
            <dt>Memory</dt>
            <dd>{{ memoryAvailable ? formatBytes(memory) : '—' }}</dd>
            <span
              ><i
                :style="{
                  width: `${memoryAvailable && memoryLimit > 0 ? Math.min((memory / memoryLimit) * 100, 100) : 0}%`,
                }"
              ></i
            ></span>
          </div>
        </dl>
        <RouterLink :to="{ name: 'resources', query: { project: projectId } }"
          >Open resource view →</RouterLink
        >
      </article>

      <article class="panel">
        <header class="panel-head">
          <div>
            <p>Repository</p>
            <h2>Git</h2>
          </div>
        </header>
        <dl class="compact-facts">
          <div>
            <dt>Branch</dt>
            <dd>{{ git?.branch ?? (git?.detached ? 'detached' : '—') }}</dd>
          </div>
          <div>
            <dt>Working tree</dt>
            <dd :class="{ warn: changes }">
              {{ changes ? `${changes} changes` : 'Clean' }}
            </dd>
          </div>
          <div>
            <dt>Upstream</dt>
            <dd>
              {{ git ? `${git.ahead} ahead · ${git.behind} behind` : '—' }}
            </dd>
          </div>
        </dl>
        <button type="button" @click="emit('selectTab', 'git')">Inspect repository →</button>
      </article>

      <article class="panel">
        <header class="panel-head">
          <div>
            <p>Checks</p>
            <h2>Health</h2>
          </div>
          <span>{{ healthState }}</span>
        </header>
        <ul class="health-list">
          <li v-for="check in requiredHealth" :key="check.checkId">
            <i :class="{ online: check.status === 'healthy' }"></i>
            <span
              ><strong>{{ check.serviceId }}</strong
              ><small>{{ check.message }}</small></span
            >
          </li>
          <li v-if="!requiredHealth.length">No required health checks declared.</li>
        </ul>
      </article>

      <article v-if="quickActions.length" class="panel">
        <header class="panel-head">
          <div>
            <p>Trusted manifest</p>
            <h2>Quick actions</h2>
          </div>
        </header>
        <div class="quick-actions">
          <button
            v-for="action in quickActions"
            :key="action.id"
            type="button"
            :disabled="actionPending"
            :title="`${action.risk} action`"
            @click="emit('runAction', action)"
          >
            {{ action.name }}
          </button>
        </div>
      </article>
    </aside>
  </div>
</template>
