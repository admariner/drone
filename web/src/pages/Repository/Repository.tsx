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

import React, { useEffect, useState } from 'react'
import { Container, Layout, PageBody, StringSubstitute, Text } from '@harnessio/uicore'
import { Falsy, Match, Truthy } from 'react-jsx-match'
import cx from 'classnames'
import { useGetResourceContent } from 'hooks/useGetResourceContent'
import { voidFn, getErrorMessage } from 'utils/Utils'
import { useGetRepositoryMetadata } from 'hooks/useGetRepositoryMetadata'
import { LoadingSpinner } from 'components/LoadingSpinner/LoadingSpinner'
import { useStrings } from 'framework/strings'
import type { OpenapiGetContentOutput, TypesRepository } from 'services/code'
import { Images } from 'images'
import { RepositoryContent } from './RepositoryContent/RepositoryContent'
import { RepositoryHeader } from './RepositoryHeader/RepositoryHeader'
import { ContentHeader } from './RepositoryContent/ContentHeader/ContentHeader'
import { EmptyRepositoryInfo } from './EmptyRepositoryInfo'
import css from './Repository.module.scss'

export default function Repository() {
  const { gitRef, resourcePath, repoMetadata, error, loading, refetch, commitRef } = useGetRepositoryMetadata()
  const {
    data: resourceContent,
    error: resourceError,
    loading: resourceLoading,
    isRepositoryEmpty
  } = useGetResourceContent({ repoMetadata, gitRef, resourcePath, includeCommit: true })
  const [fileNotExist, setFileNotExist] = useState(false)
  const { getString } = useStrings()

  useEffect(() => setFileNotExist(resourceError?.status === 404), [resourceError])

  return (
    <Container className={cx(css.main, !!resourceContent && css.withFileViewer)}>
      <Match expr={fileNotExist}>
        <Truthy>
          <RepositoryHeader repoMetadata={repoMetadata as TypesRepository} />
          <Layout.Vertical>
            <Container className={css.bannerContainer} padding={{ left: 'xlarge' }}>
              <Text font={'small'} padding={{ left: 'large' }}>
                <StringSubstitute
                  str={getString('branchDoesNotHaveFile')}
                  vars={{
                    repoName: repoMetadata?.uid,
                    fileName: resourcePath,
                    branchName: gitRef
                  }}
                />
              </Text>
            </Container>
            <Container padding={{ left: 'xlarge' }}>
              <ContentHeader
                repoMetadata={repoMetadata as TypesRepository}
                gitRef={gitRef}
                resourcePath={resourcePath}
                resourceContent={resourceContent as OpenapiGetContentOutput}
              />
            </Container>
            <PageBody
              noData={{
                when: () => fileNotExist === true,
                message: getString('error404Text'),
                image: Images.error404
              }}></PageBody>
          </Layout.Vertical>
        </Truthy>
        <Falsy>
          <PageBody error={getErrorMessage(error || resourceError)} retryOnError={voidFn(refetch)}>
            <LoadingSpinner visible={loading || resourceLoading} withBorder={!!resourceContent && resourceLoading} />

            {!!repoMetadata && (
              <>
                <RepositoryHeader repoMetadata={repoMetadata} />

                {!!resourceContent && (
                  <RepositoryContent
                    repoMetadata={repoMetadata}
                    gitRef={gitRef}
                    resourcePath={resourcePath}
                    resourceContent={resourceContent}
                    commitRef={commitRef}
                  />
                )}

                {isRepositoryEmpty && <EmptyRepositoryInfo repoMetadata={repoMetadata} />}
              </>
            )}
          </PageBody>
        </Falsy>
      </Match>
    </Container>
  )
}
