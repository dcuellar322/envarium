import { computed, ref, watch, onScopeDispose, type ComputedRef } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import type { GitBranchSwitchPlan, RuntimePlan } from '../../../api/generated/types.gen'
import { isTerminalOperation } from '../../../lib/format'
import { trackOperation, useOperationStore } from '../../operations/store'
import { previewBranchSwitch, switchBranch } from '../gitApi'
import { loadProjectRuntime, previewRuntimeAction, runRuntimeAction } from '../api'

type Step =
  | { kind: 'runtime'; id: string; action: 'stop' | 'start' | 'restart' }
  | { kind: 'branch'; id: string; branch: string }

/** Each stage needs its own fresh review; failure or cancellation ends the sequence. */
export function useCheckoutTransition(
  context: ComputedRef<string>,
  changed: (selected?: string) => void,
) {
  const client = useQueryClient()
  const store = useOperationStore()
  const runtimePlan = ref<RuntimePlan>()
  const branchPlan = ref<GitBranchSwitchPlan>()
  const error = ref('')
  const busy = ref(false)
  const operationId = ref('')
  const message = ref('')
  let steps: Step[] = []
  let generation = 0
  const pending = computed(() => busy.value || Boolean(operationId.value))
  const active = computed(() => pending.value || Boolean(runtimePlan.value || branchPlan.value))
  const operation = computed(() =>
    store.operations.value.find((item) => item.id === operationId.value),
  )

  function cancel() {
    generation++
    steps = []
    runtimePlan.value = undefined
    branchPlan.value = undefined
    // A queued operation stays visible in Operations; cancelling review never undoes it.
    operationId.value = ''
    busy.value = false
    message.value = ''
  }
  watch(context, cancel)
  onScopeDispose(cancel)
  function fail(cause: unknown) {
    cancel()
    error.value =
      cause instanceof Error
        ? cause.message
        : 'Checkout transition failed. Review the current checkout and services before retrying.'
  }
  async function advance() {
    const step = steps[0]
    if (!step) {
      message.value = 'Checkout transition completed.'
      return
    }
    busy.value = true
    const current = generation
    try {
      if (step.kind === 'branch') {
        message.value =
          'Review the branch change. Services remain stopped until you review the new start plan.'
        const plan = await previewBranchSwitch(step.id, step.branch)
        if (generation === current) branchPlan.value = plan
      } else {
        message.value =
          step.action === 'stop'
            ? 'Review stopping the source checkout before continuing.'
            : 'Review services and commands from the selected checkout before starting.'
        const plan = await previewRuntimeAction(step.id, step.action)
        if (generation === current) runtimePlan.value = plan
      }
    } catch (cause) {
      if (generation === current) fail(cause)
    } finally {
      if (generation === current) busy.value = false
    }
  }
  async function prepare(sourceIds: string[], targetId: string, branch?: string) {
    if (active.value) return
    cancel()
    error.value = ''
    busy.value = true
    const current = generation
    try {
      for (const id of [...new Set(sourceIds)]) {
        const observation = await loadProjectRuntime(id)
        if (generation !== current) return
        if (observation.state !== 'stopped') steps.push({ kind: 'runtime', id, action: 'stop' })
      }
      if (branch) steps.push({ kind: 'branch', id: targetId, branch })
      steps.push({ kind: 'runtime', id: targetId, action: branch ? 'start' : 'restart' })
      await advance()
    } catch (cause) {
      if (generation === current) fail(cause)
    } finally {
      if (generation === current) busy.value = false
    }
  }
  async function confirm() {
    if (pending.value || !steps.length) return
    const current = generation
    busy.value = true
    try {
      const plan = runtimePlan.value
      const branch = branchPlan.value
      const result = plan
        ? await runRuntimeAction(plan.projectId, plan.action, {
            confirmRisk: true,
            expectedCheckoutRevision: plan.checkoutRevision,
            expectedManifestHash: plan.manifestHash,
          })
        : branch
          ? await switchBranch(branch.projectId, branch, true)
          : undefined
      if (current !== generation || !result) return
      operationId.value = result.id
      runtimePlan.value = undefined
      branchPlan.value = undefined
      trackOperation(result)
    } catch (cause) {
      if (current === generation) fail(cause)
    } finally {
      if (current === generation) busy.value = false
    }
  }
  watch(operation, async (value) => {
    if (!value || !isTerminalOperation(value.state)) return
    operationId.value = ''
    if (value.state !== 'succeeded') {
      fail(
        new Error(
          value.errorMessage ||
            `Transition ${value.state}. Services may remain stopped; inspect Operations before retrying.`,
        ),
      )
      return
    }
    const current = generation
    const completed = steps.shift()
    busy.value = true
    await Promise.all(
      [
        'project-git',
        'project-runtime',
        'project-manifest',
        'project-actions',
        'project-environments',
      ].map((key) => client.invalidateQueries({ queryKey: [key] })),
    )
    if (current !== generation) return
    changed(steps.length ? undefined : completed?.id)
    await advance()
    if (current === generation) busy.value = false
  })
  return { runtimePlan, branchPlan, error, pending, active, message, prepare, confirm, cancel }
}
