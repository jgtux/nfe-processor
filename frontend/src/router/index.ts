import { createRouter, createWebHistory } from 'vue-router'
import UploadPage from '@/pages/UploadPage.vue'
import DashboardPage from '@/pages/DashboardPage.vue'
import UnidentifiedPage from '@/pages/UnidentifiedPage.vue'
import QuarantinePage from '@/pages/QuarantinePage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',             name: 'upload',       component: UploadPage },
    { path: '/dashboard',    name: 'dashboard',    component: DashboardPage },
    { path: '/unidentified', name: 'unidentified', component: UnidentifiedPage },
    { path: '/quarantine',   name: 'quarantine',   component: QuarantinePage }
  ]
})

export default router
