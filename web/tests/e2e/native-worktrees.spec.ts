import { expect, test, type Page } from '@playwright/test'
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync, rmSync } from 'node:fs'
import { resolve, join } from 'node:path'
import { browserBootstrapPath } from '../helpers/browserSession'

async function api(page: Page, path: string, body?: unknown) {
  return page.evaluate(
    async ({ path, body }) => {
      const response = await fetch(`/api/v1/${path}`, {
        method: body ? 'POST' : 'GET',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': sessionStorage.getItem('envarium.csrf-token') || '',
          'Idempotency-Key': crypto.randomUUID(),
        },
        ...(body ? { body: JSON.stringify(body) } : {}),
      })
      const result = await response.json()
      if (!response.ok) throw new Error(JSON.stringify(result))
      return result
    },
    { path, body },
  )
}
async function stopped(page: Page, id: string) {
  const plan = await api(page, `projects/${id}/runtime/plan`, { action: 'stop' })
  const operation = await api(page, `projects/${id}/operations`, {
    action: 'stop',
    confirmRisk: true,
    expectedCheckoutRevision: plan.checkoutRevision,
    expectedManifestHash: plan.manifestHash,
  })
  await expect
    .poll(async () => (await api(page, `operations/${operation.id}`)).state)
    .toBe('succeeded')
}

test('native worktrees isolate ports, logs and metrics and review stop/switch/start', async ({
  page,
}) => {
  test.setTimeout(150_000)
  const parent = resolve('../.cache')
  mkdirSync(parent, { recursive: true })
  const fixture = mkdtempSync(join(parent, 'native-worktrees-e2e-'))
  const main = join(fixture, 'main'),
    a = join(fixture, 'feature-a'),
    b = join(fixture, 'feature-b')
  mkdirSync(join(main, '.envarium'), { recursive: true })
  const git = (...args: string[]) =>
    execFileSync('git', ['-C', main, ...args], { encoding: 'utf8' })
  git('init', '--initial-branch=main')
  git('config', 'user.name', 'Envarium test')
  git('config', 'user.email', 'test@localhost')
  writeFileSync(
    join(main, 'package.json'),
    JSON.stringify({ name: 'Native worktree fixture', scripts: { dev: 'node server.mjs' } }),
  )
  writeFileSync(
    join(main, '.envarium/project.yml'),
    `schemaVersion: envarium.dev/v1
kind: Project
metadata: {id: native-worktrees, name: Native worktree fixture}
repository: {root: .}
runtime:
  driver: process
  process:
    processes:
      - id: web
        command: [${JSON.stringify(process.execPath)}, server.mjs, "\${ports.http}"]
services:
  - id: web
    source: {process: web}
    healthChecks:
      - {id: http, type: http, url: "http://127.0.0.1:\${ports.http}/", expectedStatus: 200, required: true, intervalSeconds: 10, timeoutSeconds: 2, retries: 3}
ports:
  - {id: http, service: web, host: 19873, target: 19873, protocol: tcp}
`,
  )
  const server = (name: string) =>
    `import http from 'node:http';http.createServer((_,res)=>res.end('${name}')).listen(Number(process.argv[2]),'127.0.0.1',()=>console.log('${name} listening'));`
  writeFileSync(join(main, 'server.mjs'), server('main'))
  git('add', '.')
  git('commit', '-m', 'Owned native fixture')
  git('branch', 'feature-a')
  git('branch', 'feature-b')
  git('switch', '-c', 'next')
  writeFileSync(join(main, 'server.mjs'), server('next'))
  git('add', '.')
  git('commit', '-m', 'Branch-specific response')
  git('switch', 'main')
  git('worktree', 'add', a, 'feature-a')
  git('worktree', 'add', b, 'feature-b')
  let environments: Array<{
    id: string
    path: string
    allocation: { portLeases: Array<{ hostPort: number }> }
  }> = []
  try {
    await page.goto(browserBootstrapPath())
    await page.getByRole('link', { name: 'Discovery' }).click()
    await page.getByLabel('Repository path').fill(main)
    await page.getByRole('button', { name: 'Scan repository' }).click()
    await page.getByRole('button', { name: 'Approve and trust project' }).click()
    await expect(page).toHaveURL(/\/projects\//)
    const projectId = new URL(page.url()).pathname.split('/').at(-1)!
    await page.getByRole('tab', { name: 'git', exact: true }).click()
    await page.getByRole('button', { name: /Reconcile worktrees/ }).click()
    await expect(page.getByRole('button', { name: 'Review start', exact: true })).toHaveCount(2)
    environments = await api(page, `projects/${projectId}/environments`)
    const envA = environments.find((item) => item.path === a)!,
      envB = environments.find((item) => item.path === b)!
    const portA = envA.allocation.portLeases[0]!.hostPort,
      portB = envB.allocation.portLeases[0]!.hostPort
    expect(portA).not.toBe(portB)
    const card = (name: string) =>
      page.getByRole('article', { name: `Worktree ${name}`, exact: true })
    for (const name of ['feature-a', 'feature-b']) {
      await card(name).getByRole('button', { name: 'Review start', exact: true }).click()
      await page.getByRole('button', { name: 'Confirm start', exact: true }).click()
      const id = name === 'feature-a' ? envA.id : envB.id
      await expect
        .poll(async () => (await api(page, `projects/${id}/runtime`)).state, { timeout: 20_000 })
        .toBe('running')
    }
    expect(await (await page.request.get(`http://127.0.0.1:${portA}`)).text()).toBe('main')
    expect(await (await page.request.get(`http://127.0.0.1:${portB}`)).text()).toBe('main')
    await expect
      .poll(async () => {
        const registry = await api(page, 'ports')
        return [envA.id, envB.id].every((id) =>
          registry.facts.some(
            (fact: { projectId?: string; kind: string }) =>
              fact.projectId === id && fact.kind === 'binding',
          ),
        )
      })
      .toBe(true)
    const cliStatus = JSON.parse(
      execFileSync(
        '../.cache/envarium-e2e',
        [
          '--address',
          process.env.ENVARIUM_E2E_DAEMON_ADDRESS || '127.0.0.1:29616',
          '--data-dir',
          '../.envarium-data/e2e',
          '--json',
          'status',
          envA.id,
        ],
        { encoding: 'utf8' },
      ),
    )
    expect(cliStatus.data.projectId).toBe(envA.id)
    expect(cliStatus.data.state).toBe('running')
    const runB = (await api(page, `projects/${envB.id}/runtime`)).services[0].process.runId
    await expect
      .poll(async () => JSON.stringify(await api(page, `projects/${envA.id}/logs`)))
      .toContain('main listening')
    await expect
      .poll(async () =>
        (await api(page, `projects/${envA.id}/metrics`)).some(
          (item: { memoryAvailable: boolean }) => item.memoryAvailable,
        ),
      )
      .toBe(true)
    await page.getByLabel('Inspect checkout').selectOption(envA.id)
    const next = page
      .locator('.git-branches li')
      .filter({ has: page.getByText('next', { exact: true }) })
    await next.getByRole('button', { name: 'Review switch' }).click()
    await page.getByRole('button', { name: 'Confirm stop', exact: true }).click()
    await expect(page.getByRole('dialog', { name: 'Switch to next' })).toBeVisible()
    await page.getByRole('checkbox', { name: /I reviewed this checkout/ }).check()
    await page.getByRole('button', { name: 'Switch branch', exact: true }).click()
    await expect(page.getByRole('dialog', { name: /start 1 native process/ })).toBeVisible()
    expect(git('-C', a, 'branch', '--show-current').trim()).toBe('next')
    await page.getByRole('button', { name: 'Confirm start', exact: true }).click()
    await expect
      .poll(async () => {
        try {
          return await (await page.request.get(`http://127.0.0.1:${portA}`)).text()
        } catch {
          return 'waiting for listener'
        }
      })
      .toBe('next')
    expect((await api(page, `projects/${envB.id}/runtime`)).services[0].process.runId).toBe(runB)
    await card('feature-b').getByRole('button', { name: 'Switch services here…' }).click()
    await page.getByRole('button', { name: 'Confirm stop', exact: true }).click()
    await page.getByRole('button', { name: 'Confirm restart', exact: true }).click()
    await expect
      .poll(async () => (await api(page, `projects/${envA.id}/runtime`)).state)
      .toBe('stopped')
    await expect
      .poll(async () => (await api(page, `projects/${envB.id}/runtime`)).services[0].process.runId)
      .not.toBe(runB)
    await expect(page.getByLabel('Inspect checkout')).toHaveValue(envB.id)
    await expect(page.getByText('Checkout transition completed.')).toBeVisible()
    await page.getByRole('heading', { name: 'Native worktree fixture', exact: true }).click()
    await page.screenshot({
      path: '../output/playwright/native-worktree-workflow.png',
      fullPage: true,
    })
  } finally {
    for (const environment of environments) await stopped(page, environment.id)
    rmSync(fixture, { recursive: true, force: true })
  }
})
