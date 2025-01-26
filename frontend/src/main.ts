import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Toast, { POSITION } from 'vue-toastification'
import App from './App.vue'
import router from './router'

// Styles
import 'vue-toastification/dist/index.css'
import '@mdi/font/css/materialdesignicons.css'
import './assets/main.css'

const app = createApp(App)

// Configure plugins
const pinia = createPinia()
const toastOptions = {
    position: POSITION.TOP_RIGHT,
    timeout: 5000,
    closeOnClick: true,
    pauseOnFocusLoss: true,
    pauseOnHover: true,
    draggable: true,
    draggablePercent: 0.6,
    showCloseButtonOnHover: false,
    hideProgressBar: true,
    closeButton: 'button',
    icon: true,
    rtl: false
}

// Use plugins
app.use(pinia)
app.use(router)
app.use(Toast, toastOptions)

app.mount('#app')
