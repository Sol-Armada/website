import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { aliases, mdi } from 'vuetify/lib/iconsets/mdi'
import '@mdi/js'
import 'vuetify/styles'

const darkTheme = {
    dark: true,
    colors: {
        background: '#0a0e27',
        surface: '#141829',
        primary: '#00d9ff',
        secondary: '#ff006e',
        accent: '#ff9f1c',
        success: '#06d6a0',
        warning: '#fca311',
        error: '#e63946',
        info: '#457b9d',
    },
}

export default createVuetify({
    components,
    directives,
    icons: {
        defaultSet: 'mdi',
        aliases,
        sets: {
            mdi,
        },
    },
    theme: {
        defaultTheme: 'darkTheme',
        themes: {
            darkTheme,
        },
    },
})
