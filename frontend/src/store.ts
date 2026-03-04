import { writable } from 'svelte/store';

export interface User {
    id: string;
    username: string;
    color: string;
    emoji: string;
    theme: string;
    timezone: string;
}

export const currentUser = writable<User | null>(null);
export const allUsers = writable<User[]>([]);
export const currentView = writable<'common' | 'personal'>('common');
