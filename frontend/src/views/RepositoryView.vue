<template>
  <div class="repository-view">
    <div v-if="loading" class="text-center p-4">
      <span class="loading-spinner"></span>
    </div>

    <div v-else-if="error" class="alert alert-error">
      {{ error }}
    </div>

    <template v-else>
      <!-- Repository Header -->
      <div class="repository-header card mb-4">
        <div class="d-flex justify-between align-center">
          <h1>{{ name }}</h1>
          <div class="repository-actions">
            <router-link :to="`/repository/${name}/commits`" class="btn btn-secondary">
              <span class="mdi mdi-source-commit"></span>
              Commits
            </router-link>
            <router-link :to="`/repository/${name}/branches`" class="btn btn-secondary">
              <span class="mdi mdi-source-branch"></span>
              Branches
            </router-link>
            <router-link :to="`/repository/${name}/settings`" class="btn btn-secondary">
              <span class="mdi mdi-cog"></span>
              Settings
            </router-link>
          </div>
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

      <!-- File Browser -->
      <div class="file-browser card">
        <div class="file-browser-header d-flex justify-between align-center mb-4">
          <h2>Files</h2>
          <button v-if="hasWriteAccess" class="btn btn-primary" @click="showUploadModal = true">
            <span class="mdi mdi-upload"></span>
            Upload Files
          </button>
        </div>

        <div class="file-list">
          <div v-for="(content, path) in repositoryContent" :key="path" class="file-item">
            <span class="mdi" :class="getFileIcon(String(path))"></span>
            <span class="file-name" @click="viewFile(String(path), content)">{{ path }}</span>
            <button class="btn btn-sm btn-secondary" @click="viewFile(String(path), content)">
              View
            </button>
          </div>

          <div v-if="Object.keys(repositoryContent).length === 0" class="text-center p-4">
            <p>No files found in this repository.</p>
          </div>
        </div>
      </div>

      <!-- File Upload Modal -->
      <div v-if="showUploadModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <h2>Upload Files</h2>
            <button class="modal-close" @click="showUploadModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <form @submit.prevent="handleUpload" class="modal-body">
            <div class="form-group">
              <label for="files">Select Files</label>
              <input
                type="file"
                id="files"
                multiple
                @change="handleFileSelect"
                :disabled="uploading"
              >
            </div>

            <div class="form-group">
              <label for="commitMessage">Commit Message</label>
              <input
                type="text"
                id="commitMessage"
                v-model="commitMessage"
                required
                :disabled="uploading"
                placeholder="Add files via upload"
              >
            </div>

            <div class="form-group">
              <label for="tags">Tags (optional, comma-separated)</label>
              <input
                type="text"
                id="tags"
                v-model="tags"
                :disabled="uploading"
                placeholder="v1.0, release"
              >
            </div>

            <div v-if="uploadError" class="alert alert-error">
              {{ uploadError }}
            </div>

            <div class="selected-files mt-4" v-if="selectedFiles.length > 0">
              <h3 class="mb-2">Selected Files:</h3>
              <ul class="file-list">
                <li v-for="file in selectedFiles" :key="file.name">
                  {{ file.name }} ({{ formatFileSize(file.size) }})
                </li>
              </ul>
            </div>

            <div class="modal-actions">
              <button type="button" class="btn btn-secondary" @click="showUploadModal = false" :disabled="uploading">
                Cancel
              </button>
              <button type="submit" class="btn btn-primary" :disabled="uploading || selectedFiles.length === 0">
                <span v-if="uploading" class="loading-spinner"></span>
                <span v-else>Upload and Commit</span>
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- File View Modal -->
      <div v-if="showFileModal" class="modal">
        <div class="modal-content card">
          <div class="modal-header">
            <h2>{{ currentFile.path }}</h2>
            <button class="modal-close" @click="showFileModal = false">
              <span class="mdi mdi-close"></span>
            </button>
          </div>

          <div class="modal-body">
            <pre class="file-content">{{ currentFile.content }}</pre>
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

const route = useRoute();
const toast = useToast();
const repositoryStore = useRepositoryStore();

const name = computed(() => String(route.params.name));
const loading = ref(false);
const error = ref('');

// File upload state
const showUploadModal = ref(false);
const uploading = ref(false);
const uploadError = ref('');
const selectedFiles = ref<File[]>([]);
const commitMessage = ref('');
const tags = ref('');

// File view state
const showFileModal = ref(false);
const currentFile = ref<{ path: string; content: string }>({ path: '', content: '' });

// Repository state from store
const branches = computed(() => repositoryStore.branches);
const currentBranch = ref('main');
const repositoryContent = computed(() => repositoryStore.content);
const hasWriteAccess = computed(() => repositoryStore.hasWriteAccess);

const loadRepository = async () => {
  loading.value = true;
  error.value = '';

  try {
    await repositoryStore.loadRepository(String(name.value));
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

const handleFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input.files) {
    selectedFiles.value = Array.from(input.files);
  }
};

const handleUpload = async () => {
  uploading.value = true;
  uploadError.value = '';

  try {
    const tagList = tags.value.split(',').map(tag => tag.trim()).filter(Boolean);
    await repositoryStore.createCommit(commitMessage.value, selectedFiles.value, tagList);
    showUploadModal.value = false;
    toast.success('Files uploaded successfully');
    
    // Reset form
    selectedFiles.value = [];
    commitMessage.value = '';
    tags.value = '';
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : 'Failed to upload files';
    toast.error(uploadError.value);
  } finally {
    uploading.value = false;
  }
};

const viewFile = (path: string, content: string) => {
  currentFile.value = { path, content };
  showFileModal.value = true;
};

const getFileIcon = (path: string) => {
  const ext = path.split('.').pop()?.toLowerCase();
  switch (ext) {
    case 'js':
    case 'ts':
      return 'mdi-language-javascript';
    case 'html':
      return 'mdi-language-html5';
    case 'css':
      return 'mdi-language-css3';
    case 'json':
      return 'mdi-code-json';
    case 'md':
      return 'mdi-markdown';
    case 'go':
      return 'mdi-language-go';
    default:
      return 'mdi-file-document-outline';
  }
};

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// Watch for route changes
watch(() => String(route.params.name), (newName) => {
  if (newName) {
    loadRepository();
  }
});

onMounted(() => {
  loadRepository();
});
</script>

<style scoped>
.repository-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.repository-header {
  padding: 2rem;
}

.repository-actions {
  display: flex;
  gap: 0.5rem;
}

.branch-selector {
  margin-top: 1rem;
}

.branch-select {
  min-width: 200px;
}

.file-browser {
  padding: 2rem;
}

.file-list {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
}

.file-item {
  display: flex;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color);
}

.file-item:last-child {
  border-bottom: none;
}

.file-item .mdi {
  margin-right: 0.5rem;
  font-size: 1.25rem;
  color: var(--secondary-color);
}

.file-name {
  flex: 1;
}

.btn-sm {
  padding: 0.4rem 0.8rem;
  font-size: 0.875rem;
}

.file-content {
  background: var(--light-gray);
  padding: 1rem;
  border-radius: var(--border-radius);
  overflow-x: auto;
  white-space: pre-wrap;
  font-family: monospace;
}

.selected-files {
  background: var(--light-gray);
  padding: 1rem;
  border-radius: var(--border-radius);
}

.selected-files ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.selected-files li {
  padding: 0.25rem 0;
}
</style> 