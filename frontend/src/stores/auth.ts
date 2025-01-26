import { defineStore } from 'pinia';
import { api } from '../services/api';
import type { User } from '../types';

interface AuthState {
    user: User | null;
    isAuthenticated: boolean;
}

export const useAuthStore = defineStore('auth', {
    state: (): AuthState => ({
        user: null,
        isAuthenticated: api.isAuthenticated(),
    }),

    actions: {
        async login(identifier: string, password: string) {
            try {
                await api.login(identifier, password);
                this.isAuthenticated = true;
                // TODO: Fetch user profile
            } catch (error) {
                console.error('Login failed:', error);
                throw error;
            }
        },

        async register(username: string, email: string, password: string, firstname?: string, lastname?: string) {
            try {
                await api.register(username, email, password, firstname, lastname);
            } catch (error) {
                console.error('Registration failed:', error);
                throw error;
            }
        },

        logout() {
            api.logout();
            this.user = null;
            this.isAuthenticated = false;
        },
    },
}); 