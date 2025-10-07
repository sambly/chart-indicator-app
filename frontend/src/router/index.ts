import { createRouter, createWebHistory } from 'vue-router'
import RSIPage from '../views/RSIPage.vue'

const routes = [
  {
    path: '/rsi',
    name: 'RSI',
    component: RSIPage
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router