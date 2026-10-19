import { useQuery } from '@tanstack/vue-query'
import { inject, provide, type InjectionKey, type Ref } from 'vue'

import { loadHostObservation } from '../api'
import { usePresentationActivity } from './usePresentationActivity'

function observeHost(active: Ref<boolean>) {
  return useQuery({
    shallow: true,
    enabled: active,
    queryKey: ['host-observation'],
    queryFn: loadHostObservation,
    refetchInterval: 15_000,
  })
}

const hostKey: InjectionKey<ReturnType<typeof observeHost>> = Symbol('host-observation')

// Sharing cached data alone does not share QueryObserver polling timers.
export function provideHostObservation(active: Ref<boolean>) {
  provide(hostKey, observeHost(active))
}

export function useHostObservation(): ReturnType<typeof observeHost> {
  return inject(hostKey) ?? observeHost(usePresentationActivity())
}
