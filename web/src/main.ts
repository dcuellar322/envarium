import { VueQueryPlugin } from '@tanstack/vue-query'
import { createApp } from 'vue'

import App from './app/App.vue'
import { bootstrapBrowserSession } from './domains/session/bootstrap'
import { queryClient } from './queryClient'
import './styles/base.css'

async function start() {
  await bootstrapBrowserSession().catch(() => undefined)
  const { router } = await import('./router')
  const app = createApp(App)
  app.use(VueQueryPlugin, { queryClient })
  app.use(router)
  app.mount('#app')
}

void start()
