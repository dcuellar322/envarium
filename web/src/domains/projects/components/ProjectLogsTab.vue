<script setup lang="ts">
import { ArrowDownToLine, Download, Pause, Play } from '@lucide/vue'
import { computed, nextTick, ref, watch } from 'vue'

import type { RuntimeLogEntry, RuntimeMetricSample } from '../../../api/generated/types.gen'
import { formatBytes } from '../../../lib/format'
import { exportProjectLogText } from '../api'

const props = defineProps<{
  projectId: string
  entries: RuntimeLogEntry[]
  connection: string
  services: string[]
  metrics: RuntimeMetricSample[]
}>()

const logContainer = ref<HTMLElement>()
const followNewest = ref(true)
const paused = ref(false)
const frozenEntries = ref<RuntimeLogEntry[]>([])
const search = ref('')
const serviceFilter = ref('all')
const levelFilter = ref('all')
const streamFilter = ref('all')
const downloadPending = ref(false)
const downloadError = ref('')

const sourceEntries = computed(() => (paused.value ? frozenEntries.value : props.entries))
const serviceOptions = computed(() =>
  [...new Set([...props.services, ...props.entries.map((entry) => entry.serviceId)])].sort(),
)
const levelOptions = computed(() =>
  [...new Set(props.entries.map((entry) => entry.level).filter(Boolean))].sort(),
)
const visibleEntries = computed(() => {
  const query = search.value.trim().toLowerCase()
  return sourceEntries.value.filter(
    (entry) =>
      (!query || `${entry.message} ${entry.serviceId}`.toLowerCase().includes(query)) &&
      (serviceFilter.value === 'all' || entry.serviceId === serviceFilter.value) &&
      (levelFilter.value === 'all' || entry.level === levelFilter.value) &&
      (streamFilter.value === 'all' || entry.stream === streamFilter.value),
  )
})
const cpu = computed(() =>
  props.metrics.reduce((total, metric) => total + (metric.cpuAvailable ? metric.cpuPercent : 0), 0),
)
const memory = computed(() =>
  props.metrics.reduce(
    (total, metric) => total + (metric.memoryAvailable ? metric.memoryBytes : 0),
    0,
  ),
)

function updateFollowState() {
  const element = logContainer.value
  if (!element) return
  followNewest.value = element.scrollHeight - element.scrollTop - element.clientHeight < 48
}

async function scrollToNewest() {
  followNewest.value = true
  await nextTick()
  if (logContainer.value) logContainer.value.scrollTop = logContainer.value.scrollHeight
}

function togglePause() {
  if (!paused.value) frozenEntries.value = [...props.entries]
  paused.value = !paused.value
}

function structuredFields(entry: RuntimeLogEntry) {
  const fields: Record<string, unknown> = { ...entry.attributes }
  const message = entry.message.trim()
  if (message.startsWith('{') && message.endsWith('}')) {
    try {
      const parsed = JSON.parse(message) as unknown
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed))
        Object.assign(fields, parsed)
    } catch {
      // The original line remains authoritative when JSON-shaped output is malformed.
    }
  }
  return Object.entries(fields)
}

async function downloadLogs(format: 'plain' | 'ndjson') {
  downloadPending.value = true
  downloadError.value = ''
  try {
    const content = await exportProjectLogText(
      props.projectId,
      serviceFilter.value === 'all' ? '' : serviceFilter.value,
      format,
    )
    const blob = new window.Blob([content], {
      type: format === 'ndjson' ? 'application/x-ndjson' : 'text/plain;charset=utf-8',
    })
    const href = window.URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = href
    anchor.download = `envarium-${props.projectId}-logs.${format === 'ndjson' ? 'ndjson' : 'log'}`
    anchor.click()
    window.URL.revokeObjectURL(href)
  } catch (cause) {
    downloadError.value = cause instanceof Error ? cause.message : 'Log export failed.'
  } finally {
    downloadPending.value = false
  }
}

watch(
  () => props.entries.at(-1)?.sequence,
  async () => {
    if (followNewest.value && !paused.value) await scrollToNewest()
  },
  { immediate: true },
)
</script>

<template>
  <article class="panel project-log-panel">
    <header class="panel-head log-panel-head">
      <div>
        <p>{{ visibleEntries.length }} of {{ sourceEntries.length }} bounded entries</p>
        <h2>Project logs</h2>
      </div>
      <div class="log-resource-pills" aria-label="Current project resources">
        <span
          >CPU <strong>{{ cpu.toFixed(1) }}%</strong></span
        >
        <span
          >Memory <strong>{{ formatBytes(memory) }}</strong></span
        >
        <span class="stream-state"
          ><i :class="{ online: connection === 'connected' }"></i>{{ connection }}</span
        >
      </div>
    </header>

    <div class="project-log-filters" role="search">
      <label
        ><span class="sr-only">Search project logs</span
        ><input v-model="search" placeholder="Search logs…"
      /></label>
      <label
        ><span class="sr-only">Service</span
        ><select v-model="serviceFilter">
          <option value="all">All services</option>
          <option v-for="service in serviceOptions" :key="service" :value="service">
            {{ service }}
          </option>
        </select></label
      >
      <label
        ><span class="sr-only">Level</span
        ><select v-model="levelFilter">
          <option value="all">All levels</option>
          <option v-for="level in levelOptions" :key="level" :value="level">{{ level }}</option>
        </select></label
      >
      <label
        ><span class="sr-only">Stream</span
        ><select v-model="streamFilter">
          <option value="all">stdout + stderr</option>
          <option value="stdout">stdout</option>
          <option value="stderr">stderr</option>
        </select></label
      >
      <button type="button" :aria-pressed="paused" @click="togglePause">
        <Play v-if="paused" :size="14" aria-hidden="true" /><Pause
          v-else
          :size="14"
          aria-hidden="true"
        />
        {{ paused ? 'Resume' : 'Pause' }}
      </button>
      <button type="button" :aria-pressed="followNewest" @click="scrollToNewest">
        <ArrowDownToLine :size="14" aria-hidden="true" />{{ followNewest ? 'Following' : 'Newest' }}
      </button>
      <details class="download-menu">
        <summary :aria-disabled="downloadPending">
          <Download :size="14" aria-hidden="true" />Download
        </summary>
        <div>
          <button type="button" :disabled="downloadPending" @click="downloadLogs('plain')">
            Plain text</button
          ><button type="button" :disabled="downloadPending" @click="downloadLogs('ndjson')">
            NDJSON
          </button>
        </div>
      </details>
    </div>

    <p v-if="paused" class="log-notice" role="status">
      Display paused at {{ sourceEntries.length }} entries. Collection continues in the background.
    </p>
    <p v-if="downloadError" class="message message--error" role="alert">{{ downloadError }}</p>
    <div
      v-if="visibleEntries.length"
      ref="logContainer"
      class="log-lines log-lines--full log-lines--structured"
      role="log"
      aria-live="off"
      @scroll.passive="updateFollowState"
    >
      <div v-for="entry in visibleEntries" :key="entry.sequence" class="structured-log-row">
        <time :datetime="entry.timestamp">{{
          new Date(entry.timestamp).toLocaleTimeString()
        }}</time>
        <span>{{ entry.serviceId }}</span>
        <em :class="`level level--${entry.level}`">{{ entry.level }}</em>
        <pre :class="{ stderr: entry.stream === 'stderr' }">{{ entry.message }}</pre>
        <details v-if="structuredFields(entry).length" class="log-fields">
          <summary>{{ structuredFields(entry).length }} fields</summary>
          <dl>
            <div v-for="[key, value] in structuredFields(entry)" :key="key">
              <dt>{{ key }}</dt>
              <dd>{{ value }}</dd>
            </div>
          </dl>
        </details>
      </div>
    </div>
    <p v-else class="panel-state">No log entries match the current filters.</p>
  </article>
</template>
