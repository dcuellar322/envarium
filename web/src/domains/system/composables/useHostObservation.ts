import { useQuery } from '@tanstack/vue-query'

import { loadHostObservation } from '../api'
import { usePresentationActivity } from './usePresentationActivity'

export function useHostObservation() {
  const active = usePresentationActivity()
  return useQuery({
    shallow: true,
    enabled: active,
    queryKey: ['host-observation'],
    queryFn: loadHostObservation,
    refetchInterval: 15_000,
  })
}
