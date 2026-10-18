import { pathToFileURL } from 'node:url'

export async function verifyReleaseQuality({ repository, sha, token, fetcher = fetch }) {
  if (!/^[\w.-]+\/[\w.-]+$/.test(repository ?? '') || !/^[a-f0-9]{40}$/.test(sha ?? '') || !token) {
    throw new Error(
      'Release quality requires repository, exact commit SHA, and an Actions read token',
    )
  }
  for (const workflow of ['ci.yml', 'site-ci.yml', 'security.yml']) {
    const response = await fetcher(
      `https://api.github.com/repos/${repository}/actions/workflows/${workflow}/runs?head_sha=${sha}&event=push&per_page=100`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
          Accept: 'application/vnd.github+json',
          'X-GitHub-Api-Version': '2022-11-28',
        },
      },
    )
    if (!response.ok) throw new Error(`Cannot verify ${workflow}: HTTP ${response.status}`)
    const runs = (await response.json()).workflow_runs
      .filter((run) => run.head_sha === sha && ['master', 'main'].includes(run.head_branch))
      .sort((left, right) => right.id - left.id)
    if (!runs[0] || runs[0].status !== 'completed' || runs[0].conclusion !== 'success') {
      throw new Error(
        `${workflow} must pass on the exact release commit before packaging; finish or rerun quality CI first`,
      )
    }
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  await verifyReleaseQuality({
    repository: process.env.GITHUB_REPOSITORY,
    sha: process.env.RELEASE_SHA,
    token: process.env.GITHUB_TOKEN,
  })
  console.log('Exact-commit Go, desktop, browser, full site and security quality workflows passed.')
}
