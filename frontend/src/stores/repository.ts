import { defineStore } from 'pinia';
import { api } from '../services/api';
import type { Branch, Commit, RepositoryAccess } from '../types';

interface RepositoryState {
    currentRepository: string | null;
    branches: Branch[];
    commits: Commit[];
    content: { [key: string]: string };
    access: RepositoryAccess[];
    loading: boolean;
    error: string | null;
}

export const useRepositoryStore = defineStore('repository', {
    state: (): RepositoryState => ({
        currentRepository: null,
        branches: [],
        commits: [],
        content: {},
        access: [],
        loading: false,
        error: null,
    }),

    getters: {
        hasWriteAccess: (state) => {
            // TODO: Compare with current user ID
            return state.access.some(a => a.access_level === 'write');
        },
    },

    actions: {
        async loadRepository(name: string) {
            this.loading = true;
            this.error = null;
            try {
                const response = await api.getRepository(name);
                this.currentRepository = name;
                this.branches = response.branches;
                this.commits = response.commits;
                this.content = response.content;
                this.access = response.access;
            } catch (error) {
                console.error('Failed to load repository:', error);
                this.error = 'Failed to load repository';
                throw error;
            } finally {
                this.loading = false;
            }
        },

        async createCommit(message: string, files: File[], tags?: string[]) {
            if (!this.currentRepository) throw new Error('No repository selected');
            
            this.loading = true;
            this.error = null;
            try {
                const result = await api.createCommit(this.currentRepository, message, files, tags);
                await this.loadRepository(this.currentRepository);
                return result;
            } catch (error) {
                console.error('Failed to create commit:', error);
                this.error = 'Failed to create commit';
                throw error;
            } finally {
                this.loading = false;
            }
        },

        async createBranch(branchName: string) {
            if (!this.currentRepository) throw new Error('No repository selected');
            
            this.loading = true;
            this.error = null;
            try {
                await api.createBranch(this.currentRepository, branchName);
                await this.loadRepository(this.currentRepository);
            } catch (error) {
                console.error('Failed to create branch:', error);
                this.error = 'Failed to create branch';
                throw error;
            } finally {
                this.loading = false;
            }
        },

        async deleteBranch(branchName: string) {
            if (!this.currentRepository) throw new Error('No repository selected');
            
            this.loading = true;
            this.error = null;
            try {
                await api.deleteBranch(this.currentRepository, branchName);
                await this.loadRepository(this.currentRepository);
            } catch (error) {
                console.error('Failed to delete branch:', error);
                this.error = 'Failed to delete branch';
                throw error;
            } finally {
                this.loading = false;
            }
        },

        async grantAccess(username: string, accessLevel: 'read' | 'write') {
            if (!this.currentRepository) throw new Error('No repository selected');
            
            this.loading = true;
            this.error = null;
            try {
                await api.grantAccess(this.currentRepository, username, accessLevel);
                await this.loadRepository(this.currentRepository);
            } catch (error) {
                console.error('Failed to grant access:', error);
                this.error = 'Failed to grant access';
                throw error;
            } finally {
                this.loading = false;
            }
        },

        async revokeAccess(username: string) {
            if (!this.currentRepository) throw new Error('No repository selected');
            
            this.loading = true;
            this.error = null;
            try {
                await api.revokeAccess(this.currentRepository, username);
                await this.loadRepository(this.currentRepository);
            } catch (error) {
                console.error('Failed to revoke access:', error);
                this.error = 'Failed to revoke access';
                throw error;
            } finally {
                this.loading = false;
            }
        },

        clearRepository() {
            this.currentRepository = null;
            this.branches = [];
            this.commits = [];
            this.content = {};
            this.access = [];
            this.error = null;
        },
    },
}); 