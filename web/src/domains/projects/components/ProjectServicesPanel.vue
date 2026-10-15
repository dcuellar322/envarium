<script setup lang="ts">
import { Ban, Pause, Play, RefreshCw, Rocket, Square } from '@lucide/vue'

import type {
  RuntimeAction,
  RuntimeMetricSample,
  RuntimeObservation,
  RuntimeServiceObservation,
} from '../../../api/generated/types.gen'
import { formatBytes, stateLabel } from '../../../lib/format'
import type { RuntimeSelection } from '../api'

const props = defineProps<{
  runtime?: RuntimeObservation
  pending: boolean
  metrics: RuntimeMetricSample[]
  actionPending: boolean
}>()

const emit = defineEmits<{
  lifecycle: [action: RuntimeAction, selection: RuntimeSelection]
}>()

function metric(serviceId: string) {
  return props.metrics.find((item) => item.serviceId === serviceId)
}

function running(service: RuntimeServiceObservation) {
  return ['running', 'starting', 'degraded'].includes(service.state)
}

function lifecycle(action: RuntimeAction, serviceId: string) {
  emit('lifecycle', action, { services: [serviceId] })
}
</script>

<template>
  <article class="panel services-panel">
    <header class="panel-head">
      <div>
        <p>Runtime</p>
        <h2>Services</h2>
      </div>
      <span>{{ runtime?.driver ?? '—' }} · {{ runtime?.origin ?? 'unobserved' }}</span>
    </header>
    <div v-if="pending" class="panel-state">Observing runtime…</div>
    <div v-else-if="runtime?.services.length" class="service-table">
      <div class="service-row service-row--head">
        <span>Service</span><span>Status</span><span>Health</span><span>Port</span
        ><span>Resource</span><span class="service-actions-heading">Controls</span>
      </div>
      <div v-for="service in runtime.services" :key="service.id" class="service-entry">
        <div class="service-row">
          <strong>{{ service.id }}</strong>
          <span
            ><i class="dot" :class="{ online: service.state === 'running' }"></i
            >{{ stateLabel(service.state) }}</span
          >
          <span>{{ service.health }}</span>
          <span>{{
            service.ports
              .map((port) => port.hostPort)
              .filter(Boolean)
              .join(', ') || '—'
          }}</span>
          <span>
            <template v-if="metric(service.id)?.memoryAvailable">
              {{ formatBytes(metric(service.id)!.memoryBytes) }}
            </template>
            <template v-else>—</template>
          </span>
          <span class="service-actions">
            <button
              v-if="!running(service) && service.state !== 'paused'"
              type="button"
              :disabled="actionPending"
              :aria-label="`Start ${service.id}`"
              title="Start service"
              @click="lifecycle('start', service.id)"
            >
              <Play :size="14" aria-hidden="true" />
            </button>
            <button
              v-if="running(service)"
              type="button"
              :disabled="actionPending"
              :aria-label="`Restart ${service.id}`"
              title="Restart service"
              @click="lifecycle('restart', service.id)"
            >
              <RefreshCw :size="14" aria-hidden="true" />
            </button>
            <button
              type="button"
              :disabled="actionPending"
              :aria-label="`Deploy ${service.id}`"
              title="Deploy current definition"
              @click="lifecycle('deploy', service.id)"
            >
              <Rocket :size="14" aria-hidden="true" />
            </button>
            <button
              v-if="runtime.driver === 'compose' && running(service)"
              type="button"
              :disabled="actionPending"
              :aria-label="`Pause ${service.id}`"
              title="Pause service"
              @click="lifecycle('pause', service.id)"
            >
              <Pause :size="14" aria-hidden="true" />
            </button>
            <button
              v-if="runtime.driver === 'compose' && service.state === 'paused'"
              type="button"
              :disabled="actionPending"
              :aria-label="`Resume ${service.id}`"
              title="Resume service"
              @click="lifecycle('unpause', service.id)"
            >
              <Play :size="14" aria-hidden="true" />
            </button>
            <button
              v-if="running(service) || service.state === 'paused'"
              type="button"
              :disabled="actionPending"
              :aria-label="`Stop ${service.id}`"
              title="Stop service"
              @click="lifecycle('stop', service.id)"
            >
              <Square :size="13" fill="currentColor" aria-hidden="true" />
            </button>
            <button
              v-if="running(service) || service.state === 'paused'"
              class="danger-text"
              type="button"
              :disabled="actionPending"
              :aria-label="`Force stop ${service.id}`"
              title="Force-stop service"
              @click="lifecycle('kill', service.id)"
            >
              <Ban :size="14" aria-hidden="true" />
            </button>
          </span>
        </div>
        <details class="service-inspector">
          <summary>Inspect {{ service.id }}</summary>
          <dl>
            <div>
              <dt>Runtime name</dt>
              <dd>{{ service.runtimeName }}</dd>
            </div>
            <template v-if="service.container">
              <div>
                <dt>Container</dt>
                <dd>
                  <code>{{ service.container.id.slice(0, 12) }}</code>
                </dd>
              </div>
              <div>
                <dt>Image</dt>
                <dd>{{ service.container.image }}</dd>
              </div>
              <div>
                <dt>Restarts</dt>
                <dd>{{ service.container.restartCount }}</dd>
              </div>
              <div>
                <dt>Exit</dt>
                <dd>{{ service.container.exitCode ?? '—' }}</dd>
              </div>
            </template>
            <template v-else-if="service.process">
              <div>
                <dt>PID</dt>
                <dd>{{ service.process.pid || '—' }}</dd>
              </div>
              <div>
                <dt>Executable</dt>
                <dd>
                  <code>{{ service.process.executable }}</code>
                </dd>
              </div>
              <div>
                <dt>Working directory</dt>
                <dd>{{ service.process.workingDirectory || '—' }}</dd>
              </div>
              <div>
                <dt>Restarts</dt>
                <dd>{{ service.process.restartCount }}</dd>
              </div>
            </template>
            <div>
              <dt>CPU</dt>
              <dd>
                {{
                  metric(service.id)?.cpuAvailable
                    ? `${metric(service.id)!.cpuPercent.toFixed(1)}%`
                    : '—'
                }}
              </dd>
            </div>
            <div>
              <dt>Observed</dt>
              <dd>{{ new Date(service.observedAt).toLocaleString() }}</dd>
            </div>
          </dl>
        </details>
      </div>
    </div>
    <p v-else class="panel-state">No runtime services are currently observed.</p>
  </article>
</template>
