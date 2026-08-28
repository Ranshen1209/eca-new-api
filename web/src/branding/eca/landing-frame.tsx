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
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { applyEcaLandingCopy } from './copy'
import { EcaOpticalFlow } from './optical-flow'

interface EcaLandingFrameProps {
  children: ReactNode
}

export function EcaLandingFrame(props: EcaLandingFrameProps) {
  const { i18n } = useTranslation()
  applyEcaLandingCopy(i18n)

  return (
    <div data-brand='eca' className='relative flex min-h-svh flex-col'>
      <EcaOpticalFlow />
      <div className='relative z-10 flex min-h-svh flex-1 flex-col'>
        {props.children}
      </div>
    </div>
  )
}
