import { useQuery } from '@tanstack/vue-query'
import { inject, provide, type InjectionKey, type Ref } from 'vue'

import { usePresentationActivity } from '../../system/composables/usePresentationActivity'
import { loadProjects } from '../api'

function observeProjects(active: Ref<boolean>) {
  return useQuery({
    shallow: true,
    enabled: active,
    queryKey: ['projects'],
    queryFn: loadProjects,
    refetchInterval: 15_000,
  })
}

const catalogKey: InjectionKey<ReturnType<typeof observeProjects>> = Symbol('project-catalog')

export function provideProjectCatalog(active: Ref<boolean>) {
  provide(catalogKey, observeProjects(active))
}

export function useProjectCatalog(): ReturnType<typeof observeProjects> {
  return inject(catalogKey) ?? observeProjects(usePresentationActivity())
}
