<template>
  <div class="register-view">
    <div class="card">
      <h1 class="text-center">Register</h1>
      
      <form @submit.prevent="handleSubmit" class="register-form">
        <div class="form-group">
          <label for="username">Username</label>
          <input
            type="text"
            id="username"
            v-model="form.username"
            required
            :disabled="loading"
          >
        </div>

        <div class="form-group">
          <label for="email">Email</label>
          <input
            type="email"
            id="email"
            v-model="form.email"
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

        <div class="grid grid-2">
          <div class="form-group">
            <label for="firstname">First Name</label>
            <input
              type="text"
              id="firstname"
              v-model="form.firstname"
              :disabled="loading"
            >
          </div>

          <div class="form-group">
            <label for="lastname">Last Name</label>
            <input
              type="text"
              id="lastname"
              v-model="form.lastname"
              :disabled="loading"
            >
          </div>
        </div>

        <div v-if="error" class="alert alert-error">
          {{ error }}
        </div>

        <button type="submit" class="btn btn-primary w-100" :disabled="loading">
          <span v-if="loading" class="loading-spinner"></span>
          <span v-else>Register</span>
        </button>

        <p class="text-center mt-3">
          Already have an account?
          <router-link to="/login">Login</router-link>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useToast } from 'vue-toastification';
import { useAuthStore } from '../stores/auth';

const router = useRouter();
const toast = useToast();
const authStore = useAuthStore();

const loading = ref(false);
const error = ref('');
const form = reactive({
  username: '',
  email: '',
  password: '',
  firstname: '',
  lastname: '',
});

const handleSubmit = async () => {
  loading.value = true;
  error.value = '';

  try {
    await authStore.register(
      form.username,
      form.email,
      form.password,
      form.firstname,
      form.lastname
    );
    toast.success('Registration successful! Please login.');
    await router.push('/login');
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to register';
    toast.error(error.value);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.register-view {
  max-width: 500px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.register-form {
  margin-top: 2rem;
}
</style> 