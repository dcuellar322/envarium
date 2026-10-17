import { getProjectObservations } from '../../api/generated/sdk.gen'
import type { ProjectObservationBatch } from '../../api/generated/types.gen'

export async function loadProjectObservations(
  kind: ProjectObservationBatch['kind'],
  ids: readonly string[],
  signal?: AbortSignal,
): Promise<ProjectObservationBatch['projects']> {
  const projects: ProjectObservationBatch['projects'] = []
  for (let offset = 0; offset < ids.length; offset += 64) {
    signal?.throwIfAborted()
    const result = await getProjectObservations({
      query: { kind, projectId: ids.slice(offset, offset + 64) },
      signal,
    })
    if (result.error || !result.data) throw new Error('Project observations are unavailable.')
    projects.push(...result.data.projects)
  }
  return projects
}
