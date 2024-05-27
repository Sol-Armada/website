/**
 * plugins/index.js
 *
 * Automatically included in `./src/main.js`
 */

// Plugins
import vuetify from './vuetify'
import pinia from '@/stores'
import router from '@/router'
import vueMatomo from 'vue-matomo'



export function registerPlugins(app) {
  let siteId = 1

  // change site id based on environment
  if (import.meta.env.MODE === 'beta') {
    siteId = 2
  } else if (import.meta.env.MODE === 'development') {
    siteId = 3
  }

  app
    .use(vuetify)
    .use(router)
    .use(pinia)
    .use(vueMatomo, {
      host: 'https://matomo.solarmada.space',
      siteId: siteId,
    })

  window._paq.push(['trackPageView'])
}
