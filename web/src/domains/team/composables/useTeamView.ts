import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, ref, watch } from 'vue'

import type { TeamBundle, TeamPublisher } from '../../../api/generated/types.gen'
import { parseBundleFile } from '../bundleFile'
import {
  installBundle,
  loadCuratedPlugins,
  loadEffectivePolicy,
  loadTeamBundles,
  loadTeamPublishers,
  trustPublisher,
} from '../api'

export function useTeamView() {
  const queryClient = useQueryClient()
  const publishers = useQuery({ queryKey: ['team-publishers'], queryFn: loadTeamPublishers })
  const bundles = useQuery({ queryKey: ['team-bundles'], queryFn: loadTeamBundles })
  const policy = useQuery({ queryKey: ['team-policy'], queryFn: loadEffectivePolicy })
  const registry = useQuery({ queryKey: ['curated-plugins'], queryFn: loadCuratedPlugins })
  const publisherName = ref('')
  const publicKey = ref('')
  const publisherReviewed = ref(false)
  const selectedBundle = ref<TeamBundle>()
  const bundleReviewed = ref(false)
  const fileError = ref('')
  let fileSelection = 0
  watch([publisherName, publicKey], () => (publisherReviewed.value = false), { flush: 'sync' })
  const trust = useMutation({
    mutationFn: () => {
      if (!publisherReviewed.value) throw new Error('Review the exact publisher key first.')
      return trustPublisher(publisherName.value, publicKey.value, publisherReviewed.value)
    },
    onSuccess: (value) => {
      queryClient.setQueryData<Array<TeamPublisher>>(['team-publishers'], (current = []) => [
        ...current.filter((item) => item.id !== value.id),
        value,
      ])
      publisherName.value = ''
      publicKey.value = ''
      publisherReviewed.value = false
    },
  })
  const install = useMutation({
    mutationFn: () => {
      if (!bundleReviewed.value || !selectedBundle.value)
        throw new Error('Review the selected bundle and payload first.')
      return installBundle(selectedBundle.value, bundleReviewed.value)
    },
    onSuccess: async () => {
      selectedBundle.value = undefined
      bundleReviewed.value = false
      await Promise.all([bundles.refetch(), policy.refetch(), registry.refetch()])
    },
  })
  const error = computed(
    () =>
      publishers.error.value ||
      bundles.error.value ||
      policy.error.value ||
      registry.error.value ||
      trust.error.value ||
      install.error.value,
  )
  const queryError = computed(
    () =>
      publishers.isError.value ||
      bundles.isError.value ||
      policy.isError.value ||
      registry.isError.value,
  )
  const loading = computed(() =>
    [publishers, bundles, policy, registry].some((query) => query.isPending.value),
  )
  const emptyConfiguration = computed(
    () =>
      !loading.value &&
      !queryError.value &&
      (publishers.data.value?.length ?? 0) === 0 &&
      (bundles.data.value?.length ?? 0) === 0 &&
      (policy.data.value?.sourceBundleIds.length ?? 0) === 0 &&
      (registry.data.value?.length ?? 0) === 0,
  )

  async function selectBundle(event: unknown) {
    const selection = ++fileSelection
    fileError.value = ''
    selectedBundle.value = undefined
    bundleReviewed.value = false
    install.reset()
    const file = (
      event as {
        target?: { files?: ArrayLike<{ size: number; text(): Promise<string> }> }
      }
    ).target?.files?.[0]
    if (!file) return
    if (file.size > 2 * 1024 * 1024) {
      fileError.value = 'Bundle files must be 2 MiB or smaller.'
      return
    }
    try {
      const bundle = parseBundleFile(await file.text())
      if (selection === fileSelection) selectedBundle.value = bundle
    } catch (error) {
      if (selection === fileSelection)
        fileError.value =
          error instanceof SyntaxError
            ? 'The selected file is not valid JSON.'
            : (error as Error).message
    }
  }
  return {
    publishers,
    bundles,
    policy,
    registry,
    publisherName,
    publicKey,
    publisherReviewed,
    selectedBundle,
    bundleReviewed,
    fileError,
    trust,
    install,
    error,
    queryError,
    retry: () =>
      Promise.all([publishers.refetch(), bundles.refetch(), policy.refetch(), registry.refetch()]),
    loading,
    emptyConfiguration,
    selectBundle,
  }
}
