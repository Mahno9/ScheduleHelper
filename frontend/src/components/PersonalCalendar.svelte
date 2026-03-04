<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import {
		currentUser,
		slots,
		events,
		undoHistory,
		pushUndo,
		connectSSE,
	} from "../stores";
	import { api } from "../api";
	import type { Slot } from "../api";

	// Settings
	let numDays = 14;
	let startHour = 7;
	let endHour = 24; // 00:00 = midnight

	// Days array (today + numDays days)
	$: days = Array.from({ length: numDays }, (_, i) => {
		const d = new Date();
		d.setHours(0, 0, 0, 0);
		d.setDate(d.getDate() + i);
		return d;
	});

	// Hours array
	$: hours = Array.from(
		{ length: endHour - startHour },
		(_, i) => startHour + i,
	);

	// User slots (filtered to current user)
	$: userSlots = $slots.filter((s) => s.user_id === $currentUser?.id);

	// Drag state
	let dragStart: { day: Date; hour: number; minute: number } | null = null;
	let dragEnd: { hour: number; minute: number } | null = null;
	let isDragging = false;

	// Edit slot modal
	let editSlot: Slot | null = null;
	let editStart = "";
	let editEnd = "";
	let editComment = "";
	let editLoading = false;
	let editError = "";

	onMount(async () => {
		await loadSlots();
		connectSSE();
	});

	async function loadSlots() {
		if (!$currentUser) return;
		const from = new Date();
		from.setHours(0, 0, 0, 0);
		const to = new Date(from);
		to.setDate(to.getDate() + numDays + 1);
		const data = await api.getSlots({
			user_id: $currentUser.id,
			from: from.toISOString(),
			to: to.toISOString(),
		});
		slots.set(data);
	}

	// Snap to 30-min grid
	function snapMinute(m: number): number {
		return m < 30 ? 0 : 30;
	}

	// Cell time helpers
	function cellToDate(day: Date, hour: number, minute: number = 0): Date {
		const d = new Date(day);
		d.setHours(hour, minute, 0, 0);
		return d;
	}

	// Check if a slot occupies a particular cell
	function slotInCell(slot: Slot, day: Date, hour: number): boolean {
		const start = new Date(slot.start_time);
		const end = new Date(slot.end_time);
		const cellStart = cellToDate(day, hour);
		const cellEnd = cellToDate(day, hour + 1);
		return start < cellEnd && end > cellStart;
	}

	// Compute slot position within a cell (for partial slots)
	function slotCellStyle(slot: Slot, day: Date, hour: number) {
		const start = new Date(slot.start_time);
		const end = new Date(slot.end_time);
		const cellStart = cellToDate(day, hour).getTime();
		const cellEnd = cellToDate(day, hour + 1).getTime();
		const slotStart = Math.max(start.getTime(), cellStart);
		const slotEnd = Math.min(end.getTime(), cellEnd);
		const duration = cellEnd - cellStart;
		const top = ((slotStart - cellStart) / duration) * 100;
		const height = ((slotEnd - slotStart) / duration) * 100;
		return `top: ${top}%; height: ${height}%;`;
	}

	// Mouse events for drag-to-create
	function onMouseDown(day: Date, hour: number, e: MouseEvent) {
		if (e.target instanceof HTMLElement && e.target.closest(".slot-block"))
			return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const relY = e.clientY - rect.top;
		const minute = snapMinute(Math.floor((relY / rect.height) * 60));
		dragStart = { day, hour, minute };
		dragEnd = { hour, minute: minute + 30 > 59 ? 30 : minute + 30 };
		isDragging = true;
	}

	function onMouseEnter(hour: number, e: MouseEvent) {
		if (!isDragging || !dragStart) return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const relY = e.clientY - rect.top;
		const minute = snapMinute(Math.floor((relY / rect.height) * 60));
		dragEnd = { hour, minute };
	}

	async function onMouseUp() {
		if (!isDragging || !dragStart || !dragEnd || !$currentUser) {
			isDragging = false;
			dragStart = null;
			dragEnd = null;
			return;
		}
		isDragging = false;

		let startDate = cellToDate(
			dragStart.day,
			dragStart.hour,
			dragStart.minute,
		);
		let endDate = cellToDate(dragStart.day, dragEnd.hour, dragEnd.minute);

		// Normalize (ensure start < end and at least 30 min)
		if (endDate <= startDate)
			endDate = new Date(startDate.getTime() + 30 * 60000);

		try {
			const slot = await api.createSlot({
				user_id: $currentUser.id,
				start_time: startDate.toISOString(),
				end_time: endDate.toISOString(),
				comment: "",
			});
			slots.update((ss) => [...ss, slot]);
			pushUndo({ type: "create_slot", slot });
		} catch {}
		dragStart = null;
		dragEnd = null;
	}

	// Click on slot to edit
	function openEdit(slot: Slot) {
		editSlot = slot;
		editStart = fmtInputDatetime(new Date(slot.start_time));
		editEnd = fmtInputDatetime(new Date(slot.end_time));
		editComment = slot.comment;
		editError = "";
	}

	function fmtInputDatetime(d: Date) {
		return d.toISOString().slice(0, 16);
	}

	async function saveEdit() {
		if (!editSlot) return;
		editLoading = true;
		editError = "";
		try {
			const start = new Date(editStart).toISOString();
			const end = new Date(editEnd).toISOString();
			const prev = { ...editSlot };
			const updated = await api.updateSlot(editSlot.id, {
				start_time: start,
				end_time: end,
				comment: editComment,
			});
			slots.update((ss) =>
				ss.map((s) => (s.id === updated.id ? updated : s)),
			);
			pushUndo({ type: "update_slot", slot: updated, prevSlot: prev });
			editSlot = null;
		} catch (e: any) {
			editError = e.message;
		} finally {
			editLoading = false;
		}
	}

	async function deleteEditSlot() {
		if (!editSlot) return;
		editLoading = true;
		const prev = { ...editSlot };
		try {
			await api.deleteSlot(editSlot.id);
			slots.update((ss) => ss.filter((s) => s.id !== prev.id));
			pushUndo({ type: "delete_slot", slot: prev });
			editSlot = null;
		} catch (e: any) {
			editError = e.message;
		} finally {
			editLoading = false;
		}
	}

	async function splitSlot() {
		if (!editSlot) return;
		// Split at exact click time (use current instant snapped to 30 min)
		const now = new Date();
		const snapMin = snapMinute(now.getMinutes());
		const splitPoint = new Date(now);
		splitPoint.setSeconds(0, 0);
		splitPoint.setMinutes(snapMin);

		const start = new Date(editSlot.start_time);
		const end = new Date(editSlot.end_time);
		if (splitPoint <= start || splitPoint >= end) {
			editError = "Точка разрыва вне границ слота";
			return;
		}
		editLoading = true;
		try {
			const prev = { ...editSlot };
			// Update original to end at split
			const updated = await api.updateSlot(editSlot.id, {
				start_time: start.toISOString(),
				end_time: splitPoint.toISOString(),
				comment: editSlot.comment,
			});
			// Create second part
			const newSlot = await api.createSlot({
				user_id: editSlot.user_id,
				start_time: splitPoint.toISOString(),
				end_time: end.toISOString(),
				comment: editSlot.comment,
			});
			slots.update((ss) => [
				...ss.map((s) => (s.id === updated.id ? updated : s)),
				newSlot,
			]);
			pushUndo({ type: "update_slot", slot: updated, prevSlot: prev });
			editSlot = null;
		} catch (e: any) {
			editError = e.message;
		} finally {
			editLoading = false;
		}
	}

	async function undo() {
		const history = $undoHistory;
		if (history.length === 0) return;
		const action = history[history.length - 1];
		undoHistory.update((h) => h.slice(0, -1));

		try {
			if (action.type === "create_slot" && action.slot) {
				await api.deleteSlot(action.slot.id);
				slots.update((ss) =>
					ss.filter((s) => s.id !== action.slot!.id),
				);
			} else if (action.type === "delete_slot" && action.slot) {
				const s = action.slot;
				const restored = await api.createSlot({
					user_id: s.user_id,
					start_time: s.start_time,
					end_time: s.end_time,
					comment: s.comment,
				});
				slots.update((ss) => [...ss, restored]);
			} else if (action.type === "update_slot" && action.prevSlot) {
				const p = action.prevSlot;
				const restored = await api.updateSlot(p.id, {
					start_time: p.start_time,
					end_time: p.end_time,
					comment: p.comment,
				});
				slots.update((ss) =>
					ss.map((s) => (s.id === restored.id ? restored : s)),
				);
			}
		} catch {}
	}

	// Format helpers
	function fmtDay(d: Date) {
		return d.toLocaleDateString("ru-RU", {
			weekday: "short",
			day: "numeric",
			month: "short",
		});
	}
	function fmtHour(h: number) {
		return `${String(h === 24 ? 0 : h).padStart(2, "0")}:00`;
	}
	function fmtTime(iso: string) {
		return new Date(iso).toLocaleTimeString("ru-RU", {
			hour: "2-digit",
			minute: "2-digit",
		});
	}

	$: color = $currentUser?.color || "#4A90D9";
	$: hasUndo = $undoHistory.length > 0;
</script>

<svelte:window on:mouseup={onMouseUp} />

<div class="personal-calendar">
	<!-- Top controls -->
	<div class="controls">
		<div class="control-group">
			<label
				>Дней: <input
					type="number"
					bind:value={numDays}
					min="1"
					max="30"
					style="width:60px"
				/></label
			>
			<label
				>С: <input
					type="number"
					bind:value={startHour}
					min="0"
					max="23"
					style="width:55px"
				/></label
			>
			<label
				>До: <input
					type="number"
					bind:value={endHour}
					min="1"
					max="24"
					style="width:55px"
				/></label
			>
		</div>
		<button class="btn btn-ghost" on:click={loadSlots}>🔄 Обновить</button>
	</div>

	<!-- Scroll container -->
	<div class="calendar-scroll">
		<table class="calendar-table">
			<thead>
				<tr>
					<th class="time-col"></th>
					{#each days as day}
						<th
							class="day-header"
							class:today={day.toDateString() ===
								new Date().toDateString()}
						>
							{fmtDay(day)}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each hours as hour}
					<tr>
						<td class="time-label">{fmtHour(hour)}</td>
						{#each days as day}
							<td
								class="calendar-cell droppable"
								role="gridcell"
								on:mousedown={(e) => onMouseDown(day, hour, e)}
								on:mouseenter={(e) => onMouseEnter(hour, e)}
							>
								<!-- Existing slots -->
								{#each userSlots.filter( (s) => slotInCell(s, day, hour), ) as slot (slot.id)}
									<button
										class="slot-block"
										style="background: {color}; {slotCellStyle(
											slot,
											day,
											hour,
										)}"
										on:click|stopPropagation={() =>
											openEdit(slot)}
										title="{fmtTime(
											slot.start_time,
										)} – {fmtTime(
											slot.end_time,
										)}{slot.comment
											? '\n' + slot.comment
											: ''}"
										aria-label="Слот {fmtTime(
											slot.start_time,
										)} – {fmtTime(slot.end_time)}"
									>
										{#if slot.comment}
											<span class="slot-comment"
												>{slot.comment}</span
											>
										{/if}
									</button>
								{/each}

								<!-- Drag preview -->
								{#if isDragging && dragStart && dragEnd && dragStart.day.toDateString() === day.toDateString()}
									{#if dragStart.hour === hour || (dragEnd.hour === hour && dragEnd.hour >= dragStart.hour)}
										<div
											class="drag-preview"
											style="background: {color}40; border: 2px dashed {color}"
										></div>
									{/if}
								{/if}
							</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<!-- Undo float button -->
	{#if hasUndo}
		<button
			class="float-btn undo-btn"
			on:click={undo}
			style="right: 2rem"
			title="Отменить"
		>
			↩
		</button>
	{/if}

	<!-- Edit slot modal -->
	{#if editSlot}
		<div class="modal-overlay" on:click|self={() => (editSlot = null)}>
			<div class="modal">
				<h3>Редактировать слот</h3>
				<div class="field">
					<label class="label">Начало</label><input
						type="datetime-local"
						bind:value={editStart}
					/>
				</div>
				<div class="field">
					<label class="label">Конец</label><input
						type="datetime-local"
						bind:value={editEnd}
					/>
				</div>
				<div class="field">
					<label class="label">Комментарий</label><textarea
						bind:value={editComment}
						rows="2"
					></textarea>
				</div>
				{#if editError}<p class="error-msg">{editError}</p>{/if}
				<div class="edit-actions">
					<button
						class="btn btn-primary"
						on:click={saveEdit}
						disabled={editLoading}>Сохранить</button
					>
					<button
						class="btn btn-ghost"
						on:click={splitSlot}
						disabled={editLoading}
						title="Разбить слот на два в текущий момент"
						>✂ Разбить</button
					>
					<button
						class="btn btn-danger"
						on:click={deleteEditSlot}
						disabled={editLoading}>Удалить</button
					>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.personal-calendar {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		position: relative;
	}
	.controls {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1.5rem;
		border-bottom: 1px solid var(--border);
		background: var(--bg-card);
		gap: 1rem;
		flex-wrap: wrap;
	}
	.control-group {
		display: flex;
		gap: 1rem;
		align-items: center;
		flex-wrap: wrap;
	}
	.control-group label {
		font-size: 0.85rem;
		color: var(--text-muted);
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}
	.control-group input {
		padding: 0.3rem 0.5rem;
		font-size: 0.85rem;
	}
	.calendar-scroll {
		flex: 1;
		overflow: auto;
	}
	.calendar-table {
		border-collapse: collapse;
		min-width: 100%;
	}
	.time-col {
		width: 56px;
		min-width: 56px;
	}
	.day-header {
		min-width: 140px;
		padding: 0.6rem 0.4rem;
		font-size: 0.78rem;
		font-weight: 600;
		text-align: center;
		border-right: 1px solid var(--border);
		border-bottom: 2px solid var(--border);
		white-space: nowrap;
		color: var(--text-muted);
		position: sticky;
		top: 0;
		background: var(--bg-card);
		z-index: 10;
	}
	.day-header.today {
		color: var(--color-primary);
		border-bottom-color: var(--color-primary);
	}
	.time-label {
		font-size: 0.72rem;
		color: var(--text-muted);
		text-align: right;
		padding: 0 8px;
		vertical-align: top;
		white-space: nowrap;
		border-top: 1px solid var(--border);
		height: 60px;
		position: sticky;
		left: 0;
		background: var(--bg);
		z-index: 5;
	}
	.calendar-cell.droppable {
		height: 60px;
		border-right: 1px solid var(--border);
		border-bottom: 1px solid var(--border);
		position: relative;
		cursor: crosshair;
		transition: background 0.1s;
		user-select: none;
	}
	.calendar-cell.droppable:hover {
		background: rgba(74, 144, 217, 0.04);
	}
	.slot-block {
		position: absolute;
		left: 2px;
		right: 2px;
		border-radius: 5px;
		cursor: pointer;
		border: none;
		opacity: 0.85;
		transition:
			opacity 0.15s,
			transform 0.12s;
		overflow: hidden;
		display: flex;
		align-items: flex-start;
		padding: 2px 4px;
		font-size: 0.7rem;
		color: white;
		font-family: inherit;
		z-index: 1;
	}
	.slot-block:hover {
		opacity: 1;
		transform: scaleY(1.02);
		z-index: 2;
	}
	.slot-comment {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.4);
	}
	.drag-preview {
		position: absolute;
		inset: 2px;
		border-radius: 5px;
		pointer-events: none;
		z-index: 0;
	}
	.undo-btn {
		background: var(--bg-card);
		color: var(--text);
		border: 1.5px solid var(--border);
		font-size: 1.2rem;
		box-shadow: var(--shadow-md);
	}
	.undo-btn:hover {
		background: var(--bg-input);
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
	.edit-actions {
		display: flex;
		gap: 0.6rem;
		flex-wrap: wrap;
		margin-top: 1rem;
	}
</style>
