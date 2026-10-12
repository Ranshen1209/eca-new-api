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

import { describe, expect, test } from 'vitest'

import { DEFAULT_LOGO, DEFAULT_SYSTEM_NAME } from '@/lib/constants'

import {
  ECA_FAVICON_HREF,
  ECA_MARK_HREF,
  resolveEcaCompactMarkHref,
  resolveEcaFaviconHref,
} from '../favicon'
import { resolveEcaSystemName } from '../system-name'

const html = readFileSync(
  path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    '../../../../index.html'
  ),
  'utf8'
)

describe('ECA first-paint defaults', () => {
  test('uses the company name and wordmark before status loads', () => {
    expect(DEFAULT_SYSTEM_NAME).toBe('ECA API Gateway')
    expect(DEFAULT_LOGO).toBe('/eca-logo.png')
    expect(html).toContain('<title>ECA API Gateway</title>')
    expect(html).toContain('content="ECA API Gateway"')
  })

  test('keeps showing ECA API Gateway when the backend still has the upstream default name', () => {
    expect(resolveEcaSystemName('New API')).toBe('ECA API Gateway')
    expect(resolveEcaSystemName('ECA API')).toBe('ECA API Gateway')
    expect(resolveEcaSystemName('')).toBe('ECA API Gateway')
    expect(resolveEcaSystemName(undefined)).toBe('ECA API Gateway')
    expect(resolveEcaSystemName('My Gateway')).toBe('My Gateway')

    const brandingDir = path.dirname(fileURLToPath(import.meta.url))
    const configSource = readFileSync(
      path.resolve(brandingDir, '../../../lib/status-query.ts'),
      'utf8'
    )
    const brandSource = readFileSync(
      path.resolve(
        brandingDir,
        '../../../components/layout/components/system-brand.tsx'
      ),
      'utf8'
    )
    const mainSource = readFileSync(
      path.resolve(brandingDir, '../../../main.tsx'),
      'utf8'
    )
    expect(configSource).toContain(
      'resolveEcaSystemName(data.system_name as string | undefined)'
    )
    expect(brandSource).toContain('resolveEcaSystemName(status?.system_name')
    expect(mainSource).toContain('resolveEcaSystemName(s.system_name')
  })

  test('first-paint favicon uses the official no-tagline SVG, not a generated plus icon', () => {
    expect(html).toContain(`href="${ECA_FAVICON_HREF}"`)
    expect(html).toContain('type="image/svg+xml"')
    expect(html).not.toContain('href="/favicon.ico"')
    expect(html).not.toContain('href="/logo.png"')
  })

  test('first-paint favicon uses a square canvas so the official mark is not stretched', () => {
    const favicon = readFileSync(
      path.resolve(
        path.dirname(fileURLToPath(import.meta.url)),
        '../../../../public/eca-favicon.svg'
      ),
      'utf8'
    )
    expect(favicon).toContain('viewBox="0 0 512 512"')
    expect(favicon).toContain('preserveAspectRatio="xMidYMid meet"')
    expect(favicon).toContain('height="512"')
    expect(favicon).not.toContain('height="254"')

    const mark = readFileSync(
      path.resolve(
        path.dirname(fileURLToPath(import.meta.url)),
        '../../../../public/eca-mark.svg'
      ),
      'utf8'
    )
    expect(mark).toContain('viewBox="0 4 512 186"')
    expect(mark).toContain('width="44"')
    expect(mark).toContain('height="16"')
    expect(mark).not.toContain('width="512"')
    expect(mark).not.toContain('viewBox="0 0 512 512"')
    expect(favicon).toContain('viewBox="0 4 512 186"')
    expect(favicon).toContain('y="168"')
  })

  test('header wordmark PNG does not replace the tab icon', () => {
    expect(resolveEcaFaviconHref('/eca-logo.png')).toBe(ECA_FAVICON_HREF)
    expect(resolveEcaFaviconHref('https://cdn.example/eca-logo.png')).toBe(
      ECA_FAVICON_HREF
    )
    expect(resolveEcaFaviconHref('/custom-mark.png')).toBe('/custom-mark.png')
    expect(resolveEcaCompactMarkHref('/eca-logo.png')).toBe(ECA_MARK_HREF)
    expect(resolveEcaCompactMarkHref('/custom-mark.png')).toBe(
      '/custom-mark.png'
    )
  })
})
