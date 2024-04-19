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

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: {
      fa,
      mdi
    },
  },
  theme: {
    defaultTheme: 'solarmada-light',
    themes: {
      'solarmada-light': {
        dark: false,
        colors: {
          background: '#999999',
          surface: '#969088',
          'surface-bright': '#FFFFFF',
          'surface-light': '#EEEEEE',
          'on-surface-variant': '#EEEEEE',
          primary: '#F8B800',
          'on-primary': '#402D00',
          secondary: '#A48E65',
          'on-secondary': '#3B2F15',
          'surface-variant': '#998F80',
          'discord-primary': '#5865F2',
          error: '#FF5449',
          info: '#2196F3',
          success: '#4CAF50',
          warning: '#FB8C00',
          'discord-primary': '#5865F2',
          tertiary: "#4B6546",
        }
      }
    }
  },
})
