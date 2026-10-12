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
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { render, screen } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { EcaLandingFrame } from '../landing-frame'

const brandingDir = path.dirname(fileURLToPath(import.meta.url))
const homeSource = readFileSync(
  path.resolve(brandingDir, '../../../features/home/index.tsx'),
  'utf8'
)

describe('EcaLandingFrame', () => {
  test('marks the landing as ECA branded and keeps children in front of the flow', () => {
    render(
      <EcaLandingFrame>
        <h1>Landing</h1>
      </EcaLandingFrame>
    )

    const root = document.querySelector('[data-brand="eca"]')
    expect(root).not.toBeNull()
    expect(root?.className).toContain('min-h-svh')
    const content = screen.getByRole('heading', {
      name: 'Landing',
    }).parentElement
    expect(root?.querySelector('.eca-optical-flow')).not.toBeNull()
    expect(content?.className).toContain('z-10')
  })

  test('default home landing stays an internal hero without marketing sections', () => {
    expect(homeSource).toContain('EcaLandingFrame')
    expect(homeSource).toContain('<Hero isAuthenticated={isAuthenticated} />')
    expect(homeSource).not.toMatch(/<Stats[\s/>]/)
    expect(homeSource).not.toMatch(/<Features[\s/>]/)
    expect(homeSource).not.toMatch(/<HowItWorks[\s/>]/)
    expect(homeSource).not.toMatch(/<CTA[\s/>]/)
  })
})
