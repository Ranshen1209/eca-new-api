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

const tokensCss = readFileSync(
  path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../tokens.css'),
  'utf8'
)

describe('ECA overlay tokens', () => {
  test('stops ambient wash when the user prefers reduced motion', () => {
    expect(tokensCss).toMatch(/prefers-reduced-motion:\s*reduce/)
    expect(tokensCss).toMatch(
      /@media \(prefers-reduced-motion: reduce\)[\s\S]*\.eca-optical-flow::before[\s\S]*animation:\s*none/
    )
    expect(tokensCss).not.toContain('eca-flow-dash')
    expect(tokensCss).toMatch(
      /\.eca-optical-flow\s*\{[^}]*inset:\s*0/
    )
  })

  test('hides the footer tagline on the company landing', () => {
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] footer a\.group \+ p[\s\S]*display:\s*none/
    )
  })

  test('tightens the company landing footer so leftover sections do not leave a tall blank band', () => {
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] footer > div[\s\S]*padding-bottom:\s*1\.25rem/
    )
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] footer \.shrink-0[\s\S]*display:\s*none/
    )
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] section\.relative\.z-10[\s\S]*flex:\s*1/
    )
  })

  test('does not crop the company wordmark into a circle', () => {
    expect(tokensCss).toMatch(/img\[src\*='eca-logo\.png'\][\s\S]*border-radius:\s*0/)
    expect(tokensCss).toMatch(/img\[src\*='eca-logo\.png'\][\s\S]*object-fit:\s*contain/)
    expect(tokensCss).toMatch(
      /div:has\(> img\[src\*='eca-logo\.png'\]\)[\s\S]*border-radius:\s*0/
    )
    expect(tokensCss).toMatch(
      /img\[src\*='eca-logo\.png'\][\s\S]*box-shadow:\s*none/
    )
    expect(tokensCss).toMatch(
      /img\[src\*='eca-logo\.png'\][\s\S]*background-color:\s*transparent/
    )
  })

  test('aligns the compact no-tagline mark with adjacent type', () => {
    expect(tokensCss).toMatch(
      /img\[src\*='eca-mark\.svg'\][\s\S]*object-fit:\s*contain/
    )
    expect(tokensCss).toMatch(
      /img\[src\*='eca-mark\.svg'\][\s\S]*height:\s*1em/
    )
    expect(tokensCss).toMatch(
      /img\[src\*='eca-mark\.svg'\][\s\S]*width:\s*2\.75em/
    )
    expect(tokensCss).toMatch(
      /div:has\(> img\[src\*='eca-mark\.svg'\]\)[\s\S]*align-items:\s*center/
    )
    expect(tokensCss).toMatch(
      /a:has\(img\[src\*='eca-mark\.svg'\]\) span[\s\S]*line-height:\s*1/
    )
    expect(tokensCss).toMatch(
      /header a:has\(img\[src\*='eca-mark\.svg'\]\)[\s\S]*overflow:\s*visible/
    )
    expect(tokensCss).toMatch(
      /\[data-eca-about='page'\] img\[src\*='eca-mark\.svg'\][\s\S]*width:\s*3\.5rem/
    )
  })

  test('keeps the four API demos on one column width so auto-switching does not jump the hero', () => {
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] section\.relative\.z-10 > div\.grid\s*\{[\s\S]*width:\s*100%/
    )
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] section\.relative\.z-10 > div\.grid > \*[\s\S]*min-width:\s*0/
    )
    expect(tokensCss).toMatch(
      /grid-template-columns:\s*repeat\(12,\s*minmax\(0,\s*1fr\)\)/
    )
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] \[data-eca-api-demo\][\s\S]*min-width:\s*0/
    )
  })

  test('keeps the landing hero headline from being selected', () => {
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] section\.relative\.z-10 h1[\s\S]*user-select:\s*none/
    )
    expect(tokensCss).toMatch(
      /\[data-brand='eca'\] section\.relative\.z-10 :has\(\+ h1\)[\s\S]*user-select:\s*none/
    )
  })

  test('keeps gold as a spark instead of a fill color', () => {
    expect(tokensCss).toContain('--eca-gold: #ffd400')
    expect(tokensCss).toContain('--primary: #01a568')
    expect(tokensCss).not.toMatch(/:root\s*\{[^}]*--primary:\s*#ffd400/)
  })
})
