import {
  acceptManifestProposal,
  cancelOperation,
  createAiManifestEnhancement,
  createActionOperation,
  createManifestProposal,
  createProjectOperation,
  explainProjectManifest,
  exportProjectLogs as exportProjectLogsRequest,
  getAiManifestEnhancement,
  getManifestProposal,
  getOperation,
  getProject,
  getProjectGit,
  getProjectGitInventory,
  getProjectGitStatus,
  getProjectHealth,
  getProjectLogs,
  getProjectMetrics,
  getProjectRuntime,
  listProjectActions,
  listAiProposalProviders,
  listProjects,
  planProjectRuntime,
  validateManifestProposal,
  previewAiManifestEvidence,
} from '../../api/generated/sdk.gen'
import type {
  AcceptedManifestProposal,
  AiEvidencePreview,
  AiGenerationLimits,
  AiManifestEnhancement,
  AiProviderDescriptor,
  EffectiveManifest,
  GitState,
  ManifestProposal,
  Operation,
  Project,
  ProjectActions,
  ProjectHealth,
  RuntimeAction,
  RuntimeActionRequest,
  RuntimePlan,
  RuntimeLogEntry,
  RuntimeMetricSample,
  RuntimeObservation,
} from '../../api/generated/types.gen'
import { mutationHeaders } from '../session/bootstrap'

export class ProjectAPIError extends Error {
  constructor(
    message: string,
    readonly code = '',
  ) {
    super(message)
    this.name = 'ProjectAPIError'
  }
}

function requestKey(): string {
  return `ui_${crypto.randomUUID()}`
}

export async function loadProjects(): Promise<Array<Project>> {
  const result = await listProjects()
  if (result.error || !result.data) throw new Error('The project catalog could not be loaded.')
  return result.data
}

export async function loadProject(projectId: string): Promise<Project> {
  const result = await getProject({ path: { projectId } })
  if (result.error || !result.data) throw new Error('The project could not be loaded.')
  return result.data
}

export async function scanRepository(
  path: string,
  allowOutsideRoots = false,
): Promise<ManifestProposal> {
  const result = await createManifestProposal({
    body: { path, allowOutsideRoots },
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data)
    throw new ProjectAPIError(
      result.error?.detail || 'The repository scan could not create a proposal.',
      result.error?.code,
    )
  return result.data
}

export async function revalidateProposal(proposalId: string): Promise<ManifestProposal> {
  const result = await validateManifestProposal({
    path: { proposalId },
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data) throw new Error('The manifest proposal could not be validated.')
  return result.data
}

export async function approveProposal(proposalId: string): Promise<AcceptedManifestProposal> {
  const result = await acceptManifestProposal({
    path: { proposalId },
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data) throw new Error('The proposal could not be approved.')
  return result.data
}

export async function loadAIProviders(): Promise<Array<AiProviderDescriptor>> {
  const result = await listAiProposalProviders()
  if (result.error || !result.data)
    throw new Error('Assisted-onboarding providers could not be loaded.')
  return result.data
}

export async function previewAIEvidence(
  proposalId: string,
  provider: string,
  limits: AiGenerationLimits,
): Promise<AiEvidencePreview> {
  const result = await previewAiManifestEvidence({
    path: { proposalId },
    body: { provider, limits },
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data)
    throw new Error('The provider evidence preview could not be prepared.')
  return result.data
}

export async function startAIEnhancement(
  proposalId: string,
  provider: string,
  limits: AiGenerationLimits,
  consentSha256: string,
): Promise<Operation> {
  const result = await createAiManifestEnhancement({
    path: { proposalId },
    body: { provider, limits, consentSha256 },
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data)
    throw new Error('The assisted-onboarding operation could not be queued.')
  return result.data
}

export async function loadOperation(operationId: string): Promise<Operation> {
  const result = await getOperation({ path: { operationId } })
  if (result.error || !result.data)
    throw new Error('The assisted-onboarding operation became unavailable.')
  return result.data
}

export async function stopOperation(operationId: string): Promise<Operation> {
  const result = await cancelOperation({
    path: { operationId },
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data)
    throw new Error('The cancellation request could not be recorded.')
  return result.data
}

export async function loadAIEnhancement(
  proposalId: string,
  operationId: string,
): Promise<AiManifestEnhancement> {
  const result = await getAiManifestEnhancement({ path: { proposalId, operationId } })
  if (result.error || !result.data)
    throw new Error('The assisted-onboarding receipt could not be loaded.')
  return result.data
}

export async function loadManifestProposal(proposalId: string): Promise<ManifestProposal> {
  const result = await getManifestProposal({ path: { proposalId } })
  if (result.error || !result.data)
    throw new Error('The generated manifest proposal could not be loaded.')
  return result.data
}

export async function loadProjectRuntime(
  projectId: string,
  signal?: AbortSignal,
): Promise<RuntimeObservation> {
  const result = await getProjectRuntime({ path: { projectId }, signal })
  if (result.error || !result.data) throw new Error('Runtime observation is unavailable.')
  return result.data
}

export async function loadProjectHealth(projectId: string): Promise<ProjectHealth> {
  const result = await getProjectHealth({ path: { projectId } })
  if (result.error || !result.data) throw new Error('Health diagnostics are unavailable.')
  return result.data
}

export async function loadProjectLogs(
  projectId: string,
  tail = 200,
): Promise<Array<RuntimeLogEntry>> {
  const result = await getProjectLogs({ path: { projectId }, query: { tail } })
  if (result.error || !result.data) throw new Error('Persisted logs are unavailable.')
  return result.data
}

export async function loadProjectMetrics(projectId: string): Promise<Array<RuntimeMetricSample>> {
  const result = await getProjectMetrics({ path: { projectId } })
  if (result.error || !result.data) throw new Error('Project resource samples are unavailable.')
  return result.data
}

export async function loadEffectiveManifest(projectId: string): Promise<EffectiveManifest> {
  const result = await explainProjectManifest({ path: { projectId } })
  if (result.error || !result.data) throw new Error('The effective manifest is unavailable.')
  return result.data
}

export async function loadProjectGit(projectId: string, signal?: AbortSignal): Promise<GitState> {
  const result = await getProjectGit({ path: { projectId }, signal })
  if (result.error || !result.data) throw new Error('Git state is unavailable.')
  return result.data
}

export async function loadProjectGitInventory(
  projectId: string,
  signal?: AbortSignal,
): Promise<GitState> {
  const result = await getProjectGitInventory({ path: { projectId }, signal })
  if (result.error || !result.data)
    throw new Error('The repository file and branch inventory could not be loaded.')
  return result.data
}

export async function loadProjectGitStatus(
  projectId: string,
  signal?: AbortSignal,
): Promise<GitState> {
  const result = await getProjectGitStatus({ path: { projectId }, signal })
  if (result.error || !result.data) throw new Error('Git status is unavailable.')
  return result.data
}

export async function loadProjectActions(
  projectId: string,
  signal?: AbortSignal,
): Promise<ProjectActions> {
  const result = await listProjectActions({ path: { projectId }, signal })
  if (result.error || !result.data) throw new Error('Project actions are unavailable.')
  return result.data
}

export async function runProjectAction(
  projectId: string,
  actionId: string,
  confirmRisk = false,
): Promise<Operation> {
  const result = await createActionOperation({
    path: { projectId, actionId },
    body: { confirmRisk, allowOutsideRoot: false },
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data) {
    throw new ProjectAPIError(
      result.error?.detail || 'The project action could not be queued.',
      result.error?.code,
    )
  }
  return result.data
}

export async function runRuntimeAction(
  projectId: string,
  action: RuntimeAction,
  selection: RuntimeSelection = {},
): Promise<Operation> {
  const result = await createProjectOperation({
    path: { projectId },
    body: runtimeRequest(action, selection),
    headers: mutationHeaders(requestKey()) as { 'Idempotency-Key': string },
  })
  if (result.error || !result.data) throw new Error(`The ${action} operation could not be queued.`)
  return result.data
}

export type RuntimeSelection = Pick<
  RuntimeActionRequest,
  | 'profiles'
  | 'services'
  | 'removeVolumes'
  | 'confirmRisk'
  | 'expectedCheckoutRevision'
  | 'expectedManifestHash'
>

export async function previewRuntimeAction(
  projectId: string,
  action: RuntimeAction,
  selection: RuntimeSelection = {},
): Promise<RuntimePlan> {
  const result = await planProjectRuntime({
    path: { projectId },
    body: runtimeRequest(action, selection),
  })
  if (result.error || !result.data) throw new Error(`The ${action} plan could not be prepared.`)
  return result.data
}

export async function exportProjectLogText(
  projectId: string,
  service = '',
  format: 'plain' | 'ndjson' = 'plain',
): Promise<string> {
  const result = await exportProjectLogsRequest({
    path: { projectId },
    query: { format, ...(service ? { service } : {}) },
  })
  if (result.error || result.data === undefined)
    throw new Error('The log export could not be prepared.')
  return result.data
}

function runtimeRequest(action: RuntimeAction, selection: RuntimeSelection) {
  const profiles = selection.profiles?.filter(Boolean) ?? []
  const services = selection.services?.filter(Boolean) ?? []
  return {
    action,
    removeVolumes: selection.removeVolumes ?? false,
    confirmRisk: selection.confirmRisk ?? false,
    ...(selection.expectedCheckoutRevision
      ? { expectedCheckoutRevision: selection.expectedCheckoutRevision }
      : {}),
    ...(selection.expectedManifestHash
      ? { expectedManifestHash: selection.expectedManifestHash }
      : {}),
    ...(profiles.length > 0 ? { profiles } : {}),
    ...(services.length > 0 ? { services } : {}),
  }
}
