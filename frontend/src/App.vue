<template>
  <div class="app">
    <header class="header">
      <nav class="nav">
        <router-link to="/" class="logo">
          <span class="mdi mdi-source-branch"></span>
          UVCS
        </router-link>
        
        <div class="nav-links" v-if="isAuthenticated">
          <router-link to="/" class="nav-link">
            <span class="mdi mdi-home"></span>
            Home
          </router-link>
          <button @click="logout" class="nav-link">
            <span class="mdi mdi-logout"></span>
            Logout
          </button>
        </div>
        <div class="nav-links" v-else>
          <router-link to="/login" class="nav-link">
            <span class="mdi mdi-login"></span>
            Login
          </router-link>
          <router-link to="/register" class="nav-link">
            <span class="mdi mdi-account-plus"></span>
            Register
          </router-link>
        </div>
      </nav>
    </header>

    <main class="main">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <footer class="footer">
      <p>&copy; 2024 UVCS. All rights reserved.</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { useAuthStore } from './stores/auth';
import { useToast } from 'vue-toastification';

const authStore = useAuthStore();
const toast = useToast();
const { isAuthenticated } = storeToRefs(authStore);

const logout = () => {
  authStore.logout();
  toast.success('Logged out successfully');
};
</script>

<style>
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: var(--primary-color);
  color: white;
  padding: 1rem;
}

.nav {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-size: 1.5rem;
  font-weight: bold;
  text-decoration: none;
  color: white;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.nav-links {
  display: flex;
  gap: 1rem;
}

.nav-link {
  color: white;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.nav-link:hover {
  background: rgba(255, 255, 255, 0.1);
}

button.nav-link {
  background: none;
  border: none;
  font: inherit;
  cursor: pointer;
}

.main {
  flex: 1;
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.footer {
  background: var(--dark-gray);
  color: white;
  text-align: center;
  padding: 1rem;
  margin-top: auto;
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
