/**
 * plugins/vuetify.js
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import '@fortawesome/fontawesome-free/css/all.css'

// Composables
import { createVuetify } from 'vuetify'
import { aliases, fa } from 'vuetify/iconsets/fa'
import { mdi } from 'vuetify/iconsets/mdi'
import { VNumberInput } from 'vuetify/labs/VNumberInput'

const rankColors = {
  admiral: '#1E52E6',
  commander: '#308CA7',
  lieutenant: '#24AD32',
  specialist: '#DA5C5C',
  technician: '#E69737',
  member: '#FFC900',
  recruit: '#1CFAC0',
  guest: '#929292',
  ally: '#F87847',
  bot: '#206694',
}

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  components: {
    VNumberInput,
  },
  icons: {
    defaultSet: 'fa',
    aliases,
    sets: {
      fa,
      mdi
    },
  },
  theme: {
    defaultTheme: 'dark',
    variations: {
      colors: ['surface', 'commander'],
      lighten: 4,
      darken: 4
    },
    themes: {
      'light': {
        dark: false,
        colors: {
          primary: '#7a5900',
          'on-primary': '#ffffff',
          'primary-container': '#ffdea1',
          'on-primary-container': '#261900',
          secondary: '#4c6700',
          'on-secondary': '#ffffff',
          'secondary-container': '#baf600',
          'on-secondary-container': '#151f00',
          tertiary: '#b32a00',
          'on-tertiary': '#ffffff',
          'tertiary-container': '#ffdbd2',
          'on-tertiary-container': '#3c0800',
          error: '#ba1a1a',
          'on-error': '#ffffff',
          'error-container': '#ffdad6',
          'on-error-container': '#410002',
          background: '#fffbff',
          'on-background': '#1e1b16',
          surface: '#fffbff',
          'on-surface': '#1e1b16',
          outline: '#7f7667',
          'surface-variant': '#ede1cf',
          'on-surface-variant': '#4d4639',

          'surface-bright': '#FFFFFF',
          'surface-light': '#EEEEEE',
          'discord-primary': '#5865F2',
          info: '#2196F3',
          success: '#4CAF50',
          warning: '#FB8C00',
          ...rankColors
        }
      },
      'dark': {
        dark: true,
        colors: {
          primary: '#fcbc00',
          'on-primary': '#402d00',
          'primary-container': '#5c4300',
          'on-primary-container': '#ffdea1',
          secondary: '#a3d800',
          'on-secondary': '#263500',
          'secondary-container': '#394e00',
          'on-secondary-container': '#baf600',
          tertiary: '#ffb4a2',
          'on-tertiary': '#611200',
          'tertiary-container': '#891e00',
          'on-tertiary-container': '#ffdbd2',
          error: '#ffb4ab',
          'on-error': '#690005',
          'error-container': '#93000a',
          'on-error-container': '#ffdad6',
          background: '#001d31',
          'on-background': '#cde5ff',
          surface: '#001d31',
          'on-surface': '#cde5ff',
          outline: '#998f80',
          'surface-variant': '#ede1cf',
          'on-surface-variant': '#d1c5b4',

          'surface-bright': '#FFFFFF',
          'surface-light': '#EEEEEE',
          'discord-primary': '#5865F2',
          info: '#2196F3',
          success: '#4CAF50',
          warning: '#FB8C00',
          'card-on-surface': '1e3048',
          ...rankColors
        }
      }
    }
  },
})
