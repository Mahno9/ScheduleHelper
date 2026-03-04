<script lang="ts">
    import { currentUser } from '../store';

    export let onCancel: () => void;

    let username = '';
    let color = '#ff0000';
    let emoji = '😀';

    const EMOJIS = ['😀', '😎', '🥳', '🚀', '🌟', '🍕', '🎉', '🎸', '🦄', '🍎', '🌈', '🐶', '🐱', '🦊', '🐻', '🐸', '🐼', '🐨', '🐯', '🦁', '🐮', '🐷', '🐧', '🦉', '🐙', '🐵', '🐣', '🦋', '🐢', '🦖', '🐉', '🐙', '🦑', '🦐', '🦀', '🐡', '🐟', '🐬', '🐳', '🐋'];
    const COLORS = ['#FF5733', '#FF8D1A', '#FFC300', '#DAF7A6', '#33FF57', '#1AFF8D', '#00C3FF', '#33A6FF', '#3357FF', '#1A8DFF', '#0033FF', '#8D1AFF', '#C300FF', '#FF33A6', '#FF1A8D', '#FF0033', '#FF3357'];

    async function register() {
        if (!username || !color || !emoji) {
            alert('Please fill all fields');
            return;
        }

        const res = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, color, emoji })
        });

        if (res.ok) {
            const user = await res.json();
            currentUser.set(user);
        } else {
            alert('Registration failed. Username might be taken.');
        }
    }
</script>

<div class="register-modal">
    <div class="modal-content">
        <h2>Create Profile</h2>
        
        <label>
            Name
            <input type="text" bind:value={username} placeholder="Your name..." />
        </label>

        <label>
            Color
            <div class="colors-grid">
                {#each COLORS as c}
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div 
                        class="color-circle" 
                        class:selected={color === c} 
                        style="background: {c};" 
                        on:click={() => color = c}
                    ></div>
                {/each}
            </div>
            <input type="color" bind:value={color} />
        </label>

        <label>
            Emoji Avatar
            <div class="emoji-grid">
                {#each EMOJIS as e}
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div 
                        class="emoji-item" 
                        class:selected={emoji === e} 
                        on:click={() => emoji = e}
                    >
                        {e}
                    </div>
                {/each}
            </div>
        </label>

        <div class="actions">
            <button class="btn secondary" on:click={onCancel}>Cancel</button>
            <button class="btn primary" on:click={register}>Create</button>
        </div>
    </div>
</div>

<style>
    .register-modal {
        position: fixed;
        inset: 0;
        background: rgba(0,0,0,0.5);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }
    .modal-content {
        background: var(--bg-card, #fff);
        padding: 2rem;
        border-radius: 12px;
        width: 90%;
        max-width: 500px;
        display: flex;
        flex-direction: column;
        gap: 1.5rem;
        max-height: 90vh;
        overflow-y: auto;
    }
    label {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        font-weight: 500;
    }
    input[type="text"] {
        padding: 8px;
        border: 1px solid #ccc;
        border-radius: 4px;
        font-size: 1rem;
    }
    .colors-grid, .emoji-grid {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }
    .color-circle {
        width: 30px;
        height: 30px;
        border-radius: 50%;
        cursor: pointer;
        border: 2px solid transparent;
    }
    .color-circle.selected {
        border-color: #000;
        transform: scale(1.1);
    }
    .emoji-item {
        font-size: 1.5rem;
        cursor: pointer;
        padding: 4px;
        border-radius: 8px;
        background: transparent;
    }
    .emoji-item.selected {
        background: #eee;
    }
    @media (prefers-color-scheme: dark) {
        .modal-content {
            background: #222;
            color: #fff;
        }
        .emoji-item.selected {
            background: #444;
        }
        .color-circle.selected {
            border-color: #fff;
        }
    }
    .actions {
        display: flex;
        justify-content: flex-end;
        gap: 1rem;
        margin-top: 1rem;
    }
    .btn {
        padding: 10px 20px;
        border: none;
        border-radius: 8px;
        font-size: 1rem;
        cursor: pointer;
    }
    .btn.primary { background: #007bff; color: white; }
    .btn.secondary { background: #eee; color: #333; }
</style>