<script setup lang="ts">
import { computed } from 'vue'
import type { TeamBundle } from '../../../api/generated/types.gen'

const props = defineProps<{ bundle: TeamBundle }>()
const payload = computed(() => {
  const compact = JSON.stringify(props.bundle.payload)
  if (compact.length > 64 << 10) return compact
  const pretty = JSON.stringify(props.bundle.payload, null, 2)
  return pretty.length <= 256 << 10 ? pretty : compact
})
</script>

<template>
  <section class="bundle-review" aria-label="Selected bundle review">
    <h3>{{ bundle.metadata.name }}</h3>
    <dl>
      <div>
        <dt>Kind</dt>
        <dd>{{ bundle.kind }}</dd>
      </div>
      <div>
        <dt>Version</dt>
        <dd>{{ bundle.metadata.version }}</dd>
      </div>
      <div>
        <dt>Bundle ID</dt>
        <dd>
          <code>{{ bundle.metadata.id }}</code>
        </dd>
      </div>
      <div>
        <dt>Publisher</dt>
        <dd>
          <code>{{ bundle.metadata.publisherId }}</code>
        </dd>
      </div>
    </dl>
    <label
      >Signed payload JSON<textarea
        :value="payload"
        readonly
        spellcheck="false"
        wrap="off"
        rows="9"
      />
    </label>
    <p>
      Review these settings before installation. The daemon will verify the signature against your
      trusted key.
    </p>
  </section>
</template>

<style scoped>
.bundle-review {
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--panel-2);
}
h3 {
  margin: 0 0 12px;
  font-size: 15px;
}
dl {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0 0 14px;
}
dt,
p {
  color: var(--muted);
  font-size: 11px;
}
dd {
  margin: 4px 0 0;
  overflow-wrap: anywhere;
}
label {
  display: grid;
  gap: 6px;
  color: var(--muted);
}
textarea {
  box-sizing: border-box;
  width: 100%;
  resize: vertical;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 10px;
  background: var(--bg);
  color: var(--text);
  font:
    11px/1.6 ui-monospace,
    monospace;
}
p {
  margin: 10px 0 0;
  line-height: 1.5;
}
@media (max-width: 600px) {
  dl {
    grid-template-columns: 1fr;
  }
}
</style>
