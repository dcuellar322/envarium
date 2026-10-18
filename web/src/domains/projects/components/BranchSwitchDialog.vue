<script setup lang="ts">
import { ref, watch } from 'vue'
import AppDialog from '../../../app/components/AppDialog.vue'
import type { GitBranchSwitchPlan } from '../../../api/generated/types.gen'
const props = defineProps<{ plan: GitBranchSwitchPlan; pending: boolean }>()
const emit = defineEmits<{ cancel: []; confirm: [] }>()
const confirmed = ref(false)
watch(
  () => props.plan,
  () => {
    confirmed.value = false
  },
)
</script>

<template>
  <AppDialog class="branch-dialog" labelledby="branch-switch-title" @close="emit('cancel')">
    <header>
      <p>Checkout preview · mutating</p>
      <h2 id="branch-switch-title">Switch to {{ plan.toBranch }}</h2>
    </header>
    <dl>
      <div>
        <dt>From</dt>
        <dd>
          {{ plan.fromBranch }} <code>{{ plan.fromHead.slice(0, 12) }}</code>
        </dd>
      </div>
      <div>
        <dt>To</dt>
        <dd>
          {{ plan.toBranch }} <code>{{ plan.toHead.slice(0, 12) }}</code>
        </dd>
      </div>
      <div>
        <dt>Checkout</dt>
        <dd>
          <code>{{ plan.root }}</code>
        </dd>
      </div>
    </dl>
    <p>
      This updates the files in this clean, stopped checkout. The runtime stays stopped. After
      switching, review the new branch’s start plan before running it.
    </p>
    <label class="review-confirmation"
      ><input v-model="confirmed" type="checkbox" :disabled="pending" /> I reviewed this checkout
      change and will review the new runtime before starting it.</label
    >
    <footer>
      <button type="button" :disabled="pending" @click="emit('cancel')">Cancel</button
      ><button
        type="button"
        class="primary"
        :disabled="pending || !confirmed"
        @click="emit('confirm')"
      >
        {{ pending ? 'Queuing…' : 'Switch branch' }}
      </button>
    </footer>
  </AppDialog>
</template>

<style scoped>
.branch-dialog {
  width: min(560px, calc(100vw - 40px));
  max-height: calc(100dvh - 40px);
  overflow: auto;
  border: 1px solid var(--border);
  border-radius: 16px;
  background: var(--panel);
  color: var(--text);
  padding: 24px;
  box-shadow: 0 24px 80px #0008;
}
header p {
  margin: 0 0 8px;
  font-size: 10px;
  color: var(--yellow);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
h2 {
  margin: 0 0 22px;
  font-size: 23px;
  overflow-wrap: anywhere;
}
dl {
  display: grid;
  gap: 14px;
}
dl > div {
  display: grid;
  grid-template-columns: 70px 1fr;
  gap: 12px;
}
dt {
  color: var(--muted);
  font-size: 12px;
}
dd {
  margin: 0;
  overflow-wrap: anywhere;
}
code {
  color: var(--soft);
  font-size: 11px;
}
dd code {
  margin-left: 8px;
}
p,
label {
  color: var(--muted);
  line-height: 1.6;
  font-size: 13px;
}
.review-confirmation {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin: 20px 0;
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
}
input {
  flex: 0 0 auto;
  margin-top: 4px;
}
footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 24px;
}
button {
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg);
  color: var(--text);
  cursor: pointer;
}
button.primary {
  background: var(--accent);
  color: var(--on-accent);
  border-color: transparent;
  font-weight: 700;
}
button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
