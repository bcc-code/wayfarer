import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: 'http://localhost:4000/graphql',
  documents: ['./**/*.vue', './**/*.graphql', './**/*.gql'],
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
