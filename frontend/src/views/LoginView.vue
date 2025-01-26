<template>
  <div class="login-view">
    <div class="card">
      <h1 class="text-center">Login</h1>
      
      <form @submit.prevent="handleSubmit" class="login-form">
        <div class="form-group">
          <label for="identifier">Username or Email</label>
          <input
            type="text"
            id="identifier"
            v-model="form.identifier"
            required
            :disabled="loading"
          >
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input
            type="password"
            id="password"
            v-model="form.password"
            required
            :disabled="loading"
          >
        </div>

        <div v-if="error" class="alert alert-error">
          {{ error }}
        </div>

        <button type="submit" class="btn btn-primary w-100" :disabled="loading">
          <span v-if="loading" class="loading-spinner"></span>
          <span v-else>Login</span>
        </button>

        <p class="text-center mt-3">
          Don't have an account?
          <router-link to="/register">Register</router-link>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useToast } from 'vue-toastification';
import { useAuthStore } from '../stores/auth';

const router = useRouter();
const route = useRoute();
const toast = useToast();
const authStore = useAuthStore();

const loading = ref(false);
const error = ref('');
const form = reactive({
  identifier: '',
  password: '',
});

const handleSubmit = async () => {
  loading.value = true;
  error.value = '';

  try {
    await authStore.login(form.identifier, form.password);
    toast.success('Logged in successfully');
    
    // Redirect to the requested page or home
    const redirect = route.query.redirect as string;
    await router.push(redirect || '/');
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to login';
    toast.error(error.value);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.login-view {
  max-width: 400px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.login-form {
  margin-top: 2rem;
}
</style> 