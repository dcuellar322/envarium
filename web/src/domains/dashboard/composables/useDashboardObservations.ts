import { useQuery } from '@tanstack/vue-query'
import { computed, type Ref } from 'vue'

import type { Project, ProjectObservationBatch } from '../../../api/generated/types.gen'
import { loadResourceOverview } from '../../resources/api'
import { usePresentationActivity } from '../../system/composables/usePresentationActivity'
import type { ProjectSnapshot } from '../api'
import { loadProjectObservations } from '../observations'

// Batches are homogeneous: a runtime event cannot refresh Git or quick actions.
export function useDashboardObservations(projects: Ref<Project[]>) {
  const active = usePresentationActivity()
  const ids = computed(() =>
    projects.value
      .filter((project) => project.trustState === 'trusted')
      .map((project) => project.id)
      .sort(),
  )
  const observe = (kind: ProjectObservationBatch['kind']) =>
    useQuery({
      shallow: true,
      queryKey: computed(() => ['project-observations', kind, ids.value] as const),
      queryFn: ({ signal, queryKey }) => loadProjectObservations(kind, queryKey[2], signal),
      enabled: computed(() => active.value && ids.value.length > 0),
      refetchInterval: 15_000,
    })
  const runtime = observe('runtime')
  const git = observe('git-status')
  const actions = observe('actions')
  const resources = useQuery({
    shallow: true,
    queryKey: ['resource-overview'],
    queryFn: ({ signal }) => loadResourceOverview(signal),
    enabled: active,
    refetchInterval: 10_000,
  })
  const data = computed<ProjectSnapshot[]>((previous) =>
    projects.value.map((project) => {
      if (project.trustState !== 'trusted') return { project, warnings: [] }
      const runtimeItem = runtime.data.value?.find((item) => item.projectId === project.id)
      const gitItem = git.data.value?.find((item) => item.projectId === project.id)
      const actionsItem = actions.data.value?.find((item) => item.projectId === project.id)
      const resource = resources.data.value?.projects.find((item) => item.projectId === project.id)
      const warnings: string[] = []
      if (runtime.isError.value || runtimeItem?.warning)
        warnings.push('Runtime observation unavailable')
      if (git.isError.value || gitItem?.warning) warnings.push('Git observation unavailable')
      if (actions.isError.value || actionsItem?.warning) warnings.push('Quick actions unavailable')
      if (resources.isError.value) warnings.push('Resource samples unavailable')
      if (
        resource?.metric.partial ||
        (resource?.active && (!resource.metric.cpuAvailable || !resource.metric.memoryAvailable))
      )
        warnings.push('Resource samples are partial')
      const prior = previous?.find((item) => item.project.id === project.id)
      const next: ProjectSnapshot = {
        project,
        runtime: runtimeItem?.runtime,
        git: gitItem?.git,
        actions: actionsItem?.actions,
        metrics:
          resource?.active || runtimeItem?.runtime?.state === 'stopped'
            ? resource?.services.map((service) => service.metric)
            : undefined,
        warnings,
      }
      if (
        prior?.metrics &&
        next.metrics?.every((metric, i) => metric === prior.metrics?.[i]) &&
        next.metrics.length === prior.metrics.length
      )
        next.metrics = prior.metrics
      if (
        prior &&
        prior.project === next.project &&
        prior.runtime === next.runtime &&
        prior.git === next.git &&
        prior.actions === next.actions &&
        prior.metrics === next.metrics &&
        prior.warnings.join(' | ') === warnings.join(' | ')
      )
        return prior
      return next
    }),
  )
  const queries = [runtime, git, actions, resources]
  const isFetching = computed(() => queries.some((query) => query.isFetching.value))
  // Cache staleness schedules reads; it does not mean a healthy observation is
  // unavailable. Toggling page layout on cache timers also repaints the backdrop.
  const isStale = computed(() => !active.value || queries.some((query) => query.isError.value))
  return { data, isFetching, isStale }
}
