<script lang="ts">
    import { currentUser, allUsers } from '../store';
    export let eventDraft: any;
    export let onSave: (e: any) => void;
    export let onCancel: () => void;
    export let onDelete: ((id: string) => void) | undefined = undefined;

    let title = eventDraft?.title || '';
    let description = eventDraft?.description || '';
    let icon = eventDraft?.icon || '📅';
    let color = eventDraft?.color || '#FFD700'; // Gold theme default
    let participants: string[] = eventDraft?.participants || [];

    const ICONS = ['📅', '🎉', '🚀', '💡', '🥂', '🔥', '🏆', '⭐'];
    const COLORS = ['#FFD700', '#DAA520', '#B8860B', '#CD853F', '#F0E68C'];

    function save() {
        if (!title) return alert('Title is required');
        onSave({ ...eventDraft, title, description, icon, color, participants });
    }

    function toggleParticipant(id: string) {
        if (participants.includes(id)) {
            participants = participants.filter(p => p !== id);
        } else {
            participants = [...participants, id];
        }
    }
</script>

<div class="modal">
    <div class="modal-content">
        <h2>{eventDraft?.id ? 'Edit Event' : 'New Event'}</h2>
        
        <label>
            Title
            <input type="text" bind:value={title} />
        </label>
        
        <label>
            Description
            <textarea bind:value={description}></textarea>
        </label>
        
        <div class="row">
            <label>
                Icon
                <select bind:value={icon}>
                    {#each ICONS as i}<option value={i}>{i}</option>{/each}
                </select>
            </label>
            <label>
                Color
                <select bind:value={color}>
                    {#each COLORS as c}<option value={c}>{c}</option>{/each}
                </select>
            </label>
        </div>

        <label>
            Participants
            <div class="participants">
                {#each $allUsers as u}
                    <label class="p-checkbox">
                        <input type="checkbox" checked={participants.includes(u.id)} on:change={() => toggleParticipant(u.id)} />
                        {u.emoji} {u.username}
                    </label>
                {/each}
            </div>
        </label>

        <div class="actions">
            {#if eventDraft?.id && onDelete}
                <button class="btn danger" on:click={() => onDelete(eventDraft.id)}>Delete</button>
            {/if}
            <button class="btn secondary" on:click={onCancel}>Cancel</button>
            <button class="btn primary" on:click={save}>Save</button>
        </div>
    </div>
</div>

<style>
    .modal {
        position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 2000;
        display: flex; align-items: center; justify-content: center;
    }
    .modal-content {
        background: var(--bg-card, #fff); padding: 20px; border-radius: 12px;
        width: 400px; max-width: 90vw; display: flex; flex-direction: column; gap: 15px;
    }
    @media (prefers-color-scheme: dark) {
        .modal-content { background: #222; }
    }
    label { display: flex; flex-direction: column; gap: 4px; font-weight: 500; }
    input, textarea, select { padding: 8px; border: 1px solid #ccc; border-radius: 4px; }
    .row { display: flex; gap: 10px; }
    .participants { display: flex; flex-wrap: wrap; gap: 10px; }
    .p-checkbox { flex-direction: row; align-items: center; font-weight: 400; cursor: pointer; }
    .actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 10px; }
    .btn { padding: 8px 16px; border: none; border-radius: 6px; cursor: pointer; }
    .btn.primary { background: #007bff; color: white; }
    .btn.secondary { background: #eee; color: #333; }
    .btn.danger { background: #dc3545; color: white; margin-right: auto; }
</style>