<template>
  <div class="branches-view">
    <div v-if="loading" class="text-center p-4">
      <span class="loading-spinner"></span>
    </div>

    <div v-else-if="error" class="alert alert-error">
      {{ error }}
    </div>

    <template v-else>
      <!-- Header -->
      <div class="header card mb-4">
        <div class="d-flex justify-between align-center">
          <h1>Branches - {{ name }}</h1>
          <div class="header-actions">
            <router-link :to="`/repository/${name}`" class="btn btn-secondary">
              <span class="mdi mdi-arrow-left"></span>
              Back to Repository
            </router-link>
            <button v-if="hasWriteAccess" class="btn btn-primary" @click="showCreateModal = true">
              <span class="mdi mdi-source-branch-plus"></span>
              New Branch
            </button>
          </div>
        </div>
      </div>

      <!-- Branches List -->
      <div class="branches-list">
        <div v-for="branch in branches" :key="branch.id" class="branch-item card mb-3">
          <div class="branch-header">
            <div class="branch-info">
              <h3>
                <span class="mdi mdi-source-branch"></span>
                {{ branch.name }}
              </h3>
              <div class="branch-meta">
                <span class="branch-commit">
                  <span class="mdi mdi-source-commit"></span>
                  Latest commit: {{ getCommitHash(branch.head_commit_id) }}
                </span>
              </div>
            </div>
            <div class="branch-actions" v-if="hasWriteAccess && branch.name !== 'main'">
              <button class="btn btn-danger btn-sm" @click="confirmDelete(branch)">
                <span class="mdi mdi-delete"></span>
                Delete
              </button>
            </div>
          </div>
        </div>

        <div v-if="branches.length === 0" class="text-center p-4">
          <p>No branches found in this repository.</p>
        </div>
      </div>

      <!-- Create Branch Modal -->
      <div v-if="showCreateModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <h2>Create New Branch</h2>
            <button class="modal-close" @click="showCreateModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <form @submit.prevent="handleCreate" class="modal-body">
            <div class="form-group">
              <label for="branchName">Branch Name</label>
              <input
                type="text"
                id="branchName"
                v-model="newBranch.name"
                required
                :disabled="creating"
                pattern="[a-zA-Z0-9_-]+"
                title="Only letters, numbers, underscores, and hyphens are allowed"
              >
            </div>

            <div class="form-group">
              <label for="sourceBranch">Source Branch</label>
              <select id="sourceBranch" v-model="newBranch.source" required :disabled="creating">
                <option v-for="branch in branches" :key="branch.id" :value="branch.name">
                  {{ branch.name }}
                </option>
              </select>
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
                <span v-else>Create Branch</span>
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Delete Confirmation Modal -->
      <div v-if="showDeleteModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <h2>Delete Branch</h2>
            <button class="modal-close" @click="showDeleteModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <div class="modal-body">
            <p>Are you sure you want to delete the branch "{{ selectedBranch?.name }}"?</p>
            <p class="text-danger">This action cannot be undone.</p>

            <div v-if="deleteError" class="alert alert-error mt-3">
              {{ deleteError }}
            </div>

            <div class="modal-actions mt-4">
              <button type="button" class="btn btn-secondary" @click="showDeleteModal = false" :disabled="deleting">
                Cancel
              </button>
              <button type="button" class="btn btn-danger" @click="handleDelete" :disabled="deleting">
                <span v-if="deleting" class="loading-spinner"></span>
                <span v-else>Delete Branch</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useToast } from 'vue-toastification';
import { useRepositoryStore } from '../stores/repository';
import type { Branch } from '../types';

const route = useRoute();
const toast = useToast();
const repositoryStore = useRepositoryStore();

const name = computed(() => route.params.name as string);
const loading = ref(false);
const error = ref('');

// Repository state from store
const branches = computed(() => repositoryStore.branches);
const commits = computed(() => repositoryStore.commits);
const hasWriteAccess = computed(() => repositoryStore.hasWriteAccess);

// Create branch state
const showCreateModal = ref(false);
const creating = ref(false);
const createError = ref('');
const newBranch = ref({
  name: '',
  source: 'main',
});

// Delete branch state
const showDeleteModal = ref(false);
const deleting = ref(false);
const deleteError = ref('');
const selectedBranch = ref<Branch | null>(null);

const loadRepository = async () => {
  loading.value = true;
  error.value = '';

  try {
    await repositoryStore.loadRepository(name.value);
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load repository';
    toast.error(error.value);
  } finally {
    loading.value = false;
  }
};

const handleCreate = async () => {
  creating.value = true;
  createError.value = '';

  try {
    await repositoryStore.createBranch(newBranch.value.name);
    showCreateModal.value = false;
    toast.success('Branch created successfully');
    
    // Reset form
    newBranch.value.name = '';
    newBranch.value.source = 'main';
  } catch (err) {
    createError.value = err instanceof Error ? err.message : 'Failed to create branch';
    toast.error(createError.value);
  } finally {
    creating.value = false;
  }
};

const confirmDelete = (branch: Branch) => {
  selectedBranch.value = branch;
  showDeleteModal.value = true;
};

const handleDelete = async () => {
  if (!selectedBranch.value) return;
  
  deleting.value = true;
  deleteError.value = '';

  try {
    await repositoryStore.deleteBranch(selectedBranch.value.name);
    showDeleteModal.value = false;
    toast.success('Branch deleted successfully');
  } catch (err) {
    deleteError.value = err instanceof Error ? err.message : 'Failed to delete branch';
    toast.error(deleteError.value);
  } finally {
    deleting.value = false;
  }
};

const getCommitHash = (commitId: number) => {
  const commit = commits.value.find(c => c.id === commitId);
  return commit ? commit.hash.substring(0, 7) : 'Unknown';
};

// Watch for route changes
watch(() => route.params.name, (newName) => {
  if (newName) {
    loadRepository();
  }
});

onMounted(() => {
  loadRepository();
});
</script>

<style scoped>
.branches-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.header {
  padding: 2rem;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.branch-item {
  padding: 1.5rem;
}

.branch-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.branch-info {
  flex: 1;
}

.branch-info h3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  font-size: 1.1rem;
}

.branch-meta {
  color: #666;
  font-size: 0.9rem;
}

.branch-meta span {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.branch-actions {
  display: flex;
  gap: 0.5rem;
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
  margin-top: 1rem;
}

.text-danger {
  color: var(--error-color);
}
</style> 