import DefaultTheme from 'vitepress/theme'
import { inject } from '@vercel/analytics'
import './style.css'
import WrongPlayground from './components/WrongPlayground.vue'

export default {
  ...DefaultTheme,
  enhanceApp({ app, router }) {
    app.component('WrongPlayground', WrongPlayground)

    // Inject Vercel Analytics script once on app init
    inject()

    // Track page views on every client-side navigation
    router.onAfterRouteChanged = (to: string) => {
      if (typeof window !== 'undefined' && (window as any).va) {
        (window as any).va('pageview', { url: to })
      }
    }
  },
}
