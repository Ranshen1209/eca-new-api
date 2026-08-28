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
export const ECA_FAVICON_HREF = '/eca-favicon.svg'
export const ECA_MARK_HREF = '/eca-mark.svg'

export function resolveEcaFaviconHref(logoUrl: string): string {
  if (!logoUrl) {
    return ECA_FAVICON_HREF
  }

  try {
    const pathname = new URL(logoUrl, 'http://local.invalid').pathname
    if (pathname.endsWith('/eca-logo.png')) {
      return ECA_FAVICON_HREF
    }
  } catch {
    if (logoUrl.includes('eca-logo.png')) {
      return ECA_FAVICON_HREF
    }
  }

  return logoUrl
}

export function resolveEcaCompactMarkHref(logoUrl: string): string {
  if (resolveEcaFaviconHref(logoUrl) === ECA_FAVICON_HREF) {
    return ECA_MARK_HREF
  }
  return logoUrl
}
