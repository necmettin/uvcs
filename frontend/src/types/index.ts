export interface User {
    id: number;
    username: string;
    email: string;
    firstname?: string;
    lastname?: string;
}

export interface Repository {
    id: number;
    name: string;
    owner_id: number;
    created_at: string;
}

export interface Branch {
    id: number;
    name: string;
    repository_id: number;
    head_commit_id: number;
}

export interface Commit {
    id: number;
    hash: string;
    message: string;
    repository_id: number;
    user_id: number;
    datetime: string;
    tags: string[];
    author: string;
    changes: CommitDetail[];
}

export interface CommitDetail {
    id: number;
    commit_id: number;
    file_path: string;
    change_type: 'A' | 'M' | 'D'; // Added, Modified, Deleted
    content_change: string;
    is_binary: boolean;
    is_diff: boolean;
}

export interface RepositoryAccess {
    repository_id: number;
    user_id: number;
    access_level: 'read' | 'write';
    granted_by: number;
    granted_at: string;
}

export interface AuthResponse {
    skey1: string;
    skey2: string;
}

export interface ApiError {
    error: string;
}

export interface RepositoryResponse {
    branches: Branch[];
    commits: Commit[];
    content: { [key: string]: string };
    access: RepositoryAccess[];
} 