/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, test } from 'vitest'

const { QueryClient, QueryClientProvider } =
  await import('@tanstack/react-query')
const { api } = await import('@/lib/api')
const { CCSwitchDialog } = await import('../cc-switch-dialog')

type ApiMethod = (url: string) => Promise<{ data: unknown }>
type MockableApi = { get: ApiMethod }

const apiClient = api as unknown as MockableApi
const originalGet = apiClient.get

function renderDialog() {
  apiClient.get = async (url) => {
    if (url === '/api/user/models') {
      return { data: { success: true, data: [] } }
    }
    throw new Error(`Unexpected GET ${url}`)
  }

  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  const view = render(
    <QueryClientProvider client={queryClient}>
      <CCSwitchDialog
        open
        onOpenChange={() => undefined}
        tokenKey='sk-test'
      />
    </QueryClientProvider>
  )

  return { queryClient, ...view }
}

afterEach(() => {
  apiClient.get = originalGet
})

describe('CC Switch import default names', () => {
  test('fills the name with ECA API Claude when the dialog opens', () => {
    const { queryClient } = renderDialog()

    expect(screen.getByDisplayValue('ECA API Claude')).toBeInTheDocument()
    expect(screen.queryByDisplayValue('My Claude')).not.toBeInTheDocument()

    queryClient.clear()
  })

  test('updates the name to ECA API Codex or Gemini when that application is selected', () => {
    const { queryClient } = renderDialog()

    fireEvent.click(screen.getByRole('radio', { name: 'Codex' }))
    expect(screen.getByDisplayValue('ECA API Codex')).toBeInTheDocument()
    expect(screen.queryByDisplayValue('ECA API Claude')).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('radio', { name: 'Gemini' }))
    expect(screen.getByDisplayValue('ECA API Gemini')).toBeInTheDocument()
    expect(screen.queryByDisplayValue('ECA API Codex')).not.toBeInTheDocument()

    queryClient.clear()
  })
})
