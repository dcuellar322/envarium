import { useQuery } from '@tanstack/vue-query'
import { computed, ref, type ComputedRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { loadPortRegistry } from '../../ports/api'
import { projectTabs, type ProjectTab } from '../projectTabs'

export function useProjectTabs(projectId: ComputedRef<string>) {
  const route = useRoute()
  const router = useRouter()
  const activeTab = ref<ProjectTab>(
    projectTabs.includes(route.query.tab as ProjectTab)
      ? (route.query.tab as ProjectTab)
      : 'overview',
  )
  const portRegistry = useQuery({
    queryKey: ['ports'],
    queryFn: loadPortRegistry,
    enabled: computed(() => activeTab.value === 'ports'),
    refetchInterval: computed(() => (activeTab.value === 'ports' ? 10_000 : false)),
  })
  const projectPorts = computed(
    () => portRegistry.data.value?.facts.filter((fact) => fact.projectId === projectId.value) ?? [],
  )

  async function selectTab(tab: ProjectTab) {
    activeTab.value = tab
    await router.replace({ query: tab === 'overview' ? {} : { tab } })
  }

  return { activeTab, projectPorts, selectTab }
}
