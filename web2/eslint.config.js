import stylistic from '@stylistic/eslint-plugin'
import vuetify from 'eslint-config-vuetify'

export default vuetify(
  {
    ts: true,
  },
  {
    plugins: {
      '@stylistic': stylistic,
    },
    rules: {
      '@stylistic/space-before-function-paren': ['error', 'never'],
    },
  },
)
