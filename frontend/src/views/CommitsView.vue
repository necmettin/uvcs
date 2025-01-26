<template>
  <div class="commits-view">
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
          <h1>Commits - {{ name }}</h1>
          <router-link :to="`/repository/${name}`" class="btn btn-secondary">
            <span class="mdi mdi-arrow-left"></span>
            Back to Repository
          </router-link>
        </div>

        <!-- Branch Selection -->
        <div class="branch-selector mt-4">
          <select v-model="currentBranch" class="branch-select">
            <option v-for="branch in branches" :key="branch.id" :value="branch.name">
              {{ branch.name }}
            </option>
          </select>
        </div>
      </div>

      <!-- Commits List -->
      <div class="commits-list">
        <div v-for="commit in commits" :key="commit.id" class="commit-item card mb-3">
          <div class="commit-header">
            <div class="commit-title">
              <h3>{{ commit.message }}</h3>
              <div class="commit-meta">
                <span class="commit-author">
                  <span class="mdi mdi-account"></span>
                  {{ commit.author }}
                </span>
                <span class="commit-date">
                  <span class="mdi mdi-clock-outline"></span>
                  {{ formatDate(commit.datetime) }}
                </span>
                <span class="commit-hash">
                  <span class="mdi mdi-source-commit"></span>
                  {{ commit.hash.substring(0, 7) }}
                </span>
              </div>
            </div>
            <button class="btn btn-secondary btn-sm" @click="viewCommitDetails(commit)">
              View Changes
            </button>
          </div>

          <div class="commit-tags" v-if="commit.tags && commit.tags.length > 0">
            <span v-for="tag in commit.tags" :key="tag" class="tag">
              {{ tag }}
            </span>
          </div>
        </div>

        <div v-if="commits.length === 0" class="text-center p-4">
          <p>No commits found in this repository.</p>
        </div>
      </div>

      <!-- Commit Details Modal -->
      <div v-if="showDetailsModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <div class="commit-title">
              <h2>{{ selectedCommit?.message }}</h2>
              <div class="commit-meta">
                <span class="commit-author">
                  <span class="mdi mdi-account"></span>
                  {{ selectedCommit?.author }}
                </span>
                <span class="commit-date">
                  <span class="mdi mdi-clock-outline"></span>
                  {{ selectedCommit ? formatDate(selectedCommit.datetime) : '' }}
                </span>
                <span class="commit-hash">
                  <span class="mdi mdi-source-commit"></span>
                  {{ selectedCommit?.hash.substring(0, 7) }}
                </span>
              </div>
            </div>
            <button class="modal-close" @click="showDetailsModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <div class="modal-body">
            <div class="changes-list">
              <div v-for="change in selectedCommit?.changes" :key="change.id" class="change-item">
                <div class="change-header">
                  <span class="change-type" :class="getChangeTypeClass(change.change_type)">
                    {{ getChangeTypeLabel(change.change_type) }}
                  </span>
                  <span class="change-path">{{ change.file_path }}</span>
                </div>

                <pre v-if="!change.is_binary && change.is_diff" class="change-content">{{ change.content_change }}</pre>
                <div v-else-if="change.is_binary" class="binary-notice">
                  Binary file changes cannot be displayed
                </div>
              </div>
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
import type { Commit } from '../types';

const route = useRoute();
const toast = useToast();
const repositoryStore = useRepositoryStore();

const name = computed(() => route.params.name as string);
const loading = ref(false);
const error = ref('');

// Repository state from store
const branches = computed(() => repositoryStore.branches);
const commits = computed(() => repositoryStore.commits);
const currentBranch = ref('main');

// Commit details state
const showDetailsModal = ref(false);
const selectedCommit = ref<Commit | null>(null);

const loadRepository = async () => {
  loading.value = true;
  error.value = '';

  try {
    await repositoryStore.loadRepository(name.value);
    if (branches.value.length > 0) {
      currentBranch.value = branches.value[0].name;
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load repository';
    toast.error(error.value);
  } finally {
    loading.value = false;
  }
};

const viewCommitDetails = (commit: Commit) => {
  selectedCommit.value = commit;
  showDetailsModal.value = true;
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

const getChangeTypeClass = (type: 'A' | 'M' | 'D') => {
  switch (type) {
    case 'A': return 'change-added';
    case 'M': return 'change-modified';
    case 'D': return 'change-deleted';
  }
};

const getChangeTypeLabel = (type: 'A' | 'M' | 'D') => {
  switch (type) {
    case 'A': return 'Added';
    case 'M': return 'Modified';
    case 'D': return 'Deleted';
  }
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
.commits-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.header {
  padding: 2rem;
}

.branch-selector {
  margin-top: 1rem;
}

.branch-select {
  min-width: 200px;
}

.commit-item {
  padding: 1.5rem;
}

.commit-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.commit-title {
  flex: 1;
}

.commit-title h3 {
  margin-bottom: 0.5rem;
  font-size: 1.1rem;
}

.commit-meta {
  display: flex;
  gap: 1rem;
  color: #666;
  font-size: 0.9rem;
}

.commit-meta span {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.commit-tags {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
}

.tag {
  background: var(--light-gray);
  color: var(--secondary-color);
  padding: 0.25rem 0.75rem;
  border-radius: var(--border-radius);
  font-size: 0.875rem;
}

.changes-list {
  margin-top: 1rem;
}

.change-item {
  margin-bottom: 1.5rem;
}

.change-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.5rem;
}

.change-type {
  padding: 0.25rem 0.5rem;
  border-radius: var(--border-radius);
  font-size: 0.875rem;
  font-weight: 500;
}

.change-added {
  background: #d4edda;
  color: #155724;
}

.change-modified {
  background: #fff3cd;
  color: #856404;
}

.change-deleted {
  background: #f8d7da;
  color: #721c24;
}

.change-path {
  font-family: monospace;
}

.change-content {
  background: var(--light-gray);
  padding: 1rem;
  border-radius: var(--border-radius);
  overflow-x: auto;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 0.875rem;
}

.binary-notice {
  padding: 1rem;
  background: var(--light-gray);
  border-radius: var(--border-radius);
  color: #666;
  font-style: italic;
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
  max-width: 900px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #666;
}
</style> 