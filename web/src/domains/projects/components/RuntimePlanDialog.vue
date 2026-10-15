<script setup lang="ts">
import { AlertTriangle, Check, X } from '@lucide/vue'

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
  <div class="modal-backdrop" @click.self="$emit('cancel')">
    <section
      class="runtime-plan-dialog"
      role="dialog"
      aria-modal="true"
      aria-labelledby="runtime-plan-title"
      @keydown.esc="$emit('cancel')"
    >
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
    </section>
  </div>
</template>
