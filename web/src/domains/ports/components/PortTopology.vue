<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { PortApplication } from '../topology'
defineProps<{ applications: PortApplication[] }>()
</script>

<template>
  <ul class="topology" aria-label="Applications, services, and host ports">
    <li v-for="app in applications" :key="app.id" class="topology-app">
      <div class="application-node">
        <span class="node-icon" aria-hidden="true">◈</span>
        <div>
          <h3>
            <RouterLink
              v-if="app.id !== 'host-listeners' && !app.id.startsWith('env-')"
              :to="{ name: 'project', params: { projectId: app.id } }"
              >{{ app.name }}</RouterLink
            ><template v-else>{{ app.name }}</template>
          </h3>
          <small
            >{{ app.services.length }}
            {{ app.services.length === 1 ? 'service' : 'services' }}</small
          >
        </div>
      </div>
      <ul class="service-nodes">
        <li v-for="service in app.services" :key="service.id" class="service-node">
          <h4>{{ service.name }}</h4>
          <ul class="endpoint-nodes">
            <li
              v-for="endpoint in service.endpoints"
              :key="endpoint.key"
              :class="{ conflict: endpoint.conflict }"
            >
              <div class="endpoint-title">
                <span
                  class="live-dot"
                  :class="{ configured: !endpoint.active }"
                  aria-hidden="true"
                ></span
                ><strong>{{ endpoint.port }}</strong
                ><small>/{{ endpoint.protocol }}</small
                ><span v-if="endpoint.conflict" class="conflict-label">Conflict</span>
              </div>
              <p>
                {{ endpoint.active ? 'Bound' : 'Configured'
                }}<span v-if="endpoint.targets.length"> → {{ endpoint.targets.join(', ') }}</span>
              </p>
              <code>{{ endpoint.hosts.join(' · ') }}</code>
            </li>
          </ul>
        </li>
      </ul>
    </li>
  </ul>
</template>

<style scoped>
.topology,
.service-nodes,
.endpoint-nodes {
  list-style: none;
  margin: 0;
  padding: 0;
}
.topology {
  display: grid;
  gap: 18px;
}
.topology-app {
  display: grid;
  grid-template-columns: minmax(150px, 200px) 1fr;
  align-items: start;
  padding: 18px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--panel);
  min-width: 0;
}
.application-node {
  display: flex;
  gap: 10px;
  padding: 12px 20px 12px 0;
  align-items: flex-start;
  min-width: 0;
}
.node-icon {
  color: var(--accent);
  font-size: 22px;
}
h3 {
  font-size: 14px;
  margin: 0 0 5px;
  overflow-wrap: anywhere;
}
small,
p,
code {
  color: var(--muted);
}
.application-node small {
  font-size: 11px;
}
.service-nodes {
  position: relative;
  padding-left: 24px;
  border-left: 1px solid var(--border);
  display: grid;
  gap: 18px;
  min-width: 0;
}
.service-node {
  position: relative;
  min-width: 0;
}
.service-node::before {
  content: '';
  position: absolute;
  width: 24px;
  height: 1px;
  background: var(--border);
  top: 10px;
  left: -25px;
}
h4 {
  margin: 0 0 10px;
  font-size: 12px;
  color: var(--soft);
}
.endpoint-nodes {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 10px;
}
.endpoint-nodes > li {
  border: 1px solid var(--border);
  border-radius: 9px;
  padding: 12px 14px;
  min-width: 0;
  background: var(--bg);
}
.endpoint-title {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.endpoint-title strong {
  color: var(--text);
  font: 700 18px var(--font-mono, monospace);
}
.endpoint-title small {
  font-size: 11px;
}
.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--green);
  margin-right: 2px;
}
.live-dot.configured {
  background: var(--muted);
}
p {
  font-size: 11px;
  margin: 8px 0 5px;
}
code {
  font-size: 10px;
  overflow-wrap: anywhere;
}
.endpoint-nodes > li.conflict {
  border-color: var(--red);
  background: color-mix(in srgb, var(--red) 5%, var(--bg));
}
.conflict-label {
  color: var(--red);
  font-size: 10px;
  margin-left: auto;
}
@media (max-width: 720px) {
  .topology-app {
    grid-template-columns: 1fr;
    padding: 14px;
  }
  .application-node {
    padding: 0 0 16px;
  }
  .service-nodes {
    margin-left: 10px;
    padding-left: 18px;
  }
  .service-node::before {
    width: 18px;
    left: -19px;
  }
  .endpoint-nodes {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  }
}
</style>
