<script lang="ts">
	import { onMount } from "svelte";
	import { users, slots, events, connectSSE } from "../stores";
	import { api } from "../api";
	import type { User, Slot, Event } from "../api";

	// Number of visible days (desktop: 10, mobile: 4)
	let isMobile = typeof window !== "undefined" && window.innerWidth < 768;
	$: numDays = isMobile ? 4 : 10;

	// Date range
	let offsetDays = 0; // scroll offset
	$: startDate = (() => {
		const d = new Date();
		d.setDate(d.getDate() + offsetDays);
		d.setHours(0, 0, 0, 0);
		return d;
	})();
	$: endDate = (() => {
		const d = new Date(startDate);
		d.setDate(d.getDate() + numDays);
		return d;
	})();
	$: dateRange = Array.from({ length: numDays }, (_, i) => {
		const d = new Date(startDate);
		d.setDate(d.getDate() + i);
		return d;
	});

	// Data
	$: activeUsers = $users.filter((u) =>
		$slots.some((s) => s.user_id === u.id),
	);
	$: periodSlots = $slots.filter(
		(s) =>
			new Date(s.end_time) >= startDate &&
			new Date(s.start_time) <= endDate,
	);
	$: periodEvents = $events.filter(
		(e) =>
			new Date(e.end_time) >= startDate &&
			new Date(e.start_time) <= endDate,
	);

	// Drag-and-drop sort
	let dragUser: User | null = null;
	let localOrder: User[] = [];
	$: if (!dragUser) localOrder = [...activeUsers];

	function startDrag(user: User) {
		dragUser = user;
	}
	function dragOver(user: User, e: DragEvent) {
		e.preventDefault();
		if (!dragUser || dragUser.id === user.id) return;
		const fromIdx = localOrder.findIndex((u) => u.id === dragUser!.id);
		const toIdx = localOrder.findIndex((u) => u.id === user.id);
		if (fromIdx < 0 || toIdx < 0) return;
		const next = [...localOrder];
		next.splice(fromIdx, 1);
		next.splice(toIdx, 0, dragUser);
		localOrder = next;
	}
	async function dropSort() {
		if (!dragUser) return;
		dragUser = null;
		await api.updateSortOrder(localOrder.map((u) => u.id));
	}

	// Hover state for events (highlight overlapping users)
	let hoveredEventId: number | null = null;
	$: highlightedUsers = hoveredEventId
		? (() => {
				const ev = periodEvents.find((e) => e.id === hoveredEventId);
				if (!ev) return new Set<number>();
				const evStart = new Date(ev.start_time);
				const evEnd = new Date(ev.end_time);
				const overlapping = new Set<number>();
				for (const s of periodSlots) {
					if (
						new Date(s.start_time) < evEnd &&
						new Date(s.end_time) > evStart
					) {
						overlapping.add(s.user_id);
					}
				}
				return overlapping;
			})()
		: new Set<number>();

	// Event form
	let showEventForm = false;
	let eventForm = {
		title: "",
		description: "",
		icon: "📅",
		color: "#D4AF37",
		start_time: "",
		end_time: "",
		created_by: 0,
	};
	let eventFormError = "";
	let editingEventId: number | null = null;

	const EVENT_ICONS = [
		"📅",
		"🎯",
		"🚀",
		"🎉",
		"💼",
		"🏆",
		"🎊",
		"⭐",
		"🔔",
		"💡",
		"🤝",
		"📊",
	];
	const GOLD_COLORS = [
		"#D4AF37",
		"#C9A227",
		"#F0D060",
		"#B8860B",
		"#FFD700",
		"#DAA520",
		"#EED202",
		"#FCC200",
	];

	function openEventForm(ev?: Event) {
		if (ev) {
			editingEventId = ev.id;
			eventForm = {
				title: ev.title,
				description: ev.description,
				icon: ev.icon,
				color: ev.color,
				start_time: fmtInput(new Date(ev.start_time)),
				end_time: fmtInput(new Date(ev.end_time)),
				created_by: ev.created_by,
			};
		} else {
			editingEventId = null;
			const now = new Date();
			now.setMinutes(0, 0, 0);
			const end = new Date(now);
			end.setHours(end.getHours() + 1);
			eventForm = {
				title: "",
				description: "",
				icon: "📅",
				color: "#D4AF37",
				start_time: fmtInput(now),
				end_time: fmtInput(end),
				created_by: $users[0]?.id || 0,
			};
		}
		eventFormError = "";
		showEventForm = true;
	}

	function fmtInput(d: Date) {
		return d.toISOString().slice(0, 16);
	}

	async function submitEvent() {
		eventFormError = "";
		if (!eventForm.title.trim()) {
			eventFormError = "Заголовок обязателен";
			return;
		}
		try {
			const data = {
				...eventForm,
				start_time: new Date(eventForm.start_time).toISOString(),
				end_time: new Date(eventForm.end_time).toISOString(),
			};
			if (editingEventId) {
				const updated = await api.updateEvent(editingEventId, data);
				events.update((evs) =>
					evs.map((e) => (e.id === updated.id ? updated : e)),
				);
			} else {
				const created = await api.createEvent(data as any);
				events.update((evs) => [...evs, created]);
			}
			showEventForm = false;
		} catch (e: any) {
			eventFormError = e.message;
		}
	}

	async function deleteEvent(id: number) {
		await api.deleteEvent(id);
		events.update((evs) => evs.filter((e) => e.id !== id));
		showEventForm = false;
		editingEventId = null;
	}

	// Event tooltip
	let tooltipEvent: Event | null = null;
	let tooltipX = 0,
		tooltipY = 0;

	function showTooltip(ev: Event, e: MouseEvent) {
		tooltipEvent = ev;
		tooltipX = e.clientX + 12;
		tooltipY = e.clientY - 8;
		hoveredEventId = ev.id;
	}
	function hideTooltip() {
		tooltipEvent = null;
		hoveredEventId = null;
	}

	// Slot tooltip
	let hoveredSlot: Slot | null = null;

	// Date rendering helpers
	function dayFrac(d: Date): number {
		return (d.getTime() - startDate.getTime()) / (numDays * 86400000);
	}
	function eventStyle(ev: Event): string {
		const s = Math.max(0, Math.min(1, dayFrac(new Date(ev.start_time))));
		const e = Math.max(0, Math.min(1, dayFrac(new Date(ev.end_time))));
		return `left: ${s * 100}%; width: ${(e - s) * 100}%; background: ${ev.color}26; border-left: 3px solid ${ev.color};`;
	}

	function slotDayFrac(
		slot: Slot,
		day: Date,
	): { start: number; end: number } {
		const dayStart = day.getTime();
		const dayEnd = dayStart + 86400000;
		const slotStart = new Date(slot.start_time).getTime();
		const slotEnd = new Date(slot.end_time).getTime();
		const s = Math.max(0, (slotStart - dayStart) / 86400000);
		const e = Math.min(1, (slotEnd - dayStart) / 86400000);
		return { start: Math.max(0, s), end: Math.min(1, e) };
	}

	function fmtDay(d: Date) {
		return d.toLocaleDateString("ru-RU", {
			weekday: "short",
			day: "numeric",
			month: "short",
		});
	}
	function fmtTime(iso: string) {
		return new Date(iso).toLocaleTimeString("ru-RU", {
			hour: "2-digit",
			minute: "2-digit",
		});
	}

	// Get color brightness for contrast
	function isDark(hex: string): boolean {
		try {
			const h = hex.replace("#", "");
			const n = parseInt(
				h.length === 3
					? h
							.split("")
							.map((c) => c + c)
							.join("")
					: h,
				16,
			);
			const r = (n >> 16) & 255,
				g = (n >> 8) & 255,
				b = n & 255;
			return (r * 299 + g * 587 + b * 114) / 1000 < 128;
		} catch {
			return true;
		}
	}

	onMount(async () => {
		const from = new Date();
		from.setHours(0, 0, 0, 0);
		const to = new Date(from);
		to.setDate(to.getDate() + 30);
		const [slotsData, eventsData] = await Promise.all([
			api.getSlots({ from: from.toISOString(), to: to.toISOString() }),
			api.getEvents({ from: from.toISOString(), to: to.toISOString() }),
		]);
		slots.set(slotsData);
		events.set(eventsData);
		connectSSE();
	});
</script>

<div class="shared-calendar">
	<!-- Navigation -->
	<div class="nav-bar">
		<button class="btn btn-ghost" on:click={() => (offsetDays -= numDays)}
			>← Назад</button
		>
		<span class="period-label">
			{dateRange[0]?.toLocaleDateString("ru-RU", {
				day: "numeric",
				month: "long",
			})} –
			{dateRange[dateRange.length - 1]?.toLocaleDateString("ru-RU", {
				day: "numeric",
				month: "long",
				year: "numeric",
			})}
		</span>
		<div style="display:flex;gap:0.5rem">
			<button class="btn btn-ghost" on:click={() => (offsetDays = 0)}
				>Сегодня</button
			>
			<button
				class="btn btn-ghost"
				on:click={() => (offsetDays += numDays)}>Вперёд →</button
			>
		</div>
		<button
			class="btn btn-primary"
			on:click={() => openEventForm()}
			id="add-event-btn">+ Событие</button
		>
	</div>

	<!-- Main grid -->
	<div class="grid-scroll">
		<div class="grid-inner">
			<!-- Events header row -->
			<div class="events-header-row">
				<div class="user-col-spacer"></div>
				<div class="days-area">
					<!-- Day headers -->
					<div class="day-headers">
						{#each dateRange as day}
							<div
								class="day-header"
								class:today={day.toDateString() ===
									new Date().toDateString()}
							>
								{fmtDay(day)}
							</div>
						{/each}
					</div>

					<!-- Events overlay -->
					<div class="events-overlay">
						{#each periodEvents as ev (ev.id)}
							<div
								class="event-block"
								style={eventStyle(ev)}
								role="button"
								tabindex="0"
								on:mouseenter={(e) => showTooltip(ev, e)}
								on:mouseleave={hideTooltip}
								on:click={() => openEventForm(ev)}
								on:keydown={(e) =>
									e.key === "Enter" && openEventForm(ev)}
							>
								<span class="event-icon">{ev.icon}</span>
								<span
									class="event-title"
									style="color: {ev.color}"
									title={ev.title}>{ev.title}</span
								>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- User rows -->
			{#each localOrder as user (user.id)}
				<div
					class="user-row"
					class:highlighted={hoveredEventId !== null &&
						highlightedUsers.has(user.id)}
					draggable="true"
					on:dragstart={() => startDrag(user)}
					on:dragover={(e) => dragOver(user, e)}
					on:drop={dropSort}
					on:dragend={() => (dragUser = null)}
				>
					<!-- User label -->
					<div class="user-label">
						<span
							class="drag-handle"
							title="Перетащить для сортировки">⠿</span
						>
						<span
							class="user-name contrast-text"
							style="color: {user.color}; --outline: {isDark(
								user.color,
							)
								? 'rgba(255,255,255,0.6)'
								: 'rgba(0,0,0,0.4)'}"
						>
							{user.emoji}
							{user.name}
						</span>
					</div>

					<!-- Slots for each day -->
					<div class="user-days">
						{#each dateRange as day}
							<div class="user-day-cell">
								{#each periodSlots.filter((s) => s.user_id === user.id && new Date(s.start_time) < new Date(day.getTime() + 86400000) && new Date(s.end_time) > day) as slot (slot.id)}
									{@const frac = slotDayFrac(slot, day)}
									<div
										class="user-slot"
										style="left: {frac.start *
											100}%; width: {(frac.end -
											frac.start) *
											100}%; background: {user.color};"
										role="img"
										on:mouseenter={() =>
											(hoveredSlot = slot)}
										on:mouseleave={() =>
											(hoveredSlot = null)}
										title="{fmtTime(
											slot.start_time,
										)} – {fmtTime(
											slot.end_time,
										)}{slot.comment
											? '\n' + slot.comment
											: ''}"
									></div>
								{/each}
							</div>
						{/each}
					</div>
				</div>
			{/each}

			{#if activeUsers.length === 0}
				<div
					class="empty-state"
					style="padding:2rem;text-align:center;color:var(--text-muted)"
				>
					Никто не отметил свободное время в этом периоде
				</div>
			{/if}
		</div>
	</div>

	<!-- Event tooltip -->
	{#if tooltipEvent}
		<div
			class="event-tooltip card"
			style="position:fixed;left:{Math.min(
				tooltipX,
				window.innerWidth - 340,
			)}px;top:{tooltipY}px;z-index:300;max-width:320px;pointer-events:none"
		>
			<div style="font-size:1.5rem;margin-bottom:0.4rem">
				{tooltipEvent.icon}
			</div>
			<strong>{tooltipEvent.title}</strong>
			<div
				style="font-size:0.8rem;color:var(--text-muted);margin:0.3rem 0"
			>
				{fmtTime(tooltipEvent.start_time)} – {fmtTime(
					tooltipEvent.end_time,
				)}
			</div>
			{#if tooltipEvent.description}
				<div
					style="font-size:0.85rem;white-space:pre-wrap;word-break:break-word"
				>
					{tooltipEvent.description}
				</div>
			{/if}
			<div
				style="margin-top:0.5rem;font-size:0.8rem;color:var(--text-muted)"
			>
				Участников пересечений: {highlightedUsers.size}
			</div>
		</div>
	{/if}

	<!-- Event form modal -->
	{#if showEventForm}
		<div
			class="modal-overlay"
			on:click|self={() => (showEventForm = false)}
		>
			<div class="modal">
				<h3>
					{editingEventId ? "Редактировать событие" : "Новое событие"}
				</h3>

				<div class="field">
					<label class="label">Иконка</label>
					<div class="icon-row">
						{#each EVENT_ICONS as ico}
							<button
								class="icon-btn"
								class:active={eventForm.icon === ico}
								on:click={() => (eventForm.icon = ico)}
								>{ico}</button
							>
						{/each}
					</div>
				</div>

				<div class="field">
					<label class="label">Цвет (оттенок золотого)</label>
					<div class="color-row">
						{#each GOLD_COLORS as c}
							<button
								class="color-dot"
								class:selected={eventForm.color === c}
								style="background:{c}"
								on:click={() => (eventForm.color = c)}
							></button>
						{/each}
					</div>
				</div>

				<div class="field">
					<label class="label" for="ev-title"
						>Заголовок (до 120 симв.)</label
					>
					<input
						id="ev-title"
						type="text"
						bind:value={eventForm.title}
						maxlength="120"
					/>
				</div>

				<div class="field">
					<label class="label" for="ev-desc">Описание</label>
					<textarea
						id="ev-desc"
						bind:value={eventForm.description}
						rows="3"
					></textarea>
				</div>

				<div
					style="display:grid;grid-template-columns:1fr 1fr;gap:0.8rem"
				>
					<div class="field">
						<label class="label">Начало</label>
						<input
							type="datetime-local"
							bind:value={eventForm.start_time}
						/>
					</div>
					<div class="field">
						<label class="label">Конец</label>
						<input
							type="datetime-local"
							bind:value={eventForm.end_time}
						/>
					</div>
				</div>

				{#if eventFormError}<p class="error-msg">
						{eventFormError}
					</p>{/if}

				<div
					style="display:flex;gap:0.6rem;margin-top:1rem;flex-wrap:wrap"
				>
					<button class="btn btn-primary" on:click={submitEvent}
						>{editingEventId ? "Сохранить" : "Создать"}</button
					>
					{#if editingEventId}
						<button
							class="btn btn-danger"
							on:click={() => deleteEvent(editingEventId!)}
							>Удалить</button
						>
					{/if}
					<button
						class="btn btn-ghost"
						on:click={() => (showEventForm = false)}>Отмена</button
					>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.shared-calendar {
		display: flex;
		flex-direction: column;
		flex: 1;
		overflow: hidden;
	}
	.nav-bar {
		display: flex;
		align-items: center;
		gap: 0.8rem;
		padding: 0.75rem 1.5rem;
		border-bottom: 1px solid var(--border);
		background: var(--bg-card);
		flex-wrap: wrap;
	}
	.period-label {
		flex: 1;
		text-align: center;
		font-weight: 500;
		font-size: 0.9rem;
		color: var(--text-muted);
		min-width: 160px;
	}
	.grid-scroll {
		flex: 1;
		overflow: auto;
	}
	.grid-inner {
		min-width: 600px;
	}
	.events-header-row {
		position: sticky;
		top: 0;
		z-index: 20;
		background: var(--bg-card);
		border-bottom: 2px solid var(--border);
		display: flex;
	}
	.user-col-spacer {
		width: 160px;
		min-width: 160px;
		flex-shrink: 0;
		border-right: 1px solid var(--border);
	}
	.days-area {
		flex: 1;
		position: relative;
		overflow: hidden;
	}
	.day-headers {
		display: flex;
	}
	.day-header {
		flex: 1;
		padding: 0.5rem 0.2rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-align: center;
		color: var(--text-muted);
		border-right: 1px solid var(--border);
		white-space: nowrap;
	}
	.day-header.today {
		color: var(--color-primary);
		font-weight: 700;
	}
	.events-overlay {
		position: relative;
		height: 48px;
	}
	.event-block {
		position: absolute;
		top: 4px;
		bottom: 4px;
		border-radius: 5px;
		display: flex;
		align-items: center;
		padding: 0 6px;
		gap: 4px;
		cursor: pointer;
		overflow: hidden;
		transition: opacity 0.15s;
		min-width: 20px;
	}
	.event-block:hover {
		opacity: 0.85;
	}
	.event-icon {
		font-size: 0.85rem;
		flex-shrink: 0;
	}
	.event-title {
		font-size: 0.78rem;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.user-row {
		display: flex;
		align-items: stretch;
		border-bottom: 1px solid var(--border);
		transition: background 0.15s;
		min-height: 44px;
	}
	.user-row.highlighted {
		background: rgba(212, 175, 55, 0.08);
	}
	.user-label {
		width: 160px;
		min-width: 160px;
		flex-shrink: 0;
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.4rem 0.6rem;
		border-right: 1px solid var(--border);
		overflow: hidden;
	}
	.drag-handle {
		color: var(--text-muted);
		cursor: grab;
		font-size: 1rem;
		flex-shrink: 0;
		opacity: 0.5;
	}
	.drag-handle:active {
		cursor: grabbing;
	}
	.user-name {
		font-size: 0.82rem;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.user-days {
		flex: 1;
		display: flex;
	}
	.user-day-cell {
		flex: 1;
		border-right: 1px solid var(--border);
		position: relative;
		min-height: 44px;
	}
	.user-slot {
		position: absolute;
		top: 6px;
		bottom: 6px;
		border-radius: 4px;
		opacity: 0.75;
		transition: opacity 0.15s;
		cursor: default;
		min-width: 3px;
	}
	.user-slot:hover {
		opacity: 1;
	}
	.icon-row {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}
	.icon-btn {
		font-size: 1.2rem;
		padding: 4px 6px;
		border-radius: 6px;
		background: var(--bg-input);
		border: 1.5px solid transparent;
		cursor: pointer;
		transition: all 0.12s;
	}
	.icon-btn.active {
		border-color: var(--color-primary);
		background: var(--bg-card);
	}
	.icon-btn:hover {
		transform: scale(1.1);
	}
	.color-row {
		display: flex;
		flex-wrap: wrap;
		gap: 7px;
	}
	.color-dot {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		border: 2.5px solid transparent;
		cursor: pointer;
		transition: transform 0.12s;
	}
	.color-dot:hover {
		transform: scale(1.18);
	}
	.color-dot.selected {
		border-color: white;
		box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.3);
	}
	h3 {
		margin-bottom: 1rem;
		font-size: 1.1rem;
	}
	.field {
		margin-bottom: 0.8rem;
	}
	textarea {
		resize: vertical;
	}
</style>
