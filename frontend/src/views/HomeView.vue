<template>
  <div class="home-view">
    <div class="header-actions d-flex justify-between align-center mb-4">
      <h1>My Repositories</h1>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <span class="mdi mdi-plus"></span>
        New Repository
      </button>
    </div>

    <div v-if="loading" class="text-center p-4">
      <span class="loading-spinner"></span>
    </div>

    <div v-else-if="error" class="alert alert-error">
      {{ error }}
    </div>

    <div v-else class="repositories-grid">
      <div v-for="repo in repositories" :key="repo.id" class="card repository-card">
        <div class="repository-header">
          <h3>
            <router-link :to="`/repository/${repo.name}`">
              {{ repo.name }}
            </router-link>
          </h3>
          <span class="repository-date">Created {{ formatDate(repo.created_at) }}</span>
        </div>

        <div class="repository-actions">
          <router-link :to="`/repository/${repo.name}`" class="btn btn-secondary">
            <span class="mdi mdi-source-repository"></span>
            View
          </router-link>
          <router-link :to="`/repository/${repo.name}/settings`" class="btn btn-secondary">
            <span class="mdi mdi-cog"></span>
            Settings
          </router-link>
        </div>
      </div>

      <div v-if="repositories.length === 0" class="text-center p-4">
        <p>No repositories found. Create your first repository to get started!</p>
      </div>
    </div>

    <!-- Create Repository Modal -->
    <div v-if="showCreateModal" class="modal">
      <div class="modal-content card">
        <div class="modal-header">
          <h2>Create New Repository</h2>
          <button class="modal-close" @click="showCreateModal = false">
            <span class="mdi mdi-close"></span>
          </button>
        </div>

        <form @submit.prevent="handleCreate" class="modal-body">
          <div class="form-group">
            <label for="repoName">Repository Name</label>
            <input
              type="text"
              id="repoName"
              v-model="newRepo.name"
              required
              :disabled="creating"
              pattern="[a-zA-Z0-9_-]+"
              title="Only letters, numbers, underscores, and hyphens are allowed"
            >
          </div>

          <div class="form-group">
            <label for="repoDescription">Description (optional)</label>
            <textarea
              id="repoDescription"
              v-model="newRepo.description"
              :disabled="creating"
              rows="3"
            ></textarea>
          </div>

          <div v-if="createError" class="alert alert-error">
            {{ createError }}
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false" :disabled="creating">
              Cancel
            </button>
            <button type="submit" class="btn btn-primary" :disabled="creating">
              <span v-if="creating" class="loading-spinner"></span>
              <span v-else>Create Repository</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useToast } from 'vue-toastification';
import type { Repository } from '../types';
import { api } from '../services/api';

const router = useRouter();
const toast = useToast();

const loading = ref(false);
const error = ref('');
const repositories = ref<Repository[]>([]);

const showCreateModal = ref(false);
const creating = ref(false);
const createError = ref('');
const newRepo = ref({
  name: '',
  description: '',
});

const loadRepositories = async () => {
  loading.value = true;
  error.value = '';

  try {
    // TODO: Implement API call to get repositories
    repositories.value = [];
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load repositories';
    toast.error(error.value);
  } finally {
    loading.value = false;
  }
};

const handleCreate = async () => {
  creating.value = true;
  createError.value = '';

  try {
    const formData = new FormData();
    formData.append('name', newRepo.value.name);
    formData.append('description', newRepo.value.description);

    await api.createRepository(newRepo.value.name, newRepo.value.description);
    toast.success('Repository created successfully');
    showCreateModal.value = false;
    await router.push(`/repository/${newRepo.value.name}`);
  } catch (err) {
    createError.value = err instanceof Error ? err.message : 'Failed to create repository';
    toast.error(createError.value);
  } finally {
    creating.value = false;
  }
};

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

onMounted(() => {
  loadRepositories();
});
</script>

<style scoped>
.home-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.repositories-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.repository-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100%;
}

.repository-header {
  margin-bottom: 1rem;
}

.repository-date {
  font-size: 0.875rem;
  color: #666;
}

.repository-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
}

/* Modal Styles */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1rem;
  z-index: 1000;
}

.modal-content {
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #666;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 2rem;
}
</style> 