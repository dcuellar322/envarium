<script setup lang="ts">
import BrandIcon from '../../../app/components/BrandIcon.vue'
import { ArrowRight, Play, Square } from '@lucide/vue'
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import type { RuntimeAction } from '../../../api/generated/types.gen'
import {
  formatBytes,
  isActiveState,
  projectInitials,
  stateLabel,
  tagTone,
} from '../../../lib/format'
import type { ProjectSnapshot } from '../api'

const props = defineProps<{ snapshot: ProjectSnapshot; pending?: boolean }>()
const emit = defineEmits<{
  runtime: [snapshot: ProjectSnapshot, action: RuntimeAction]
  open: [projectId: string]
}>()

const state = computed(() => props.snapshot.runtime?.state ?? 'unknown')
const trusted = computed(() => props.snapshot.project.trustState === 'trusted')
const dockerUnavailable = computed(
  () =>
    props.snapshot.runtime?.driver === 'compose' &&
    props.snapshot.runtime.engine?.connected === false,
)
const tone = computed(() => {
  if (dockerUnavailable.value || ['degraded', 'partially_running', 'failed'].includes(state.value))
    return 'degraded'
  if (isActiveState(state.value)) return 'running'
  return 'stopped'
})
const changeCount = computed(() => {
  const changes = props.snapshot.git?.changes
  return changes ? changes.staged + changes.modified + changes.untracked + changes.conflicted : 0
})
const endpoints = computed(
  () =>
    props.snapshot.actions?.actions.filter(
      (action) => action.type === 'browser.open' && action.target,
    ) ?? [],
)
const endpointSummary = computed(() => {
  const first = endpoints.value[0]?.target?.replace(/^https?:\/\//, '')
  if (!first) return 'Not declared'
  return endpoints.value.length > 1 ? `${first} +${endpoints.value.length - 1}` : first
})
const endpointTitle = computed(() =>
  endpoints.value.map((action) => `${action.name}: ${action.target}`).join('\n'),
)
const memory = computed(
  () => props.snapshot.metrics?.reduce((total, sample) => total + sample.memoryBytes, 0) ?? 0,
)
const cpu = computed(
  () => props.snapshot.metrics?.reduce((total, sample) => total + sample.cpuPercent, 0) ?? 0,
)
const memoryAvailable = computed(() =>
  props.snapshot.metrics?.every((sample) => sample.memoryAvailable),
)
const cpuAvailable = computed(() => props.snapshot.metrics?.every((sample) => sample.cpuAvailable))
const primaryAction = computed<RuntimeAction>(() => (isActiveState(state.value) ? 'stop' : 'start'))
const services = computed(() => props.snapshot.runtime?.services.slice(0, 4) ?? [])
</script>

<template>
  <article class="project-card" :aria-label="`${snapshot.project.displayName} project`">
    <div class="project-top">
      <div class="project-title-line">
        <div class="project-avatar" aria-hidden="true">
          {{ projectInitials(snapshot.project.displayName) }}
        </div>
        <div class="project-copy">
          <h3>{{ snapshot.project.displayName }}</h3>
          <span :title="snapshot.project.primaryLocation">{{
            snapshot.project.primaryLocation
          }}</span>
        </div>
      </div>
      <span class="status" :class="`status--${tone}`">
        <i aria-hidden="true"></i
        >{{
          !trusted
            ? 'Review required'
            : dockerUnavailable
              ? 'Docker unavailable'
              : stateLabel(state)
        }}
      </span>
    </div>

    <div class="project-meta">
      <div class="meta-box">
        <span>Git</span
        ><strong
          >{{ snapshot.git?.branch ?? (snapshot.git?.detached ? 'detached' : '—') }}
          ·
          <em :class="{ dirty: changeCount }">{{
            !snapshot.git
              ? 'unavailable'
              : !snapshot.git.repository
                ? 'not a repository'
                : changeCount
                  ? `${changeCount} changes`
                  : 'clean'
          }}</em></strong
        >
      </div>
      <div class="meta-box">
        <span>Endpoints</span><strong :title="endpointTitle">{{ endpointSummary }}</strong>
      </div>
      <div class="meta-box">
        <span>Resources</span
        ><strong>{{
          snapshot.metrics
            ? `${memoryAvailable ? formatBytes(memory) : '—'} · ${cpuAvailable ? cpu.toFixed(1) + '%' : '—'}`
            : '—'
        }}</strong>
      </div>
    </div>

    <div class="service-row">
      <span
        v-for="service in services"
        :key="service.id"
        class="service-chip"
        :class="`service-chip--${tagTone(service.id)}`"
        :title="`${service.id}: ${stateLabel(service.state)}`"
        >{{ service.id }}</span
      >
      <span v-if="!services.length" class="service-chip">{{
        snapshot.runtime?.driver ?? 'runtime pending'
      }}</span>
      <span v-if="snapshot.runtime && snapshot.runtime.services.length > 4" class="service-chip"
        >+{{ snapshot.runtime.services.length - 4 }}</span
      >
      <span v-if="snapshot.warnings.length" class="partial" :title="snapshot.warnings.join('. ')"
        >Partial data</span
      >
    </div>

    <div class="project-actions">
      <button
        v-if="trusted"
        type="button"
        class="button"
        :class="{
          'button--primary': primaryAction === 'start',
          'button--danger': primaryAction === 'stop',
        }"
        :disabled="pending"
        @click="emit('runtime', snapshot, primaryAction)"
      >
        <Play v-if="primaryAction === 'start'" :size="15" aria-hidden="true" />
        <Square v-else :size="14" fill="currentColor" aria-hidden="true" />
        {{ primaryAction === 'start' ? 'Start' : 'Stop' }}
      </button>
      <RouterLink
        v-else
        class="button button--primary"
        :to="{ name: 'discovery', query: { path: snapshot.project.primaryLocation } }"
      >
        Review project
      </RouterLink>
      <RouterLink
        class="button"
        :to="{
          name: 'project',
          params: { projectId: snapshot.project.id },
          query: { tab: 'logs' },
        }"
        @click="emit('open', snapshot.project.id)"
        ><BrandIcon name="logs" :size="15" aria-hidden="true" />Logs</RouterLink
      >
      <RouterLink
        class="button"
        :to="{
          name: 'project',
          params: { projectId: snapshot.project.id },
          query: { tab: 'terminal' },
        }"
        @click="emit('open', snapshot.project.id)"
      >
        <BrandIcon name="terminal" :size="15" aria-hidden="true" />Terminal
      </RouterLink>
      <RouterLink
        class="button open-detail"
        :to="{ name: 'project', params: { projectId: snapshot.project.id } }"
        @click="emit('open', snapshot.project.id)"
        >Open <ArrowRight :size="15" aria-hidden="true"
      /></RouterLink>
    </div>
  </article>
</template>

<style scoped>
.project-card {
  min-width: 0;
  contain: layout paint;
  border: 1px solid var(--border);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--panel) 98%, transparent),
    color-mix(in srgb, var(--bg) 98%, transparent)
  );
  border-radius: 15px;
  padding: 17px;
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.025);
  transition: 0.18s ease;
}
.project-card:hover {
  transform: translateY(-2px);
  border-color: var(--accent);
  box-shadow: 0 22px 70px rgba(0, 0, 0, 0.32);
}
.project-top,
.project-title-line,
.project-actions,
.service-row {
  display: flex;
  align-items: center;
}
.project-top {
  justify-content: space-between;
  gap: 15px;
  margin-bottom: 15px;
}
.project-title-line {
  gap: 11px;
  min-width: 0;
}
.project-avatar {
  flex-shrink: 0;
  width: 38px;
  height: 38px;
  border: 1px solid var(--border);
  border-radius: 11px;
  background: var(--panel-3);
  color: var(--accent);
  font-weight: 800;
  display: grid;
  place-items: center;
}
.project-copy {
  min-width: 0;
}
.project-copy h3 {
  overflow-wrap: anywhere;
  margin: 0 0 3px;
  font-size: 14px;
}
.project-copy span {
  display: block;
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--soft);
  font-size: 10px;
}
.status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border: 1px solid var(--border);
  border-radius: 99px;
  color: var(--muted);
  background: rgba(255, 255, 255, 0.025);
  font-size: 10px;
  white-space: nowrap;
  text-transform: capitalize;
}
.status i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}
.status--running {
  color: var(--green);
  border-color: color-mix(in srgb, var(--green) 28%, transparent);
  background: color-mix(in srgb, var(--green) 8%, transparent);
}
.status--degraded {
  color: var(--yellow);
  border-color: color-mix(in srgb, var(--yellow) 30%, transparent);
  background: color-mix(in srgb, var(--yellow) 8%, transparent);
}
.project-meta {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1fr;
  gap: 10px;
  margin: 12px 0 14px;
}
.meta-box {
  min-width: 0;
  padding: 9px;
  border: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.023);
}
.meta-box span {
  display: block;
  margin-bottom: 4px;
  color: var(--soft);
  font-size: 9px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}
.meta-box strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text);
  font-size: 11px;
  font-weight: 600;
}
.meta-box em {
  font-style: normal;
  color: var(--green);
}
.meta-box .dirty {
  color: var(--yellow);
}
.service-row {
  gap: 7px;
  min-height: 22px;
  margin-bottom: 13px;
  color: var(--muted);
  font-size: 11px;
  flex-wrap: wrap;
}
.service-chip {
  padding: 3px 6px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--panel-2);
}
.service-chip--blue {
  color: var(--accent);
  border-color: color-mix(in srgb, var(--accent) 28%, transparent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
}
.service-chip--purple {
  color: var(--accent-2);
  border-color: color-mix(in srgb, var(--accent-2) 30%, transparent);
  background: color-mix(in srgb, var(--accent-2) 10%, transparent);
}
.service-chip--green {
  color: var(--green);
  border-color: color-mix(in srgb, var(--green) 28%, transparent);
  background: color-mix(in srgb, var(--green) 9%, transparent);
}
.service-chip--orange {
  color: var(--yellow);
  border-color: color-mix(in srgb, var(--yellow) 30%, transparent);
  background: color-mix(in srgb, var(--yellow) 9%, transparent);
}
.service-chip--cyan {
  color: var(--cyan);
  border-color: color-mix(in srgb, var(--cyan) 28%, transparent);
  background: color-mix(in srgb, var(--cyan) 9%, transparent);
}
.partial {
  color: var(--yellow);
  font-size: 10px;
}
.project-actions {
  gap: 7px;
  padding-top: 13px;
  border-top: 1px solid rgba(255, 255, 255, 0.055);
}
.button {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  border: 1px solid var(--border);
  border-radius: 9px;
  background: var(--panel-2);
  color: var(--text);
  font-size: 11px;
  text-decoration: none;
}
.button:disabled {
  opacity: 0.45;
}
.button--primary {
  border: 0;
  background: var(--primary-gradient);
  color: var(--on-accent);
  font-weight: 750;
}
.button--danger {
  color: var(--red);
}
.open-detail {
  margin-left: auto;
  color: var(--accent);
}
@media (max-width: 700px) {
  .project-meta {
    grid-template-columns: 1fr;
  }
  .project-actions {
    flex-wrap: wrap;
  }
  .open-detail {
    margin-left: 0;
  }
  .project-top {
    align-items: flex-start;
  }
}
@media (max-width: 500px) {
  .project-top {
    flex-wrap: wrap;
    gap: 10px;
  }
  .project-title-line {
    width: 100%;
  }
  .project-copy {
    flex: 1;
  }
  .project-copy span {
    max-width: 100%;
  }
  .project-top > .status {
    margin-left: 49px;
  }
}
</style>
