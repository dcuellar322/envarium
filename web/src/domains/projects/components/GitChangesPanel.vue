<script setup lang="ts">
import { computed } from 'vue'
import type { GitDetails } from '../../../api/generated/types.gen'

const props = defineProps<{ details?: GitDetails }>()
const groups = computed(() => {
  const files = props.details?.files ?? []
  return [
    { name: 'Conflicts', files: files.filter((file) => file.conflicted), tone: 'danger' },
    {
      name: 'Staged',
      files: files.filter((file) => !file.conflicted && !file.untracked && file.index !== ' '),
      tone: 'staged',
    },
    {
      name: 'Unstaged',
      files: files.filter(
        (file) => !file.conflicted && !file.untracked && file.workingTree !== ' ',
      ),
      tone: 'modified',
    },
    { name: 'Untracked', files: files.filter((file) => file.untracked), tone: 'untracked' },
  ]
})
</script>

<template>
  <article class="panel git-changes">
    <header class="panel-head">
      <div>
        <p>Working directory &amp; index</p>
        <h2>Changed files</h2>
      </div>
    </header>
    <p v-if="!details" class="panel-state">
      File inventory has not been loaded. Refresh the repository snapshot.
    </p>
    <template v-else>
      <p v-if="details.filesTruncated" class="message message--error" role="status">
        Showing the first 1,024 changed files. Use Git locally for the complete inventory.
      </p>
      <p v-if="!details.files.length" class="panel-state">
        Working tree clean. No staged, unstaged, or untracked files.
      </p>
      <div v-else class="change-groups">
        <section
          v-for="group in groups.filter((item) => item.files.length)"
          :key="group.name"
          :aria-label="group.name"
        >
          <h3>
            {{ group.name }} <span>{{ group.files.length }}</span>
          </h3>
          <ul>
            <li v-for="file in group.files" :key="file.path" :class="group.tone">
              <code
                class="status"
                :aria-label="`Index ${file.index}; working tree ${file.workingTree}`"
                >{{ file.index }}{{ file.workingTree }}</code
              >
              <span
                ><code class="file-path">{{ file.path }}</code
                ><small v-if="file.originalPath"
                  >from <code>{{ file.originalPath }}</code></small
                ></span
              >
            </li>
          </ul>
        </section>
      </div>
    </template>
  </article>
</template>

<style scoped>
.change-groups {
  display: grid;
  gap: 22px;
}
h3 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 10px;
  font-size: 13px;
}
h3 span {
  color: var(--muted);
  font-weight: 500;
}
ul {
  margin: 0;
  padding: 0;
  list-style: none;
}
li {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 12px;
  border-top: 1px solid var(--border);
  min-width: 0;
}
li:first-child {
  border-top: 0;
}
li > span {
  min-width: 0;
}
.status {
  white-space: pre;
  color: var(--yellow);
  flex: 0 0 2ch;
}
.staged .status {
  color: var(--green);
}
.danger .status {
  color: var(--red);
}
.file-path,
small code {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
small {
  display: block;
  margin-top: 5px;
  color: var(--muted);
}
</style>
