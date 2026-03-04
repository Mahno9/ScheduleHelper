<script lang="ts">
    import { onMount } from 'svelte';
    import { currentUser, allUsers } from '../store';
    import EventForm from './EventForm.svelte';

    let data: any = { users: [], slots: [], events: [] };
    let days: Date[] = [];
    let numDays = 10;
    
    $: activeUsers = data.users.filter((u: any) => data.slots.some((s: any) => s.user_id === u.id));
    $: if (activeUsers && activeUsers.length === 0) {
        activeUsers = data.users;
    }

    for (let i = 0; i < numDays; i++) {
        let d = new Date();
        d.setDate(d.getDate() + i);
        d.setHours(0,0,0,0);
        days.push(d);
    }

    onMount(async () => {
        await loadData();
    });

    async function loadData() {
        const res = await fetch('/api/calendar');
        if (res.ok) {
            data = await res.json();
            allUsers.set(data.users);
        }
    }

    let showEventForm = false;
    let currentEventDraft: any = null;

    function newEvent() {
        let st = new Date();
        let en = new Date(st.getTime() + 3600000);
        currentEventDraft = { start_time: st.toISOString(), end_time: en.toISOString(), participants: [$currentUser?.id] };
        showEventForm = true;
    }

    function editEvent(e: any) {
        currentEventDraft = e;
        showEventForm = true;
    }

    async function saveEvent(e: any) {
        const method = e.id ? 'PUT' : 'POST';
        await fetch('/api/events', {
            method,
            headers: { 'Content-Type': 'application/json', 'X-User-ID': $currentUser?.id || '' },
            body: JSON.stringify(e)
        });
        showEventForm = false;
        await loadData();
    }

    async function deleteEvent(id: string) {
        await fetch(`/api/events?id=${id}`, {
            method: 'DELETE',
            headers: { 'X-User-ID': $currentUser?.id || '' }
        });
        showEventForm = false;
        await loadData();
    }

    function getSlotsForUser(userId: string, day: Date) {
        const nextDay = new Date(day);
        nextDay.setDate(nextDay.getDate() + 1);
        return data.slots.filter((s: any) => 
            s.user_id === userId && 
            new Date(s.start_time) < nextDay && 
            new Date(s.end_time) > day
        );
    }
</script>

<div class="toolbar">
    <h2>Общий календарь</h2>
    <button class="btn primary" on:click={newEvent}>+ Создать событие</button>
</div>

<div class="calendar-wrapper">
    <div class="users-col">
        <div class="corner-header">Участники</div>
        {#each activeUsers as u}
            <div class="user-row-header">
                <span class="user-name" style="color: {u.color}; text-shadow: -1px -1px 0 #fff, 1px -1px 0 #fff, -1px 1px 0 #fff, 1px 1px 0 #fff;">
                    {u.emoji} {u.username}
                </span>
            </div>
        {/each}
    </div>

    <div class="grid-col">
        <div class="days-header-row">
            {#each days as day}
                <div class="day-col-header">
                    {day.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })}
                </div>
            {/each}
        </div>
        
        <div class="grid-body" style="position: relative;">
            {#each activeUsers as u}
                <div class="grid-row">
                    {#each days as day}
                        <div class="grid-cell">
                            {#each getSlotsForUser(u.id, day) as slot}
                                <div class="slot-block" style="background: {u.color}aa;" title="{new Date(slot.start_time).toLocaleTimeString()} - {new Date(slot.end_time).toLocaleTimeString()}\n{slot.comment}"></div>
                            {/each}
                        </div>
                    {/each}
                </div>
            {/each}

            <div class="events-overlay">
                {#each data.events as ev}
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div class="event-card" style="background: {ev.color};" on:click={() => editEvent(ev)}>
                        <div class="event-icon">{ev.icon}</div>
                        <div class="event-title">{ev.title}</div>
                    </div>
                {/each}
            </div>
        </div>
    </div>
</div>

{#if showEventForm}
    <EventForm 
        eventDraft={currentEventDraft} 
        onSave={saveEvent} 
        onCancel={() => showEventForm = false} 
        onDelete={deleteEvent}
    />
{/if}

<style>
    .toolbar { display: flex; justify-content: space-between; align-items: center; padding: 10px 20px; }
    .btn { padding: 8px 16px; border: none; border-radius: 6px; cursor: pointer; }
    .btn.primary { background: #007bff; color: white; }

    .calendar-wrapper { display: flex; flex: 1; border-top: 1px solid #eee; overflow: hidden; }
    .users-col { width: 150px; flex-shrink: 0; border-right: 1px solid #eee; display: flex; flex-direction: column; background: #fafafa; z-index: 2; }
    @media (prefers-color-scheme: dark) { .users-col { background: #1a1a1a; border-color: #333; } .calendar-wrapper { border-color: #333; } }
    
    .corner-header { height: 50px; display: flex; align-items: center; justify-content: center; font-weight: bold; border-bottom: 1px solid #eee; }
    .user-row-header { height: 60px; display: flex; align-items: center; padding: 0 10px; border-bottom: 1px solid #eee; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

    .grid-col { flex: 1; display: flex; flex-direction: column; overflow-x: auto; }
    .days-header-row { display: flex; height: 50px; border-bottom: 1px solid #eee; background: #fff; }
    @media (prefers-color-scheme: dark) { .days-header-row, .corner-header, .user-row-header { border-color: #333; background: #121212; } }
    .day-col-header { flex: 0 0 120px; display: flex; align-items: center; justify-content: center; border-right: 1px solid #eee; font-weight: 500; }

    .grid-body { display: flex; flex-direction: column; }
    .grid-row { display: flex; height: 60px; border-bottom: 1px solid #eee; }
    .grid-cell { flex: 0 0 120px; border-right: 1px solid #eee; padding: 4px; display: flex; flex-direction: column; gap: 2px; }
    .slot-block { height: 8px; border-radius: 4px; flex-shrink: 0; }

    .events-overlay { position: absolute; top: 0; left: 0; right: 0; bottom: 0; pointer-events: none; display: flex; padding: 10px; gap: 10px; }
    .event-card { width: 100px; background: gold; border-radius: 8px; padding: 8px; pointer-events: auto; cursor: pointer; box-shadow: 0 4px 6px rgba(0,0,0,0.1); height: max-content; }
    .event-icon { font-size: 24px; text-align: center; }
    .event-title { font-size: 12px; font-weight: 600; text-align: center; margin-top: 4px; overflow: hidden; text-overflow: ellipsis; color: #333; }
</style>