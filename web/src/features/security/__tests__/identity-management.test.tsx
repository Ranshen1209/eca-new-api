import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { UserProfile } from '@/features/profile/types'
import { api } from '@/lib/api'

import { Security } from '..'

const { profileState } = vi.hoisted(() => ({
  profileState: { profile: null as UserProfile | null },
}))
vi.mock('@/features/profile/hooks/use-profile', () => ({
  useProfile: () => ({
    profile: profileState.profile,
    loading: false,
    refreshProfile: vi.fn(),
    fetchProfile: vi.fn(),
  }),
}))
vi.mock('@tanstack/react-router', async (original) => ({
  ...(await original<typeof import('@tanstack/react-router')>()),
  useNavigate: () => vi.fn(),
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

function renderSecurity(oidc: boolean) {
  profileState.profile = {
    id: 1,
    username: 'employee',
    display_name: 'Employee',
    role: 1,
    status: 1,
    email: 'employee@example.test',
    group: 'default',
    quota: 0,
    used_quota: 0,
    request_count: 0,
    created_time: 0,
    has_password: true,
    oidc_id: oidc ? 'synthetic-subject' : '',
    identity_managed: oidc,
  } as UserProfile
  const requests = vi.spyOn(api, 'get').mockImplementation(async (url) => {
    const path = String(url)
    let data: object = []
    if (path.endsWith('/catalog')) {
      data = { groups: [], max_tokens: 10 }
    } else if (path === '/api/user/access_tokens') {
      data = { items: [], legacy: null }
    } else if (path.endsWith('/2fa/status')) {
      data = { enabled: false, locked: false }
    } else if (path.endsWith('/passkey')) {
      data = { enabled: false, registered: false }
    } else if (path === '/api/status') {
      data = {}
    }
    return { data: { success: true, data } }
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={client}>
      <Security />
    </QueryClientProvider>
  )
  return requests
}

it('renders the OIDC identity notice, keeps sessions/access/preferences and never mounts local enrollment controls', async () => {
  const requests = renderSecurity(true)
  expect(screen.getByRole('note')).toHaveTextContent('identity provider')
  for (const name of [
    'Change Password',
    'Set Password',
    'Delete Account',
    'Account Bindings',
    'Two-Factor Authentication',
  ]) {
    expect(screen.queryByText(name)).not.toBeInTheDocument()
  }
  expect(screen.getByText('Sessions & Access')).toBeVisible()
  expect(screen.getByText('Privacy')).toBeVisible()
  await waitFor(() => expect(requests).toHaveBeenCalled())
  expect(
    requests.mock.calls.some(([url]) =>
      /\/(passkey|2fa|oauth\/bindings)/.test(String(url))
    )
  ).toBe(false)
})

it('keeps the operational password dialog keyboard-accessible for an unbound account', async () => {
  renderSecurity(false)
  const user = userEvent.setup()
  const button = screen.getByRole('button', { name: 'Change Password' })
  button.focus()
  await user.keyboard('{Enter}')
  expect(await screen.findByRole('dialog')).toBeVisible()
})
