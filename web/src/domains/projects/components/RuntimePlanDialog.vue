<script setup lang="ts">
import { AlertTriangle, Check, X } from '@lucide/vue'
import AppDialog from '../../../app/components/AppDialog.vue'

import type { RuntimePlan } from '../../../api/generated/types.gen'

defineProps<{
  plan: RuntimePlan
  pending: boolean
}>()

defineEmits<{
  cancel: []
  confirm: []
}>()

function commandLine(executable: string, arguments_: string[]) {
  return [executable, ...arguments_]
    .map((part) => (/^[\w./:=@-]+$/.test(part) ? part : JSON.stringify(part)))
    .join(' ')
}
</script>

<template>
  <AppDialog class="runtime-plan-dialog" labelledby="runtime-plan-title" @close="$emit('cancel')">
    <header>
      <div>
        <p>Lifecycle preview</p>
        <h2 id="runtime-plan-title">{{ plan.summary }}</h2>
      </div>
      <button type="button" aria-label="Close lifecycle preview" @click="$emit('cancel')">
        <X :size="17" aria-hidden="true" />
      </button>
    </header>

    <div class="plan-risk" :class="`plan-risk--${plan.risk}`">
      <AlertTriangle v-if="plan.risk !== 'safe'" :size="17" aria-hidden="true" />
      <Check v-else :size="17" aria-hidden="true" />
      <span
        ><strong>{{ plan.risk }}</strong> · Review the exact effects before continuing.</span
      >
    </div>

    <dl class="plan-facts">
      <div>
        <dt>Runtime</dt>
        <dd>{{ plan.driver }}</dd>
      </div>
      <div>
        <dt>Action</dt>
        <dd>{{ plan.action }}</dd>
      </div>
      <div>
        <dt>Services</dt>
        <dd>{{ plan.services.join(', ') || 'All declared services' }}</dd>
      </div>
      <div v-if="plan.profiles?.length">
        <dt>Profiles</dt>
        <dd>{{ plan.profiles.join(', ') }}</dd>
      </div>
    </dl>

    <div class="plan-section">
      <h3>Expected effects</h3>
      <ul>
        <li v-for="effect in plan.effects" :key="effect">{{ effect }}</li>
      </ul>
    </div>

    <div class="plan-section">
      <h3>Commands</h3>
      <div
        v-for="command in plan.commands"
        :key="commandLine(command.executable, command.arguments)"
        class="plan-command"
      >
        <code>{{ commandLine(command.executable, command.arguments) }}</code>
        <small>{{ command.workingDirectory }}</small>
      </div>
    </div>

    <footer>
      <button type="button" :disabled="pending" @click="$emit('cancel')">Cancel</button>
      <button
        class="primary"
        :class="{ danger: plan.risk === 'destructive' }"
        type="button"
        :disabled="pending"
        @click="$emit('confirm')"
      >
        {{ pending ? 'Queuing…' : `Confirm ${plan.action}` }}
      </button>
    </footer>
  </AppDialog>
</template>

<style scoped>
.runtime-plan-dialog {
  width: min(680px, 100%);
  max-height: min(820px, calc(100vh - 48px));
  overflow: auto;
  padding: 18px;
  border: 1px solid var(--border);
  border-radius: 13px;
  background: var(--panel);
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.5);
}
.runtime-plan-dialog > header,
.runtime-plan-dialog > footer,
.plan-risk {
  display: flex;
  align-items: center;
}
.runtime-plan-dialog > header,
.runtime-plan-dialog > footer {
  justify-content: space-between;
  gap: 12px;
}
.runtime-plan-dialog header p {
  margin: 0;
  color: var(--accent);
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}
.runtime-plan-dialog h2 {
  margin: 4px 0 0;
  font-size: 18px;
}
.runtime-plan-dialog button {
  display: inline-flex;
  min-height: 35px;
  align-items: center;
  justify-content: center;
  padding: 7px 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel-2);
  color: var(--text);
}
.runtime-plan-dialog .primary {
  border-color: color-mix(in srgb, var(--accent) 50%, transparent);
  background: var(--accent);
  color: var(--on-accent);
  font-weight: 800;
}
.runtime-plan-dialog .primary.danger {
  border-color: color-mix(in srgb, var(--red) 55%, transparent);
  background: var(--red);
}
.plan-risk {
  gap: 8px;
  margin: 15px 0;
  padding: 10px;
  border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
  border-radius: 8px;
  color: var(--accent);
  text-transform: capitalize;
}
.plan-risk--caution {
  border-color: color-mix(in srgb, var(--yellow) 25%, transparent);
  color: var(--yellow);
}
.plan-risk--destructive {
  border-color: color-mix(in srgb, var(--red) 30%, transparent);
  color: var(--red);
}
.plan-facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 7px;
}
.plan-facts > div,
.plan-command {
  padding: 9px;
  border-radius: 7px;
  background: var(--panel);
}
.plan-facts dt {
  color: var(--soft);
}
.plan-facts dd {
  margin: 3px 0 0;
  color: var(--text);
}
.plan-section {
  margin-top: 15px;
}
.plan-section h3 {
  margin: 0 0 8px;
  font-size: 12px;
}
.plan-section ul {
  margin: 0;
  padding-left: 20px;
  color: var(--muted);
}
.plan-command {
  display: grid;
  gap: 5px;
  margin-top: 6px;
}
.plan-command code {
  overflow-wrap: anywhere;
  color: var(--text);
}
.plan-command small {
  color: var(--soft);
}
.runtime-plan-dialog > footer {
  margin-top: 18px;
  justify-content: flex-end;
}

@media (max-width: 600px) {
  .plan-facts {
    grid-template-columns: 1fr;
  }
}
</style>
