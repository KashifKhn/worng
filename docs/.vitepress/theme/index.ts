import DefaultTheme from 'vitepress/theme'
import './style.css'
import WrongPlayground from './components/WrongPlayground.vue'

export default {
  ...DefaultTheme,
  enhanceApp({ app }) {
    app.component('WrongPlayground', WrongPlayground)
  },
}
