/*
 * Copyright 2023 Harness, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useState } from 'react'
import { Button, ButtonVariation, Container, Layout, Text } from '@harnessio/uicore'
import { Color, FontVariation } from '@harnessio/design-system'
import { useStrings } from 'framework/strings'
import { CopyButton } from 'components/CopyButton/CopyButton'
import { CodeIcon } from 'utils/GitUtils'
import CloneCredentialDialog from 'components/CloneCredentialDialog/CloneCredentialDialog'
import css from './CloneButtonTooltip.module.scss'

interface CloneButtonTooltipProps {
  httpsURL: string
}

export function CloneButtonTooltip({ httpsURL }: CloneButtonTooltipProps) {
  const { getString } = useStrings()
  const [flag, setFlag] = useState(false)

  return (
    <Container className={css.container} padding="xlarge">
      <Layout.Vertical spacing="small">
        <Text font={{ variation: FontVariation.H4 }}>{getString('cloneHTTPS')}</Text>
        <Text
          icon={'code-info'}
          iconProps={{ size: 16 }}
          color={Color.GREY_700}
          font={{ variation: FontVariation.BODY2_SEMI, size: 'small' }}>
          {getString('generateCloneText')}
        </Text>

        <Container>
          <Layout.Horizontal className={css.layout}>
            <Text className={css.url}>{httpsURL}</Text>

            <CopyButton content={httpsURL} id={css.cloneCopyButton} icon={CodeIcon.Copy} iconProps={{ size: 14 }} />
          </Layout.Horizontal>
        </Container>
        <Button
          onClick={() => {
            setFlag(true)
          }}
          variation={ButtonVariation.SECONDARY}>
          {getString('generateCloneCred')}
        </Button>
      </Layout.Vertical>
      <CloneCredentialDialog flag={flag} setFlag={setFlag} />
    </Container>
  )
}
