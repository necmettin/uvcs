import axios from 'axios';
import type { AxiosInstance } from 'axios';
import type { AuthResponse, RepositoryResponse } from '../types';

class ApiService {
    private api: AxiosInstance;

    constructor() {
        this.api = axios.create({
            baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        // Add request interceptor to include auth keys
        this.api.interceptors.request.use((config) => {
            const skey1 = localStorage.getItem('skey1');
            const skey2 = localStorage.getItem('skey2');
            
            if (skey1 && skey2) {
                config.headers.skey1 = skey1;
                config.headers.skey2 = skey2;
            }
            
            return config;
        });
    }

    private addAuthToFormData(formData: FormData): FormData {
        const skey1 = localStorage.getItem('skey1');
        const skey2 = localStorage.getItem('skey2');
        if (skey1 && skey2) {
            formData.append('skey1', skey1);
            formData.append('skey2', skey2);
        }
        return formData;
    }

    // Auth endpoints
    async register(username: string, email: string, password: string, firstname?: string, lastname?: string): Promise<void> {
        const formData = new FormData();
        formData.append('username', username);
        formData.append('email', email);
        formData.append('password', password);
        if (firstname) formData.append('firstname', firstname);
        if (lastname) formData.append('lastname', lastname);

        await this.api.postForm('/api/register', formData);
    }

    async login(identifier: string, password: string): Promise<AuthResponse> {
        const formData = new FormData();
        formData.append('identifier', identifier);
        formData.append('password', password);

        const response = await this.api.postForm<AuthResponse>('/api/login', formData);
        localStorage.setItem('skey1', response.data.skey1);
        localStorage.setItem('skey2', response.data.skey2);
        return response.data;
    }

    // Repository endpoints
    async getRepository(name: string): Promise<RepositoryResponse> {
        const formData = new FormData();
        formData.append('name', name);
        const response = await this.api.postForm<RepositoryResponse>('/api/repository', this.addAuthToFormData(formData));
        return response.data;
    }

    async createRepository(name: string, description?: string): Promise<RepositoryResponse> {
        const formData = new FormData();
        formData.append('name', name);
        if (description) formData.append('description', description);
        
        const response = await this.api.postForm<RepositoryResponse>('/api/repository', this.addAuthToFormData(formData));
        return response.data;
    }

    async createCommit(name: string, message: string, files: File[], tags?: string[]): Promise<{ commit_id: string; commit_hash: string }> {
        const formData = new FormData();
        formData.append('name', name);
        formData.append('message', message);
        files.forEach(file => formData.append('files[]', file));
        if (tags) formData.append('tags', tags.join(','));

        const response = await this.api.postForm('/api/repository/commit', this.addAuthToFormData(formData));
        return response.data;
    }

    // Branch operations
    async listBranches(repoName: string): Promise<void> {
        const formData = new FormData();
        formData.append('name', repoName);
        await this.api.postForm(`/api/repository/${encodeURIComponent(repoName)}/branches`, this.addAuthToFormData(formData));
    }

    async createBranch(repoName: string, branchName: string): Promise<void> {
        const formData = new FormData();
        formData.append('name', branchName);
        await this.api.postForm(`/api/repository/${encodeURIComponent(repoName)}/branches`, this.addAuthToFormData(formData));
    }

    async deleteBranch(repoName: string, branchName: string): Promise<void> {
        const formData = new FormData();
        await this.api.postForm(`/api/repository/${encodeURIComponent(repoName)}/branches/${encodeURIComponent(branchName)}/delete`, this.addAuthToFormData(formData));
    }

    // Access control
    async grantAccess(repoName: string, username: string, accessLevel: 'read' | 'write'): Promise<void> {
        const formData = new FormData();
        formData.append('username', username);
        formData.append('access_level', accessLevel);
        await this.api.postForm(`/api/repository/${encodeURIComponent(repoName)}/access`, this.addAuthToFormData(formData));
    }

    async revokeAccess(repoName: string, username: string): Promise<void> {
        const formData = new FormData();
        formData.append('username', username);
        await this.api.postForm(`/api/repository/${encodeURIComponent(repoName)}/access/revoke`, this.addAuthToFormData(formData));
    }

    async listAccess(repoName: string): Promise<void> {
        const formData = new FormData();
        formData.append('name', repoName);
        await this.api.postForm(`/api/repository/${encodeURIComponent(repoName)}/access/list`, this.addAuthToFormData(formData));
    }

    // Utility methods
    isAuthenticated(): boolean {
        return !!(localStorage.getItem('skey1') && localStorage.getItem('skey2'));
    }

    logout(): void {
        localStorage.removeItem('skey1');
        localStorage.removeItem('skey2');
    }
}

export const api = new ApiService(); 