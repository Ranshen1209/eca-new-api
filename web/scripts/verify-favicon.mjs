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
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const html = readFileSync(
  process.argv[2] || new URL('../dist/index.html', import.meta.url),
  'utf8'
)
const icons = [...html.matchAll(/<link\b[^>]*>/gi)].filter((match) => {
  const rel = match[0].match(/\brel\s*=\s*["']([^"']*)["']/i)?.[1]
  return rel?.toLowerCase().split(/\s+/).includes('icon')
})

assert.equal(icons.length, 1, 'Built HTML must contain exactly one tab icon')
assert.match(icons[0][0], /\bhref\s*=\s*["']\/eca-favicon\.svg["']/i)
