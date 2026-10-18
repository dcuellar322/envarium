import { createGitBranchSwitch, planGitBranchSwitch } from '../../api/generated/sdk.gen'
import type { GitBranchSwitchPlan, Operation } from '../../api/generated/types.gen'
import { mutationHeaders } from '../session/bootstrap'
function headers() {
  return mutationHeaders(`ui_${crypto.randomUUID()}`) as { 'Idempotency-Key': string }
}
export async function previewBranchSwitch(
  projectId: string,
  branch: string,
): Promise<GitBranchSwitchPlan> {
  const result = await planGitBranchSwitch({
    path: { projectId },
    body: { branch },
    headers: headers(),
  })
  if (result.error || !result.data)
    throw new Error(result.error?.detail || 'The branch switch could not be prepared.')
  return result.data
}
export async function switchBranch(
  projectId: string,
  plan: GitBranchSwitchPlan,
  confirmRisk: boolean,
): Promise<Operation> {
  const result = await createGitBranchSwitch({
    path: { projectId },
    body: { plan, confirmRisk },
    headers: headers(),
  })
  if (result.error || !result.data)
    throw new Error(result.error?.detail || 'The branch switch could not be queued.')
  return result.data
}
