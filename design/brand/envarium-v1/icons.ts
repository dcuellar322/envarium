export const brandIconNames = [
  'projects',
  'workspaces',
  'services',
  'ports',
  'logs',
  'terminal',
  'resources',
  'agents',
  'git',
  'health',
  'operations',
  'settings',
] as const

export type BrandIconName = (typeof brandIconNames)[number]
