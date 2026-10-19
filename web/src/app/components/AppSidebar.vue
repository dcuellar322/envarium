<script setup lang="ts">
import { FolderSearch, LayoutDashboard, MonitorCog, Plug, Users } from '@lucide/vue'
import { RouterLink } from 'vue-router'

import { useProjectCatalog } from '../../domains/projects/composables/useProjectCatalog'
import { useHostObservation } from '../../domains/system/composables/useHostObservation'
import { useSystemInfo } from '../../domains/system/composables/useSystemInfo'
import BrandIcon from './BrandIcon.vue'

defineProps<{ connection: 'connecting' | 'connected' | 'disconnected' }>()
const projects = useProjectCatalog()
const host = useHostObservation()
const system = useSystemInfo()

const navigation = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/projects', label: 'Projects', brandIcon: 'projects', count: 'projects' },
  { to: '/ports', label: 'Ports', brandIcon: 'ports' },
  { to: '/resources', label: 'Resources', brandIcon: 'resources' },
  { to: '/logs', label: 'Logs', brandIcon: 'logs' },
  { to: '/discovery', label: 'Discovery', icon: FolderSearch },
] as const
const tools = [
  { to: '/workspaces', label: 'Workspaces', brandIcon: 'workspaces' },
  { to: '/fleet', label: 'Machines', icon: MonitorCog },
  { to: '/team', label: 'Team', icon: Users },
  { to: '/agents', label: 'Agents', brandIcon: 'agents' },
  { to: '/plugins', label: 'Plugins', icon: Plug },
  { to: '/settings', label: 'Settings', brandIcon: 'settings' },
] as const
</script>

<template>
  <aside class="sidebar" aria-label="Primary navigation">
    <RouterLink class="brand" to="/" aria-label="Envarium dashboard">
      <img class="brand-mark" src="/brand/mark.svg" alt="" width="40" height="40" />
      <img
        class="brand-mark brand-mark--contrast"
        src="/brand/mark-contrast.svg"
        alt=""
        width="40"
        height="40"
      />
      <span class="brand-copy"
        ><img class="brand-wordmark" src="/brand/logo.svg" alt="" width="610" height="144" /><img
          class="brand-contrast"
          src="/brand/logo-contrast.svg"
          alt=""
          width="610"
          height="144"
        /><small>Local development control</small></span
      >
    </RouterLink>
    <p class="nav-label">Command center</p>
    <nav>
      <RouterLink
        v-for="item in navigation"
        :key="item.to"
        class="nav-item"
        :to="item.to"
        :aria-label="item.label"
      >
        <BrandIcon v-if="'brandIcon' in item" :name="item.brandIcon" class="nav-icon" />
        <component
          :is="item.icon"
          v-else
          class="nav-icon"
          :size="18"
          :stroke-width="1.8"
          aria-hidden="true"
        />
        <span class="nav-copy">{{ item.label }}</span>
        <span v-if="'count' in item && item.count === 'projects'" class="count">{{
          projects.data.value?.length ?? '—'
        }}</span>
      </RouterLink>
    </nav>
    <p class="nav-label nav-label--tools">Tools</p>
    <nav>
      <RouterLink
        v-for="item in tools"
        :key="item.to"
        class="nav-item"
        :to="item.to"
        :aria-label="item.label"
      >
        <BrandIcon v-if="'brandIcon' in item" :name="item.brandIcon" class="nav-icon" />
        <component
          :is="item.icon"
          v-else
          class="nav-icon"
          :size="18"
          :stroke-width="1.8"
          aria-hidden="true"
        />
        <span class="nav-copy">{{ item.label }}</span>
      </RouterLink>
    </nav>
    <div class="sidebar-footer">
      <div class="agent-card">
        <div>
          <span
            class="connection-dot"
            :class="`connection-dot--${connection}`"
            aria-hidden="true"
          ></span
          ><strong>Daemon {{ connection }}</strong
          ><span>{{ system.data.value?.version ?? '—' }}</span>
        </div>
        <p>
          Docker
          {{ host.data.value?.docker.connected ? 'connected' : 'unavailable' }}
          · {{ projects.data.value?.length ?? 0 }}
          {{ projects.data.value?.length === 1 ? 'project' : 'projects' }}
          indexed
        </p>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  width: 230px;
  height: 100vh;
  padding: 22px 16px;
  border-right: 1px solid var(--border);
  background: color-mix(in srgb, var(--panel) 86%, transparent);
  backdrop-filter: blur(18px);
  z-index: 30;
}
.brand {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 0 9px 22px;
  color: var(--text);
  text-decoration: none;
}
.brand-mark {
  display: none;
  width: 38px;
  height: 38px;
}
.brand-copy {
  display: grid;
  gap: 2px;
}
.brand-wordmark,
.brand-contrast {
  display: block;
  width: 180px;
  height: auto;
  margin-left: -5px;
}
.brand-contrast {
  display: none;
}
:global(html[data-theme='high-contrast'] .sidebar .brand-wordmark) {
  display: none;
}
:global(html[data-theme='high-contrast'] .sidebar .brand-contrast) {
  display: block;
}
.brand-copy small {
  color: var(--muted);
  font-size: 11px;
}
.nav-label {
  margin: 0;
  padding: 16px 11px 8px;
  color: var(--soft);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.13em;
  text-transform: uppercase;
}
.nav-label--tools {
  padding-top: 22px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  margin: 2px 0;
  padding: 10px 11px;
  border-radius: 9px;
  color: var(--muted);
  text-decoration: none;
  transition: 0.18s ease;
}
.nav-item:hover {
  color: var(--text);
  background: rgba(255, 255, 255, 0.04);
}
.nav-item.router-link-exact-active {
  color: var(--text);
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--accent) 18%, transparent),
    color-mix(in srgb, var(--accent) 5.5%, transparent)
  );
  box-shadow: inset 2px 0 var(--accent);
}
.nav-icon {
  flex: 0 0 20px;
  color: var(--soft);
}
.router-link-exact-active .nav-icon {
  color: var(--accent);
}
.count {
  margin-left: auto;
  padding: 2px 6px;
  border: 1px solid var(--border);
  border-radius: 99px;
  color: var(--soft);
  font-size: 10px;
}
.sidebar-footer {
  margin-top: auto;
}
.agent-card {
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--panel);
}
.agent-card > div {
  display: flex;
  align-items: center;
  font-size: 11px;
}
.agent-card > div > span:last-child {
  margin-left: auto;
  color: var(--soft);
}
.agent-card p {
  margin: 9px 0 0;
  color: var(--soft);
  font-size: 10px;
  line-height: 1.35;
}
.connection-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 7px;
  border-radius: 50%;
  background: var(--yellow);
}
.connection-dot--connected {
  background: var(--green);
  box-shadow: 0 0 12px color-mix(in srgb, var(--green) 60%, transparent);
}
.connection-dot--disconnected {
  background: var(--red);
}
@media (max-width: 760px) {
  .sidebar {
    width: 72px;
    padding: 18px 10px;
  }
  .brand {
    justify-content: center;
    padding-inline: 0;
  }
  .brand-mark {
    display: block;
  }
  .brand-mark--contrast {
    display: none;
  }
  :global(html[data-theme='high-contrast'] .sidebar .brand-mark:not(.brand-mark--contrast)) {
    display: none;
  }
  :global(html[data-theme='high-contrast'] .sidebar .brand-mark--contrast) {
    display: block;
  }
  .brand-copy,
  .nav-copy,
  .nav-label,
  .count,
  .agent-card {
    display: none;
  }
  .nav-item {
    justify-content: center;
  }
  .nav-icon {
    width: auto;
  }
}
</style>
