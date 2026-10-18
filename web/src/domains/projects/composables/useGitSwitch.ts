import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { computed, ref, watch, type ComputedRef } from 'vue'
import type { GitBranchSwitchPlan } from '../../../api/generated/types.gen'
import { isTerminalOperation } from '../../../lib/format'
import { trackOperation, useOperationStore } from '../../operations/store'
import { previewBranchSwitch, switchBranch } from '../gitApi'

export function useGitSwitch(projectId: ComputedRef<string>, switched: () => void) {
  const queryClient = useQueryClient()
  const store = useOperationStore()
  const plan = ref<GitBranchSwitchPlan>()
  const operationId = ref('')
  const error = ref('')
  const preview = useMutation({
    mutationFn: (branch: string) => previewBranchSwitch(projectId.value, branch),
  })
  const execute = useMutation({
    mutationFn: (reviewed: GitBranchSwitchPlan) => switchBranch(reviewed.projectId, reviewed, true),
  })
  const operation = computed(() =>
    store.operations.value.find((item) => item.id === operationId.value),
  )
  const pending = computed(
    () =>
      preview.isPending.value ||
      execute.isPending.value ||
      Boolean(operation.value && !isTerminalOperation(operation.value.state)),
  )
  watch(projectId, () => {
    plan.value = undefined
    operationId.value = ''
    error.value = ''
  })
  watch(operation, async (value) => {
    if (!value || !isTerminalOperation(value.state)) return
    operationId.value = ''
    if (value.state !== 'succeeded') {
      error.value =
        value.errorMessage ||
        'Branch switch did not complete. Inspect the repository before retrying.'
      return
    }
    await Promise.all(
      [
        'project-git',
        'project-runtime',
        'project-manifest',
        'project-actions',
        'project-environments',
      ].map((key) => queryClient.invalidateQueries({ queryKey: [key, value.projectId] })),
    )
    if (value.projectId === projectId.value) switched()
  })
  async function prepare(branch: string) {
    error.value = ''
    plan.value = undefined
    const id = projectId.value
    try {
      const result = await preview.mutateAsync(branch)
      if (projectId.value === id) plan.value = result
    } catch (cause) {
      if (projectId.value === id)
        error.value = cause instanceof Error ? cause.message : 'Branch preview failed.'
    }
  }
  async function confirm() {
    if (!plan.value || execute.isPending.value) return
    error.value = ''
    try {
      const operation = await execute.mutateAsync(plan.value)
      operationId.value = operation.id
      trackOperation(operation)
      plan.value = undefined
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : 'Branch switch failed.'
    }
  }
  return {
    plan,
    error,
    pending,
    prepare,
    confirm,
    cancel: () => {
      plan.value = undefined
    },
  }
}
