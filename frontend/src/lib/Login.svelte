<script lang="ts">
    import { onMount } from 'svelte';
    import { allUsers, currentUser } from '../store';
    import Profile from './Profile.svelte';

    export let onRegister: () => void;

    onMount(async () => {
        try {
            const res = await fetch('/api/users');
            if (res.ok) {
                const users = await res.json();
                allUsers.set(users);
            }
        } catch (e) {
            console.error("Failed to fetch users", e);
        }
    });

    async function login(id: string) {
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id })
        });
        if (res.ok) {
            const user = await res.json();
            currentUser.set(user);
        } else {
            alert('Login failed');
        }
    }
</script>

<div class="login-screen">
    <h1>ScheduleHelper</h1>
    <p>Select your profile to continue</p>

    <div class="users-grid">
        {#each $allUsers as user}
            <Profile {user} size="large" onClick={() => login(user.id)} />
        {/each}
    </div>

    <button class="btn primary" on:click={onRegister}>Create New Profile</button>
</div>

<style>
    .login-screen {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        gap: 2rem;
    }
    .users-grid {
        display: flex;
        flex-wrap: wrap;
        gap: 2rem;
        justify-content: center;
        max-width: 800px;
    }
    .btn.primary {
        padding: 12px 24px;
        background: #007bff;
        color: white;
        border: none;
        border-radius: 8px;
        font-size: 1.1em;
        cursor: pointer;
        transition: background 0.2s;
    }
    .btn.primary:hover {
        background: #0056b3;
    }
</style>