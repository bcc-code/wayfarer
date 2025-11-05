import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: '../gql/*.graphqls',
  documents: ['./**/*.vue', './**/*.ts', './**/*.graphql', './**/*.gql'],
  ignoreNoDocuments: true, // for better experience with the watcher
  verbose: true,
  generates: {
    './app/api/generated.ts': {
      plugins: ['typescript', 'typescript-operations', 'typescript-vue-urql'],
      config: {
        withComposition: true,
        useTypeImports: true,
      },
    },
  },
}

export default config
