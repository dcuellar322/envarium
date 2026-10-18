import { useMutation, useQuery } from '@tanstack/vue-query'
import { computed, ref, watch, type ComputedRef } from 'vue'

import type {
  ActionDefinition,
  RuntimeAction,
  RuntimeLogEntry,
  RuntimePlan,
} from '../../../api/generated/types.gen'
import { isActiveState } from '../../../lib/format'
import { loadProjectEnvironments, registerEnvironments } from '../../environments/api'
import { trackOperation } from '../../operations/store'
import {
  loadEffectiveManifest,
  loadProject,
  loadProjectActions,
  loadProjectGitInventory,
  loadProjectGitStatus,
  loadProjectHealth,
  loadProjectLogs,
  loadProjectMetrics,
  loadProjectRuntime,
  previewRuntimeAction,
  runProjectAction,
  runRuntimeAction,
  type RuntimeSelection,
} from '../api'
import { markProjectAccess } from '../recent'
import { useProjectLogStream } from './useProjectLogStream'

export function useProjectDetail(
  projectId: ComputedRef<string>,
  gitDetailsActive: ComputedRef<boolean> = computed(() => false),
) {
  const operationError = ref('')
  const liveLogs = ref<Array<RuntimeLogEntry>>([])
  const lifecyclePlan = ref<RuntimePlan>()
  const lifecycleSelection = ref<RuntimeSelection>({})
  const lifecycleTarget = ref('')
  const project = useQuery({
    queryKey: computed(() => ['project', projectId.value]),
    queryFn: () => loadProject(projectId.value),
  })
  const runtime = useQuery({
    queryKey: computed(() => ['project-runtime', projectId.value]),
    queryFn: () => loadProjectRuntime(projectId.value),
    refetchInterval: 5_000,
  })
  const health = useQuery({
    queryKey: computed(() => ['project-health', projectId.value]),
    queryFn: () => loadProjectHealth(projectId.value),
    refetchInterval: 5_000,
  })
  const logs = useQuery({
    queryKey: computed(() => ['project-logs', projectId.value]),
    queryFn: () => loadProjectLogs(projectId.value),
    refetchInterval: 10_000,
  })
  const metrics = useQuery({
    queryKey: computed(() => ['project-metrics', projectId.value]),
    queryFn: () => loadProjectMetrics(projectId.value),
    refetchInterval: 5_000,
  })
  const git = useQuery({
    queryKey: computed(() => [
      'project-git',
      projectId.value,
      gitDetailsActive.value ? 'full' : 'status',
    ]),
    queryFn: ({ signal }) =>
      gitDetailsActive.value
        ? loadProjectGitInventory(projectId.value, signal)
        : loadProjectGitStatus(projectId.value, signal),
    shallow: true,
    refetchInterval: 10_000,
  })
  const actions = useQuery({
    queryKey: computed(() => ['project-actions', projectId.value]),
    queryFn: () => loadProjectActions(projectId.value),
  })
  const manifest = useQuery({
    queryKey: computed(() => ['project-manifest', projectId.value]),
    queryFn: () => loadEffectiveManifest(projectId.value),
  })
  const environments = useQuery({
    queryKey: computed(() => ['project-environments', projectId.value]),
    queryFn: () => loadProjectEnvironments(projectId.value),
    refetchInterval: 10_000,
  })

  watch(projectId, markProjectAccess, { immediate: true })
  watch(projectId, closeLifecyclePlan)
  watch(
    () => logs.data.value,
    (entries) => {
      if (entries) liveLogs.value = entries.slice(-300)
    },
    { immediate: true },
  )
  const logConnection = useProjectLogStream(projectId, (entry) => {
    if (liveLogs.value.some((item) => item.sequence === entry.sequence)) return
    liveLogs.value = [...liveLogs.value, entry].slice(-300)
  })

  const lifecycle = useMutation({
    mutationFn: ({ action, selection }: { action: RuntimeAction; selection: RuntimeSelection }) =>
      runRuntimeAction(lifecycleTarget.value || projectId.value, action, selection),
    onSuccess: trackOperation,
  })
  const lifecyclePreview = useMutation({
    mutationFn: ({ action, selection }: { action: RuntimeAction; selection: RuntimeSelection }) =>
      previewRuntimeAction(lifecycleTarget.value || projectId.value, action, selection),
  })
  const customAction = useMutation({
    mutationFn: (action: ActionDefinition) => runProjectAction(projectId.value, action.id),
    onSuccess: trackOperation,
  })
  const registerWorktrees = useMutation({
    mutationFn: () => registerEnvironments(projectId.value),
    onSuccess: async () => {
      await Promise.all([environments.refetch(), git.refetch()])
    },
  })

  const state = computed(() => runtime.data.value?.state ?? 'unknown')
  const active = computed(() => isActiveState(state.value))
  const stateTone = computed(() =>
    ['degraded', 'failed', 'partially_running'].includes(state.value)
      ? 'warning'
      : active.value
        ? 'ready'
        : 'neutral',
  )
  const availableMetrics = computed(
    () => metrics.data.value?.filter((item) => item.memoryAvailable) ?? [],
  )
  const memory = computed(() =>
    availableMetrics.value.reduce((sum, item) => sum + item.memoryBytes, 0),
  )
  const cpu = computed(
    () =>
      metrics.data.value?.reduce(
        (sum, item) => sum + (item.cpuAvailable ? item.cpuPercent : 0),
        0,
      ) ?? 0,
  )
  const cpuAvailable = computed(() =>
    Boolean(metrics.data.value?.some((item) => item.cpuAvailable)),
  )
  const memoryAvailable = computed(() => availableMetrics.value.length > 0)
  const memoryLimit = computed(() =>
    Math.max(0, ...availableMetrics.value.map((item) => item.memoryLimit)),
  )
  const changes = computed(() => {
    const value = git.data.value?.changes
    return value ? value.staged + value.modified + value.untracked + value.conflicted : 0
  })
  const recentLogs = computed(() => liveLogs.value.slice(-80))
  const terminalAction = computed(() =>
    actions.data.value?.actions.find(
      (action) => action.id === 'terminal' || action.type === 'terminal.open',
    ),
  )
  const browserAction = computed(() =>
    actions.data.value?.actions.find((action) => action.type === 'browser.open'),
  )
  const endpoints = computed(
    () =>
      actions.data.value?.actions.filter(
        (action) => action.type === 'browser.open' && action.target,
      ) ?? [],
  )
  const quickActions = computed(
    () =>
      actions.data.value?.actions
        .filter((action) => !['terminal.open', 'browser.open'].includes(action.type))
        .slice(0, 5) ?? [],
  )
  const requiredHealth = computed(
    () => health.data.value?.results.filter((item) => item.required) ?? [],
  )
  const isPartial = computed(() =>
    [runtime, health, metrics, git, actions].some(
      (query) => query.isError.value && query.data.value === undefined,
    ),
  )

  async function requestLifecycle(
    action: RuntimeAction,
    selection: RuntimeSelection = {},
    target = projectId.value,
  ) {
    if (lifecyclePreview.isPending.value || lifecycle.isPending.value) return
    const requestedProject = projectId.value
    operationError.value = ''
    try {
      lifecycleSelection.value = selection
      lifecycleTarget.value = target
      const plan = await lifecyclePreview.mutateAsync({ action, selection })
      if (projectId.value === requestedProject) lifecyclePlan.value = plan
    } catch (cause) {
      operationError.value =
        cause instanceof Error ? cause.message : 'The lifecycle preview could not be prepared.'
    }
  }

  async function confirmLifecycle() {
    if (!lifecyclePlan.value) return
    operationError.value = ''
    try {
      await lifecycle.mutateAsync({
        action: lifecyclePlan.value.action,
        selection: {
          ...lifecycleSelection.value,
          confirmRisk: true,
          ...(lifecyclePlan.value.checkoutRevision
            ? { expectedCheckoutRevision: lifecyclePlan.value.checkoutRevision }
            : {}),
          ...(lifecyclePlan.value.manifestHash
            ? { expectedManifestHash: lifecyclePlan.value.manifestHash }
            : {}),
        },
      })
      closeLifecyclePlan()
    } catch (cause) {
      operationError.value =
        cause instanceof Error ? cause.message : 'The lifecycle operation could not be queued.'
    }
  }

  function closeLifecyclePlan() {
    lifecyclePlan.value = undefined
    lifecycleSelection.value = {}
    lifecycleTarget.value = ''
  }

  async function runAction(action: ActionDefinition | undefined) {
    if (!action) return
    operationError.value = ''
    try {
      await customAction.mutateAsync(action)
    } catch (cause) {
      operationError.value =
        cause instanceof Error ? cause.message : 'The trusted action could not be queued.'
    }
  }

  function refreshGitContext() {
    void Promise.all([git.refetch(), runtime.refetch(), environments.refetch()])
  }

  return {
    refreshGitContext,
    project,
    runtime,
    health,
    metrics,
    git,
    actions,
    manifest,
    environments,
    operationError,
    liveLogs,
    logConnection,
    lifecycle,
    lifecyclePreview,
    lifecyclePlan,
    customAction,
    registerWorktrees,
    state,
    active,
    stateTone,
    memory,
    cpu,
    cpuAvailable,
    memoryAvailable,
    memoryLimit,
    changes,
    recentLogs,
    terminalAction,
    browserAction,
    endpoints,
    quickActions,
    requiredHealth,
    isPartial,
    requestLifecycle,
    confirmLifecycle,
    closeLifecyclePlan,
    runAction,
  }
}
