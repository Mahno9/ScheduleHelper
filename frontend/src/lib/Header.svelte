<script lang="ts">
    import { currentUser, currentView } from '../store';
    import Profile from './Profile.svelte';

    let showSettings = false;

    function toggleSettings() {
        showSettings = !showSettings;
    }

    async function updateProfile(fields: Partial<typeof $currentUser>) {
        if (!$currentUser) return;
        const res = await fetch('/api/profile', {
            method: 'PUT',
            headers: { 
                'Content-Type': 'application/json',
                'X-User-ID': $currentUser.id
            },
            body: JSON.stringify(fields)
        });
        if (res.ok) {
            currentUser.update(u => ({ ...u, ...fields }));
        }
    }

    function logout() {
        currentUser.set(null);
    }
</script>

<header>
    <div class="tabs">
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="tab" class:active={$currentView === 'common'} on:click={() => currentView.set('common')}>
            Общий календарь
        </div>
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="tab" class:active={$currentView === 'personal'} on:click={() => currentView.set('personal')}>
            Личный календарь
        </div>
    </div>

    <div class="profile-section">
        {#if $currentUser}
            <Profile user={$currentUser} onClick={toggleSettings} />
            
            {#if showSettings}
                <div class="settings-dropdown">
                    <button class="btn secondary" on:click={logout}>Выйти</button>
                </div>
            {/if}
        {/if}
    </div>
</header>

<style>
    header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0 20px;
        height: 60px;
        background: var(--bg-header, #fff);
        border-bottom: 1px solid #eee;
        position: relative;
    }
    .tabs {
        display: flex;
        gap: 20px;
        height: 100%;
    }
    .tab {
        display: flex;
        align-items: center;
        padding: 0 10px;
        cursor: pointer;
        font-weight: 500;
        border-bottom: 3px solid transparent;
        color: #666;
    }
    .tab:hover {
        color: #000;
    }
    .tab.active {
        border-bottom-color: #007bff;
        color: #000;
    }
    .profile-section {
        position: relative;
    }
    .settings-dropdown {
        position: absolute;
        top: 100%;
        right: 0;
        background: #fff;
        border: 1px solid #eee;
        box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        border-radius: 8px;
        padding: 10px;
        margin-top: 10px;
        z-index: 100;
    }
    @media (prefers-color-scheme: dark) {
        header {
            background: #1e1e1e;
            border-bottom-color: #333;
        }
        .tab {
            color: #aaa;
        }
        .tab:hover {
            color: #fff;
        }
        .tab.active {
            color: #fff;
            border-bottom-color: #4da3ff;
        }
        .settings-dropdown {
            background: #2a2a2a;
            border-color: #444;
        }
    }
    .btn {
        padding: 8px 16px;
        border: none;
        border-radius: 6px;
        cursor: pointer;
    }
    .btn.secondary { background: #eee; color: #333; }
</style>