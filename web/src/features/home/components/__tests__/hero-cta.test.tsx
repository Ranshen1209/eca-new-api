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
import { render, screen } from '@testing-library/react'
import i18n from 'i18next'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import { applyEcaLandingCopy } from '@/branding/eca/copy'

vi.mock('@tanstack/react-router', () => ({
  Link: (props: { to: string; children?: ReactNode }) => (
    <a href={props.to}>{props.children}</a>
  ),
}))

vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({
    status: { docs_link: 'https://docs.eca-healthcare.org/' },
  }),
}))

vi.mock('../hero-terminal-demo', () => ({
  HeroTerminalDemo: () => <div data-testid='hero-terminal-demo' />,
}))

vi.mock('@lobehub/icons', () => ({
  CherryStudio: {
    Color: () => <span data-testid='cherry-studio-icon' />,
  },
}))

const { Hero } = await import('../sections/hero')

describe('Hero CTAs', () => {
  afterEach(async () => {
    await i18n.changeLanguage('en')
  })

  test('keeps the View Models button when the visitor is signed in', async () => {
    applyEcaLandingCopy(i18n)
    await i18n.changeLanguage('zhCN')

    render(<Hero isAuthenticated />)

    expect(
      screen.getByRole('link', { name: 'Go to Dashboard' })
    ).toHaveAttribute('href', '/dashboard')
    expect(screen.getByRole('link', { name: '查看模型' })).toHaveAttribute(
      'href',
      '/pricing'
    )
    expect(screen.queryByRole('link', { name: 'Get Started' })).toBeNull()
  })

  test('keeps the View Models button for guests next to Get Started', async () => {
    applyEcaLandingCopy(i18n)
    await i18n.changeLanguage('zhCN')

    render(<Hero isAuthenticated={false} />)

    expect(screen.getByRole('link', { name: 'Get Started' })).toHaveAttribute(
      'href',
      '/sign-up'
    )
    expect(screen.getByRole('link', { name: '查看模型' })).toHaveAttribute(
      'href',
      '/pricing'
    )
    expect(screen.queryByRole('link', { name: 'Go to Dashboard' })).toBeNull()
  })
})
