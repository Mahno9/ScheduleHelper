<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { activeTab, currentUser } from "../stores";
	import Profile from "./Profile.svelte";
	import ProfileSettings from "./ProfileSettings.svelte";
	import type { User } from "../api";

	const dispatch = createEventDispatcher<{
		logout: void;
		profileUpdated: User;
	}>();

	let showSettings = false;

	function handleUpdated(e: CustomEvent<User>) {
		currentUser.set(e.detail);
		dispatch("profileUpdated", e.detail);
		showSettings = false;
	}

	function handleDeleted() {
		currentUser.set(null);
		showSettings = false;
		dispatch("logout");
	}
</script>

<header class="header">
	<div class="header-inner">
		<div class="logo">📅 ScheduleHelper</div>

		<nav class="tabs" role="tablist">
			<button
				class="tab"
				class:active={$activeTab === "shared"}
				role="tab"
				aria-selected={$activeTab === "shared"}
				on:click={() => activeTab.set("shared")}
				id="tab-shared"
			>
				Общий
			</button>
			<button
				class="tab"
				class:active={$activeTab === "personal"}
				role="tab"
				aria-selected={$activeTab === "personal"}
				on:click={() => activeTab.set("personal")}
				id="tab-personal"
			>
				Личный
			</button>
		</nav>

		<div class="user-area">
			{#if $currentUser}
				<button
					class="user-btn"
					on:click={() => (showSettings = !showSettings)}
					id="profile-btn"
					aria-label="Профиль"
				>
					<Profile user={$currentUser} size="sm" showName={true} />
				</button>

				{#if showSettings}
					<!-- Floating settings panel -->
					<div
						class="settings-panel card"
						role="dialog"
						aria-label="Настройки профиля"
					>
						<ProfileSettings
							user={$currentUser}
							on:updated={handleUpdated}
							on:deleted={handleDeleted}
						/>
					</div>
					<!-- Click outside to close -->
					<div
						class="settings-backdrop"
						on:click={() => (showSettings = false)}
					></div>
				{/if}
			{/if}
		</div>
	</div>
</header>

<style>
	.header {
		position: sticky;
		top: 0;
		z-index: 90;
		background: var(--header-bg);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		border-bottom: 1px solid var(--border);
		box-shadow: 0 1px 12px rgba(0, 0, 0, 0.06);
	}
	.header-inner {
		max-width: 1400px;
		margin: 0 auto;
		padding: 0 1.5rem;
		height: 60px;
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}
	.logo {
		font-weight: 700;
		font-size: 1.1rem;
		white-space: nowrap;
		flex-shrink: 0;
		background: linear-gradient(135deg, var(--color-primary), #a78bfa);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}
	.tabs {
		display: flex;
		gap: 0.2rem;
		flex: 1;
		justify-content: center;
	}
	.tab {
		padding: 0.4rem 1.2rem;
		border-radius: 20px;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text-muted);
		background: none;
		transition: all 0.18s ease;
		cursor: pointer;
		border: none;
	}
	.tab.active {
		background: var(--color-primary);
		color: white;
	}
	.tab:hover:not(.active) {
		background: var(--bg-input);
		color: var(--text);
	}
	.user-area {
		position: relative;
		flex-shrink: 0;
	}
	.user-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.settings-panel {
		position: absolute;
		top: calc(100% + 12px);
		right: 0;
		width: 320px;
		max-height: 80vh;
		overflow-y: auto;
		z-index: 200;
		animation: slideDown 0.2s ease;
	}
	@keyframes slideDown {
		from {
			transform: translateY(-8px);
			opacity: 0;
		}
		to {
			transform: translateY(0);
			opacity: 1;
		}
	}
	.settings-backdrop {
		position: fixed;
		inset: 0;
		z-index: 199;
	}
</style>
