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

    // Auth endpoints
    async register(username: string, email: string, password: string, firstname?: string, lastname?: string): Promise<void> {
        const formData = new FormData();
        formData.append('username', username);
        formData.append('email', email);
        formData.append('password', password);
        if (firstname) formData.append('firstname', firstname);
        if (lastname) formData.append('lastname', lastname);

        await this.api.post('/api/register', formData);
    }

    async login(identifier: string, password: string): Promise<AuthResponse> {
        const formData = new FormData();
        formData.append('identifier', identifier);
        formData.append('password', password);

        const response = await this.api.post<AuthResponse>('/api/login', formData);
        localStorage.setItem('skey1', response.data.skey1);
        localStorage.setItem('skey2', response.data.skey2);
        return response.data;
    }

    // Repository endpoints
    async getRepository(name: string): Promise<RepositoryResponse> {
        const response = await this.api.get<RepositoryResponse>(`/api/repository?name=${encodeURIComponent(name)}`);
        return response.data;
    }

    async createCommit(name: string, message: string, files: File[], tags?: string[]): Promise<{ commit_id: string; commit_hash: string }> {
        const formData = new FormData();
        formData.append('name', name);
        formData.append('message', message);
        files.forEach(file => formData.append('files[]', file));
        if (tags) formData.append('tags', tags.join(','));

        const response = await this.api.post('/api/repository/commit', formData);
        return response.data;
    }

    // Branch operations
    async listBranches(repoName: string): Promise<void> {
        await this.api.get(`/api/repository/${encodeURIComponent(repoName)}/branches`);
    }

    async createBranch(repoName: string, branchName: string): Promise<void> {
        const formData = new FormData();
        formData.append('name', branchName);
        await this.api.post(`/api/repository/${encodeURIComponent(repoName)}/branches`, formData);
    }

    async deleteBranch(repoName: string, branchName: string): Promise<void> {
        await this.api.delete(`/api/repository/${encodeURIComponent(repoName)}/branches/${encodeURIComponent(branchName)}`);
    }

    // Access control
    async grantAccess(repoName: string, username: string, accessLevel: 'read' | 'write'): Promise<void> {
        const formData = new FormData();
        formData.append('username', username);
        formData.append('access_level', accessLevel);
        await this.api.post(`/api/repository/${encodeURIComponent(repoName)}/access`, formData);
    }

    async revokeAccess(repoName: string, username: string): Promise<void> {
        await this.api.delete(`/api/repository/${encodeURIComponent(repoName)}/access/${encodeURIComponent(username)}`);
    }

    async listAccess(repoName: string): Promise<void> {
        await this.api.get(`/api/repository/${encodeURIComponent(repoName)}/access`);
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