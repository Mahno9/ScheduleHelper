<script lang="ts">
    import { onMount } from 'svelte';
    import { currentUser } from '../store';

    interface Slot {
        id?: string;
        user_id?: string;
        start_time: string;
        end_time: string;
        comment: string;
    }

    let slots: Slot[] = [];
    let days: Date[] = [];
    let hours: string[] = [];

    // Settings
    let numDays = 14;
    let startHour = 7;
    let endHour = 24; // 00:00 next day

    // Initialize calendar grid
    for (let i = 0; i < numDays; i++) {
        let d = new Date();
        d.setDate(d.getDate() + i);
        d.setHours(0,0,0,0);
        days.push(d);
    }
    
    for (let h = startHour; h < endHour; h++) {
        hours.push(`${h.toString().padStart(2, '0')}:00`);
        hours.push(`${h.toString().padStart(2, '0')}:30`);
    }

    onMount(async () => {
        await loadSlots();
    });

    async function loadSlots() {
        if (!$currentUser) return;
        const res = await fetch('/api/slots', {
            headers: { 'X-User-ID': $currentUser.id }
        });
        if (res.ok) {
            slots = await res.json();
        }
    }

    async function saveSlot(s: Slot) {
        if (!$currentUser) return;
        const method = s.id ? 'PUT' : 'POST';
        const res = await fetch('/api/slots', {
            method,
            headers: { 
                'Content-Type': 'application/json',
                'X-User-ID': $currentUser.id
            },
            body: JSON.stringify(s)
        });
        if (res.ok) {
            await loadSlots();
        }
    }

    async function deleteSlot(id: string) {
        if (!$currentUser) return;
        const res = await fetch(`/api/slots?id=${id}`, {
            method: 'DELETE',
            headers: { 'X-User-ID': $currentUser.id }
        });
        if (res.ok) {
            await loadSlots();
        }
    }

    // Interaction state
    let isDrawing = false;
    let dragStartCell: { dayIdx: number, hourIdx: number } | null = null;
    let currentDraft: { dayIdx: number, startHourIdx: number, endHourIdx: number } | null = null;

    function handleMouseDown(dayIdx: number, hourIdx: number) {
        isDrawing = true;
        dragStartCell = { dayIdx, hourIdx };
        currentDraft = { dayIdx, startHourIdx: hourIdx, endHourIdx: hourIdx };
    }

    function handleMouseEnter(dayIdx: number, hourIdx: number) {
        if (!isDrawing || !dragStartCell || dragStartCell.dayIdx !== dayIdx) return;
        currentDraft = {
            dayIdx,
            startHourIdx: Math.min(dragStartCell.hourIdx, hourIdx),
            endHourIdx: Math.max(dragStartCell.hourIdx, hourIdx)
        };
    }

    function handleMouseUp() {
        if (isDrawing && currentDraft) {
            createSlotFromDraft(currentDraft);
        }
        isDrawing = false;
        dragStartCell = null;
        currentDraft = null;
    }

    function createSlotFromDraft(draft: { dayIdx: number, startHourIdx: number, endHourIdx: number }) {
        const day = days[draft.dayIdx];
        
        const startH = startHour + Math.floor(draft.startHourIdx / 2);
        const startM = (draft.startHourIdx % 2) * 30;
        
        const endIdx = draft.endHourIdx + 1; // exclusive end
        const endH = startHour + Math.floor(endIdx / 2);
        const endM = (endIdx % 2) * 30;

        const startT = new Date(day);
        startT.setHours(startH, startM, 0, 0);

        const endT = new Date(day);
        endT.setHours(endH, endM, 0, 0);

        saveSlot({
            start_time: startT.toISOString(),
            end_time: endT.toISOString(),
            comment: ''
        });
    }

    function isDraft(dayIdx: number, hourIdx: number) {
        if (!currentDraft) return false;
        return currentDraft.dayIdx === dayIdx && 
               hourIdx >= currentDraft.startHourIdx && 
               hourIdx <= currentDraft.endHourIdx;
    }

    function hasSlot(day: Date, hourIdx: number) {
        const checkH = startHour + Math.floor(hourIdx / 2);
        const checkM = (hourIdx % 2) * 30;
        const cellStart = new Date(day);
        cellStart.setHours(checkH, checkM, 0, 0);
        
        const cellEnd = new Date(day);
        cellEnd.setHours(checkH, checkM + 30, 0, 0);

        return slots.find(s => {
            const ss = new Date(s.start_time).getTime();
            const se = new Date(s.end_time).getTime();
            const cs = cellStart.getTime();
            const ce = cellEnd.getTime();
            // simple overlap check
            return ss < ce && se > cs;
        });
    }

    function onSlotClick(s: Slot) {
        const c = prompt('Comment for this slot:', s.comment);
        if (c !== null) {
            if (c === '') {
                // simple hack: if user clears comment and wants to delete, let's just ask
                if (confirm('Delete this slot?')) {
                    if (s.id) deleteSlot(s.id);
                    return;
                }
            }
            saveSlot({ ...s, comment: c });
        }
    }

</script>

<svelte:window on:mouseup={handleMouseUp} on:app-update={loadSlots} />

<div class="calendar-container">
    <div class="header-row">
        <div class="time-col-header">Time</div>
        {#each days as day}
            <div class="day-header">
                {day.toLocaleDateString(undefined, { weekday: 'short', month: 'numeric', day: 'numeric' })}
            </div>
        {/each}
    </div>

    <div class="grid">
        {#each hours as hour, hourIdx}
            <div class="row">
                <div class="time-label">{hour}</div>
                {#each days as day, dayIdx}
                    {@const slot = hasSlot(day, hourIdx)}
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div 
                        class="cell" 
                        class:is-draft={isDraft(dayIdx, hourIdx)}
                        class:has-slot={!!slot}
                        style={slot ? `background-color: ${$currentUser?.color}` : ''}
                        on:mousedown={() => { if (!slot) handleMouseDown(dayIdx, hourIdx); }}
                        on:mouseenter={() => { if (!slot) handleMouseEnter(dayIdx, hourIdx); }}
                        on:click={() => { if (slot) onSlotClick(slot); }}
                    >
                        {#if slot && hourIdx % 2 === 0}
                            <span class="slot-comment">{slot.comment || ''}</span>
                        {/if}
                    </div>
                {/each}
            </div>
        {/each}
    </div>
</div>

<style>
    .calendar-container {
        overflow-x: auto;
        padding: 10px;
        user-select: none;
    }
    .header-row {
        display: flex;
        position: sticky;
        top: 0;
        background: var(--bg-card, #fff);
        z-index: 10;
        border-bottom: 2px solid #eee;
    }
    @media (prefers-color-scheme: dark) {
        .header-row { background: #121212; border-color: #333; }
    }
    .time-col-header, .day-header {
        flex: 1;
        min-width: 100px;
        padding: 10px;
        text-align: center;
        font-weight: 500;
    }
    .time-col-header {
        flex: 0 0 60px;
        min-width: 60px;
    }
    .row {
        display: flex;
    }
    .time-label {
        flex: 0 0 60px;
        min-width: 60px;
        padding-right: 10px;
        text-align: right;
        font-size: 0.8em;
        color: #888;
        display: flex;
        align-items: center;
        justify-content: flex-end;
    }
    .cell {
        flex: 1;
        min-width: 100px;
        height: 30px;
        border-right: 1px solid #eee;
        border-bottom: 1px solid #eee;
        cursor: crosshair;
        position: relative;
        transition: background-color 0.1s;
    }
    @media (prefers-color-scheme: dark) {
        .cell { border-color: #333; }
    }
    .cell:hover:not(.has-slot) {
        background-color: rgba(0, 123, 255, 0.1);
    }
    .cell.is-draft {
        background-color: rgba(0, 123, 255, 0.3);
    }
    .cell.has-slot {
        cursor: pointer;
        border-right-color: transparent;
        border-bottom-color: transparent;
    }
    .slot-comment {
        font-size: 0.75em;
        color: white;
        padding-left: 4px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        position: absolute;
        top: 2px;
        left: 2px;
        right: 2px;
        text-shadow: 0 1px 2px rgba(0,0,0,0.5);
    }
</style>