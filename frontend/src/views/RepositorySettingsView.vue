<template>
  <div class="settings-view">
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
          <h1>Settings - {{ name }}</h1>
          <router-link :to="`/repository/${name}`" class="btn btn-secondary">
            <span class="mdi mdi-arrow-left"></span>
            Back to Repository
          </router-link>
        </div>
      </div>

      <!-- Access Control -->
      <div class="card mb-4">
        <div class="section-header d-flex justify-between align-center mb-4">
          <h2>Access Control</h2>
          <button v-if="hasWriteAccess" class="btn btn-primary" @click="showGrantModal = true">
            <span class="mdi mdi-account-plus"></span>
            Grant Access
          </button>
        </div>

        <div class="access-list">
          <div v-for="access in repositoryAccess" :key="`${access.repository_id}-${access.user_id}`" class="access-item">
            <div class="access-info">
              <span class="access-username">{{ access.user_id }}</span>
              <span class="access-level" :class="access.access_level">
                {{ access.access_level }}
              </span>
              <span class="access-meta">
                Granted by {{ access.granted_by }} on {{ formatDate(access.granted_at) }}
              </span>
            </div>
            <div class="access-actions" v-if="hasWriteAccess && access.user_id !== currentUserId">
              <button class="btn btn-danger btn-sm" @click="confirmRevoke(access)">
                <span class="mdi mdi-account-remove"></span>
                Revoke
              </button>
            </div>
          </div>

          <div v-if="repositoryAccess.length === 0" class="text-center p-4">
            <p>No additional users have access to this repository.</p>
          </div>
        </div>
      </div>

      <!-- Danger Zone -->
      <div class="card danger-zone">
        <h2 class="text-danger mb-4">Danger Zone</h2>

        <div class="danger-actions">
          <div class="danger-action">
            <div class="danger-info">
              <h3>Delete Repository</h3>
              <p>Once you delete a repository, there is no going back. Please be certain.</p>
            </div>
            <button class="btn btn-danger" @click="confirmDelete">
              Delete Repository
            </button>
          </div>
        </div>
      </div>

      <!-- Grant Access Modal -->
      <div v-if="showGrantModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <h2>Grant Repository Access</h2>
            <button class="modal-close" @click="showGrantModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <form @submit.prevent="handleGrant" class="modal-body">
            <div class="form-group">
              <label for="username">Username</label>
              <input
                type="text"
                id="username"
                v-model="grantAccess.username"
                required
                :disabled="granting"
              >
            </div>

            <div class="form-group">
              <label for="accessLevel">Access Level</label>
              <select id="accessLevel" v-model="grantAccess.level" required :disabled="granting">
                <option value="read">Read</option>
                <option value="write">Write</option>
              </select>
            </div>

            <div v-if="grantError" class="alert alert-error">
              {{ grantError }}
            </div>

            <div class="modal-actions">
              <button type="button" class="btn btn-secondary" @click="showGrantModal = false" :disabled="granting">
                Cancel
              </button>
              <button type="submit" class="btn btn-primary" :disabled="granting">
                <span v-if="granting" class="loading-spinner"></span>
                <span v-else>Grant Access</span>
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Revoke Access Modal -->
      <div v-if="showRevokeModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <h2>Revoke Access</h2>
            <button class="modal-close" @click="showRevokeModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <div class="modal-body">
            <p>Are you sure you want to revoke access for user "{{ selectedAccess?.user_id }}"?</p>

            <div v-if="revokeError" class="alert alert-error mt-3">
              {{ revokeError }}
            </div>

            <div class="modal-actions mt-4">
              <button type="button" class="btn btn-secondary" @click="showRevokeModal = false" :disabled="revoking">
                Cancel
              </button>
              <button type="button" class="btn btn-danger" @click="handleRevoke" :disabled="revoking">
                <span v-if="revoking" class="loading-spinner"></span>
                <span v-else>Revoke Access</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Delete Repository Modal -->
      <div v-if="showDeleteModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <h2>Delete Repository</h2>
            <button class="modal-close" @click="showDeleteModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <div class="modal-body">
            <p>Are you sure you want to delete this repository? This action cannot be undone.</p>
            <p class="text-danger">All repository data, including commits, branches, and files will be permanently deleted.</p>

            <div class="form-group mt-4">
              <label for="confirmName">Please type the repository name to confirm:</label>
              <input
                type="text"
                id="confirmName"
                v-model="deleteConfirmName"
                :placeholder="name"
                :disabled="deleting"
              >
            </div>

            <div v-if="deleteError" class="alert alert-error mt-3">
              {{ deleteError }}
            </div>

            <div class="modal-actions mt-4">
              <button type="button" class="btn btn-secondary" @click="showDeleteModal = false" :disabled="deleting">
                Cancel
              </button>
              <button
                type="button"
                class="btn btn-danger"
                @click="handleDelete"
                :disabled="deleting || deleteConfirmName !== name"
              >
                <span v-if="deleting" class="loading-spinner"></span>
                <span v-else>Delete Repository</span>
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
import { useRoute, useRouter } from 'vue-router';
import { useToast } from 'vue-toastification';
import { useRepositoryStore } from '../stores/repository';
import type { RepositoryAccess } from '../types';

const route = useRoute();
const router = useRouter();
const toast = useToast();
const repositoryStore = useRepositoryStore();

const name = computed(() => route.params.name as string);
const loading = ref(false);
const error = ref('');

// Repository state from store
const repositoryAccess = computed(() => repositoryStore.access);
const hasWriteAccess = computed(() => repositoryStore.hasWriteAccess);
const currentUserId = ref(1); // TODO: Get from auth store

// Grant access state
const showGrantModal = ref(false);
const granting = ref(false);
const grantError = ref('');
const grantAccess = ref({
  username: '',
  level: 'read' as 'read' | 'write',
});

// Revoke access state
const showRevokeModal = ref(false);
const revoking = ref(false);
const revokeError = ref('');
const selectedAccess = ref<RepositoryAccess | null>(null);

// Delete repository state
const showDeleteModal = ref(false);
const deleting = ref(false);
const deleteError = ref('');
const deleteConfirmName = ref('');

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

const handleGrant = async () => {
  granting.value = true;
  grantError.value = '';

  try {
    await repositoryStore.grantAccess(grantAccess.value.username, grantAccess.value.level);
    showGrantModal.value = false;
    toast.success('Access granted successfully');
    
    // Reset form
    grantAccess.value.username = '';
    grantAccess.value.level = 'read';
  } catch (err) {
    grantError.value = err instanceof Error ? err.message : 'Failed to grant access';
    toast.error(grantError.value);
  } finally {
    granting.value = false;
  }
};

const confirmRevoke = (access: RepositoryAccess) => {
  selectedAccess.value = access;
  showRevokeModal.value = true;
};

const handleRevoke = async () => {
  if (!selectedAccess.value) return;
  
  revoking.value = true;
  revokeError.value = '';

  try {
    await repositoryStore.revokeAccess(selectedAccess.value.user_id.toString());
    showRevokeModal.value = false;
    toast.success('Access revoked successfully');
  } catch (err) {
    revokeError.value = err instanceof Error ? err.message : 'Failed to revoke access';
    toast.error(revokeError.value);
  } finally {
    revoking.value = false;
  }
};

const confirmDelete = () => {
  showDeleteModal.value = true;
};

const handleDelete = async () => {
  if (deleteConfirmName.value !== name.value) return;
  
  deleting.value = true;
  deleteError.value = '';

  try {
    // TODO: Implement repository deletion
    showDeleteModal.value = false;
    toast.success('Repository deleted successfully');
    await router.push('/');
  } catch (err) {
    deleteError.value = err instanceof Error ? err.message : 'Failed to delete repository';
    toast.error(deleteError.value);
  } finally {
    deleting.value = false;
  }
};

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
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
.settings-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.header {
  padding: 2rem;
}

.section-header {
  margin-bottom: 2rem;
}

.access-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 0;
  border-bottom: 1px solid var(--border-color);
}

.access-item:last-child {
  border-bottom: none;
}

.access-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.access-username {
  font-weight: 500;
}

.access-level {
  padding: 0.25rem 0.5rem;
  border-radius: var(--border-radius);
  font-size: 0.875rem;
  font-weight: 500;
}

.access-level.read {
  background: var(--light-gray);
  color: var(--secondary-color);
}

.access-level.write {
  background: #d4edda;
  color: #155724;
}

.access-meta {
  color: #666;
  font-size: 0.875rem;
}

.access-actions {
  display: flex;
  gap: 0.5rem;
}

/* Danger Zone */
.danger-zone {
  border: 1px solid var(--error-color);
}

.danger-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 0;
  border-bottom: 1px solid var(--border-color);
}

.danger-action:last-child {
  border-bottom: none;
}

.danger-info h3 {
  color: var(--error-color);
  margin-bottom: 0.5rem;
}

.danger-info p {
  color: #666;
  font-size: 0.875rem;
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