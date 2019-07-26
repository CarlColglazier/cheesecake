module.exports = {
  root: true,
  env: {
    browser: true,
    node: true
  },
  parserOptions: {
    parser: 'babel-eslint'
  },
  extends: [
    'prettier',
    'prettier/vue',
    'plugin:prettier/recommended',
    '@nuxtjs',
    'plugin:nuxt/recommended'
  ],
  plugins: [
		'prettier'
  ],
  rules: {
			'indent': [2, 'tab', { 'SwitchCase': 1, 'VariableDeclarator': 1 }],
			'no-tabs': 0,
			'vue/html-indent': ['error', 'tab', {
					'attribute': 1,
					'baseIndent': 1,
					'closeBracket': 0,
					'alignAttributesVertically': true,
					'ignores': []
			}]
  }
}
