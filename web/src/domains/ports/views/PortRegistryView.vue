<script setup lang="ts">
import { useMutation, useQuery } from '@tanstack/vue-query'
import { computed, ref } from 'vue'
import PortTopology from '../components/PortTopology.vue'
import { portTopology } from '../topology'

import { loadDaemonSettings } from '../../system/settingsApi'
import { loadPortRegistry, suggestPort } from '../api'

const registry = useQuery({
  queryKey: ['ports'],
  queryFn: loadPortRegistry,
  refetchInterval: 10_000,
  shallow: true,
})
const settings = useQuery({ queryKey: ['daemon-settings'], queryFn: loadDaemonSettings })
const preferred = computed(
  () =>
    settings.data.value?.settings.ports ?? { rangeStart: 15_000, rangeEnd: 19_999, excluded: [] },
)
const suggestion = useMutation({
  mutationFn: () =>
    suggestPort(preferred.value.rangeStart, preferred.value.rangeEnd, preferred.value.excluded),
})
const facts = computed(() => registry.data.value?.facts ?? [])
const conflicts = computed(() => registry.data.value?.conflicts ?? [])
const conflictingFacts = computed(
  () => new Set(conflicts.value.flatMap((conflict) => conflict.facts.map((fact) => fact.id))),
)
const configured = ref(false)
const host = ref(false)
const search = ref('')
const applications = computed(() =>
  portTopology(facts.value, conflictingFacts.value, {
    configured: configured.value,
    host: host.value,
    search: search.value,
  }),
)
const managedPorts = computed(
  () =>
    new Set(
      facts.value
        .filter((fact) => fact.projectId && fact.kind === 'binding')
        .map((fact) => `${fact.port}/${fact.protocol}`),
    ).size,
)
</script>

<template>
  <section class="ports-view" aria-labelledby="ports-title">
    <header class="page-heading">
      <div>
        <p class="eyebrow">Host intelligence</p>
        <h1 id="ports-title">Port registry</h1>
        <p>Follow each application through its services to the host ports it is using.</p>
      </div>
      <button type="button" :disabled="suggestion.isPending.value" @click="suggestion.mutate()">
        {{ suggestion.isPending.value ? 'Checking…' : 'Find next free port' }}
      </button>
    </header>

    <div v-if="registry.isPending.value" class="state-panel" aria-live="polite">
      Inspecting manifests, runtimes, and OS listeners…
    </div>
    <div v-else-if="registry.isError.value" class="state-panel state-panel--error" role="alert">
      <strong>Port registry unavailable</strong
      ><button type="button" @click="registry.refetch()">Retry</button>
    </div>
    <template v-else>
      <div class="summary-grid">
        <article>
          <span>Active endpoints</span><strong>{{ managedPorts }}</strong
          ><small>Published by managed services</small>
        </article>
        <article>
          <span>Current facts</span><strong>{{ facts.length }}</strong
          ><small>Declarations, leases, listeners</small>
        </article>
        <article :class="{ danger: conflicts.length }">
          <span>Conflicts</span><strong>{{ conflicts.length }}</strong
          ><small>{{ conflicts.length ? 'Action recommended' : 'No overlap detected' }}</small>
        </article>
      </div>

      <p v-if="suggestion.data.value" class="suggestion" role="status">
        <strong>{{ suggestion.data.value.port }}</strong
        >/tcp is free in the preferred {{ preferred.rangeStart }}–{{ preferred.rangeEnd }} range.
      </p>
      <p v-else-if="suggestion.isError.value" class="message message--error" role="alert">
        No free port could be suggested. Refresh the registry and try again.
      </p>
      <div v-for="warning in registry.data.value?.warnings ?? []" :key="warning" class="warning">
        Partial data: {{ warning }}
      </div>

      <article class="panel">
        <div class="panel-heading">
          <div>
            <p class="eyebrow">Visual map</p>
            <h2>Application connections</h2>
          </div>
          <span>Live refresh · 10s</span>
        </div>
        <div class="topology-controls">
          <label class="port-search"
            >Filter connections<input
              v-model="search"
              type="search"
              placeholder="Application, service, or port"
          /></label>
          <label><input v-model="configured" type="checkbox" /> Include configured ports</label>
          <label><input v-model="host" type="checkbox" /> Include other host listeners</label>
        </div>
        <PortTopology v-if="applications.length" :applications="applications" />
        <p v-else class="empty">
          {{
            search
              ? 'No connections match this filter.'
              : 'No managed port bindings are currently observed. Enable configured ports to see declarations and reservations.'
          }}
        </p>
      </article>

      <section v-if="conflicts.length" class="panel conflict-summary" aria-label="Port conflicts">
        <h2>Conflicts need attention</h2>
        <ul>
          <li v-for="conflict in conflicts" :key="conflict.id">
            <strong>{{ conflict.port }}</strong> · {{ conflict.summary }}
          </li>
        </ul>
      </section>
      <details class="port-evidence">
        <summary>All port facts · {{ facts.length }} observations</summary>
        <article class="panel">
          <div class="panel-heading">
            <div>
              <p class="eyebrow">Evidence</p>
              <h2>All port facts</h2>
            </div>
            <span>{{
              registry.data.value?.observedAt
                ? new Date(registry.data.value.observedAt).toLocaleTimeString()
                : '—'
            }}</span>
          </div>
          <div v-if="facts.length" class="port-table" role="table" aria-label="All port facts">
            <div class="port-row port-row--header" role="row">
              <span role="columnheader">Port</span><span role="columnheader">Project</span
              ><span role="columnheader">Service</span><span role="columnheader">Source</span
              ><span role="columnheader">State</span>
            </div>
            <div v-for="fact in facts" :key="`row-${fact.id}`" class="port-row" role="row">
              <strong role="cell" :class="{ conflict: conflictingFacts.has(fact.id) }"
                >{{ fact.port }}/{{ fact.protocol }}</strong
              ><span role="cell">{{ fact.projectName ?? 'Unknown process' }}</span
              ><span role="cell">{{ fact.serviceId ?? '—' }}</span
              ><span role="cell" :title="fact.evidence">{{ fact.source }}</span
              ><em role="cell">{{ conflictingFacts.has(fact.id) ? 'conflict' : fact.kind }}</em>
            </div>
          </div>
          <p v-else class="empty">Register and trust a project to protect its declared ports.</p>
        </article>
      </details>
    </template>
  </section>
</template>

<style scoped>
.ports-view {
  max-width: 1240px;
  padding: 42px 28px;
  margin: 0 auto;
}
.page-heading,
.panel-heading {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: flex-start;
}
.page-heading {
  margin-bottom: 24px;
}
.eyebrow {
  margin: 0;
  color: var(--accent);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  font-weight: 800;
}
h1 {
  margin: 5px 0 8px;
  font-size: 42px;
}
.page-heading p:last-child {
  margin: 0;
  color: var(--muted);
}
button {
  padding: 10px 14px;
  border: 0;
  border-radius: 9px;
  background: var(--accent);
  color: var(--on-accent);
  font-weight: 800;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 15px;
}
.summary-grid article,
.panel {
  padding: 18px;
  border: 1px solid var(--border);
  border-radius: 13px;
  background: var(--panel);
}
.summary-grid article {
  display: grid;
  gap: 5px;
}
.summary-grid span,
.summary-grid small,
.panel-heading > span {
  color: var(--muted);
}
.summary-grid strong {
  font-size: 28px;
}
.summary-grid .danger strong,
.conflict {
  color: var(--red);
}
.suggestion,
.warning,
.state-panel {
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: color-mix(in srgb, var(--green) 7%, transparent);
  color: var(--green);
}
.warning {
  margin: 8px 0;
  background: color-mix(in srgb, var(--yellow) 7%, transparent);
  color: var(--yellow);
}
.state-panel--error {
  color: var(--red);
}
.panel {
  margin-top: 15px;
}
.panel-heading {
  align-items: center;
  margin-bottom: 14px;
}
.panel-heading h2 {
  margin: 4px 0 0;
}
.port-table {
  border: 1px solid var(--border);
  border-radius: 9px;
  overflow: hidden;
}
.port-row {
  display: grid;
  grid-template-columns: 100px 1.3fr 1fr 110px 100px;
  gap: 12px;
  padding: 10px 12px;
  border-top: 1px solid var(--border);
  align-items: center;
}
.port-row:first-child {
  border-top: 0;
}
.port-row--header {
  background: var(--panel);
  color: var(--soft);
  font-size: 10px;
  text-transform: uppercase;
}
.port-row span {
  color: var(--muted);
}
.port-row em {
  font-style: normal;
  color: var(--soft);
}
.empty {
  padding: 18px;
  color: var(--muted);
  text-align: center;
}
@media (max-width: 800px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
  .page-heading {
    display: grid;
  }
  .port-row {
    grid-template-columns: 90px 1fr;
  }
  .port-row--header {
    display: none;
  }
}
.topology-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 20px;
  align-items: center;
  margin: 18px 0 22px;
}
.topology-controls label {
  display: flex;
  gap: 8px;
  align-items: center;
  color: var(--muted);
  font-size: 12px;
}
.topology-controls .port-search {
  display: grid;
  gap: 7px;
  flex: 1 1 260px;
}
.port-search input {
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  min-width: 0;
}
.port-evidence {
  margin-top: 22px;
}
.port-evidence summary {
  cursor: pointer;
  padding: 12px 0;
  color: var(--soft);
  font-size: 13px;
}
.conflict-summary h2 {
  font-size: 15px;
  color: var(--red);
  margin-top: 0;
}
.conflict-summary ul {
  padding-left: 20px;
  color: var(--muted);
}
.conflict-summary li {
  margin: 8px 0;
}
</style>
