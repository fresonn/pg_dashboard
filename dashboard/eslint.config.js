import globals from 'globals'
import js from '@eslint/js'
import pluginReact from 'eslint-plugin-react'
import tseslint from '@typescript-eslint/eslint-plugin'
import parserTs from '@typescript-eslint/parser'
import eslintPluginPrettierRecommended from 'eslint-plugin-prettier/recommended'
import queryPlugin from '@tanstack/eslint-plugin-query'

/** @type {import('eslint').Linter.Config[]} */
export default [
  js.configs.recommended,
  eslintPluginPrettierRecommended,

  // General configuration for all files
  {
    files: ['**/*.{ts,tsx,js,jsx}'],
    languageOptions: {
      parser: parserTs,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true
        }
      },
      globals: {
        ...globals.browser,
        ...globals.node,
        ...globals.es2021
      }
    },
    ignores: ['**/.kubb/**', 'dist/**', 'pnpm-lock.yaml'],
    plugins: {
      react: pluginReact,
      'react-hooks': pluginReact,
      '@typescript-eslint': tseslint,
      '@tanstack/query': queryPlugin
    },
    rules: {
      'prettier/prettier': ['warn', {}, { usePrettierrc: true }],
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': [
        'warn',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_'
        }
      ]
    },
    settings: {
      react: {
        version: 'detect'
      }
    }
  }
]
