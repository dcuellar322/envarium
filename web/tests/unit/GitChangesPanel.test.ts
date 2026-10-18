import { render, screen, within } from '@testing-library/vue'
import { expect, test } from 'vitest'
import GitChangesPanel from '../../src/domains/projects/components/GitChangesPanel.vue'
import BranchSwitchDialog from '../../src/domains/projects/components/BranchSwitchDialog.vue'
import { fireEvent } from '@testing-library/vue'

test('partial staging appears in both groups and paths remain literal', () => {
  render(GitChangesPanel, {
    props: {
      details: {
        files: [
          {
            path: 'space and\nnewline',
            originalPath: 'old name',
            index: 'M',
            workingTree: 'M',
            untracked: false,
            conflicted: false,
          },
          { path: 'new', index: '?', workingTree: '?', untracked: true, conflicted: false },
          { path: 'conflict', index: 'U', workingTree: 'U', untracked: false, conflicted: true },
        ],
        branches: [],
        filesTruncated: true,
        branchesTruncated: false,
      },
    },
  })
  for (const group of ['Staged', 'Unstaged'])
    expect(
      within(screen.getByRole('region', { name: group })).getByText(/space and/),
    ).toBeInTheDocument()
  expect(
    within(screen.getByRole('region', { name: 'Untracked' })).getByText('new'),
  ).toBeInTheDocument()
  expect(
    within(screen.getByRole('region', { name: 'Conflicts' })).getByText('conflict'),
  ).toBeInTheDocument()
  expect(screen.getByRole('status')).toHaveTextContent('first 1,024')
})
test('branch confirmation is explicit and resets for a changed plan', async () => {
  const plan = {
    projectId: 'p',
    root: '/repo',
    fromBranch: 'main',
    fromHead: 'abc',
    toBranch: 'feature',
    toHead: 'def',
    risk: 'mutating' as const,
    reviewRestart: true,
  }
  const view = render(BranchSwitchDialog, { props: { plan, pending: false } })
  const button = screen.getByRole('button', { name: 'Switch branch' })
  expect(button).toBeDisabled()
  await fireEvent.click(screen.getByRole('checkbox'))
  expect(button).toBeEnabled()
  await view.rerender({ plan: { ...plan, toHead: 'changed' }, pending: false })
  expect(button).toBeDisabled()
  await fireEvent.click(screen.getByRole('checkbox'))
  await fireEvent.click(button)
  expect(view.emitted().confirm).toHaveLength(1)
})
