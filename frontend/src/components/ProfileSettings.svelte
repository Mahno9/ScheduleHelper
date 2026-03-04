<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { currentUser, theme, applyTheme } from "../stores";
	import { api } from "../api";
	import type { User } from "../api";

	const dispatch = createEventDispatcher<{ updated: User; deleted: void }>();

	export let user: User;

	const EMOJIS = [
		"😊",
		"😎",
		"🥳",
		"🤩",
		"🦊",
		"🐱",
		"🐶",
		"🐼",
		"🦁",
		"🐯",
		"🦋",
		"🦄",
		"🐸",
		"🐧",
		"🦅",
		"🐺",
		"🦝",
		"🐨",
		"🦔",
		"🐻",
		"🌟",
		"⭐",
		"🔥",
		"💎",
		"🚀",
		"🎯",
		"🎨",
		"🎭",
		"🎪",
		"🎸",
		"🌙",
		"☀️",
		"🌈",
		"❄️",
		"🌺",
		"🌸",
		"🍀",
		"🌊",
		"⚡",
		"🎋",
		"💜",
		"💙",
		"💚",
		"❤️",
		"🧡",
		"💛",
		"🤍",
		"🖤",
		"💕",
		"🫀",
	];

	const COLORS = [
		"#4A90D9",
		"#E74C3C",
		"#2ECC71",
		"#9B59B6",
		"#F39C12",
		"#1ABC9C",
		"#E91E63",
		"#3F51B5",
		"#00BCD4",
		"#8BC34A",
		"#FF5722",
		"#607D8B",
		"#795548",
		"#FFC107",
		"#9C27B0",
		"#03A9F4",
		"#4CAF50",
		"#FF9800",
		"#F44336",
		"#673AB7",
	];

	let name = user.name;
	let selectedEmoji = user.emoji;
	let selectedColor = user.color;
	let selectedTheme = user.theme || "system";
	let timezone =
		user.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone;
	let autoTZ = user.auto_timezone;
	let loading = false;
	let error = "";
	let showEmojiPicker = false;
	let confirmDelete = false;

	async function save() {
		error = "";
		const n = name.trim();
		if (!n) {
			error = "Имя не может быть пустым";
			return;
		}
		loading = true;
		try {
			const updated = await api.updateProfile(user.id, {
				name: n,
				emoji: selectedEmoji,
				color: selectedColor,
				theme: selectedTheme,
				timezone,
				auto_timezone: autoTZ,
			});
			theme.set(selectedTheme);
			applyTheme(selectedTheme);
			dispatch("updated", updated);
		} catch (e: any) {
			error = e.message || "Ошибка сохранения";
		} finally {
			loading = false;
		}
	}

	async function deleteAccount() {
		loading = true;
		try {
			await api.deleteProfile(user.id);
			dispatch("deleted");
		} catch (e: any) {
			error = e.message || "Ошибка удаления";
		} finally {
			loading = false;
			confirmDelete = false;
		}
	}
</script>

<div class="settings">
	<h3>Настройки профиля</h3>

	<!-- Avatar -->
	<div class="field">
		<label class="label">Аватар</label>
		<button
			class="emoji-trigger"
			on:click={() => (showEmojiPicker = !showEmojiPicker)}
		>
			<span style="font-size: 1.3rem">{selectedEmoji}</span> Изменить
		</button>
		{#if showEmojiPicker}
			<div class="emoji-grid">
				{#each EMOJIS as e}
					<button
						class="emoji-btn"
						class:active={e === selectedEmoji}
						on:click={() => {
							selectedEmoji = e;
							showEmojiPicker = false;
						}}>{e}</button
					>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Color -->
	<div class="field">
		<label class="label">Цвет</label>
		<div class="color-row">
			{#each COLORS as c}
				<button
					class="color-dot"
					class:selected={c === selectedColor}
					style="background:{c}"
					on:click={() => (selectedColor = c)}
				></button>
			{/each}
		</div>
	</div>

	<!-- Name -->
	<div class="field">
		<label class="label" for="settings-name">Имя</label>
		<input
			id="settings-name"
			type="text"
			bind:value={name}
			maxlength="40"
		/>
	</div>

	<!-- Timezone -->
	<div class="field">
		<label class="label">Часовой пояс</label>
		<label class="checkbox-row">
			<input type="checkbox" bind:checked={autoTZ} />
			<span>Определять автоматически</span>
		</label>
		{#if !autoTZ}
			<input
				type="text"
				bind:value={timezone}
				placeholder="Europe/Moscow"
				style="margin-top: 0.4rem"
			/>
		{/if}
	</div>

	<!-- Theme -->
	<div class="field">
		<label class="label">Тема оформления</label>
		<div class="theme-tabs">
			{#each [["light", "☀️ Светлая"], ["dark", "🌙 Тёмная"], ["system", "💻 Системная"]] as [val, label]}
				<button
					class="theme-btn"
					class:active={selectedTheme === val}
					on:click={() => (selectedTheme = val)}>{label}</button
				>
			{/each}
		</div>
	</div>

	{#if error}<p class="error-msg">{error}</p>{/if}

	<button
		class="btn btn-primary"
		style="width:100%;justify-content:center;margin-bottom:0.8rem"
		on:click={save}
		disabled={loading}
	>
		{loading ? "..." : "Сохранить"}
	</button>

	{#if confirmDelete}
		<div class="confirm-delete">
			<p>Удалить профиль? Это невозможно отменить.</p>
			<div style="display:flex;gap:0.5rem;margin-top:0.5rem">
				<button
					class="btn btn-danger"
					on:click={deleteAccount}
					disabled={loading}>Удалить</button
				>
				<button
					class="btn btn-ghost"
					on:click={() => (confirmDelete = false)}>Отмена</button
				>
			</div>
		</div>
	{:else}
		<button
			class="btn btn-danger"
			style="width:100%;justify-content:center"
			on:click={() => (confirmDelete = true)}
		>
			🗑 Удалить профиль
		</button>
	{/if}
</div>

<style>
	.settings {
		padding: 0.5rem 0;
	}
	h3 {
		font-size: 1rem;
		font-weight: 600;
		margin-bottom: 1.2rem;
	}
	.field {
		margin-bottom: 1rem;
	}
	.emoji-trigger {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: var(--bg-input);
		border: 1.5px solid var(--border);
		border-radius: var(--radius-sm);
		padding: 0.45rem 0.8rem;
		font-size: 0.9rem;
		color: var(--text);
		cursor: pointer;
		width: 100%;
	}
	.emoji-trigger:hover {
		border-color: var(--color-primary);
	}
	.emoji-grid {
		display: grid;
		grid-template-columns: repeat(10, 1fr);
		gap: 3px;
		margin-top: 0.4rem;
		padding: 0.5rem;
		background: var(--bg-input);
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
		max-height: 130px;
		overflow-y: auto;
	}
	.emoji-btn {
		font-size: 1.2rem;
		padding: 3px;
		border-radius: 5px;
		background: none;
		cursor: pointer;
		transition:
			background 0.1s,
			transform 0.1s;
	}
	.emoji-btn:hover {
		background: var(--bg-card);
		transform: scale(1.1);
	}
	.emoji-btn.active {
		background: var(--color-primary);
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
		box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.25);
	}
	.checkbox-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.9rem;
		cursor: pointer;
	}
	.checkbox-row input {
		width: auto;
	}
	.theme-tabs {
		display: flex;
		gap: 0.4rem;
	}
	.theme-btn {
		flex: 1;
		padding: 0.4rem;
		font-size: 0.8rem;
		background: var(--bg-input);
		border: 1.5px solid var(--border);
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition: all 0.15s;
		color: var(--text);
	}
	.theme-btn.active {
		background: var(--color-primary);
		color: white;
		border-color: var(--color-primary);
	}
	.confirm-delete {
		background: rgba(239, 68, 68, 0.08);
		border: 1px solid rgba(239, 68, 68, 0.3);
		border-radius: var(--radius-sm);
		padding: 0.8rem;
		font-size: 0.9rem;
		margin-bottom: 0.5rem;
	}
</style>
