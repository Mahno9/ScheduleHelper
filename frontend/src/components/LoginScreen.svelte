<script lang="ts">
	import { createEventDispatcher, onMount } from "svelte";
	import { users, loadUsers } from "../stores";
	import Profile from "./Profile.svelte";
	import RegisterForm from "./RegisterForm.svelte";
	import type { User } from "../api";

	const dispatch = createEventDispatcher<{ login: User }>();

	let showRegister = false;

	onMount(() => {
		loadUsers();
	});

	async function login(user: User) {
		dispatch("login", user);
	}

	function handleRegistered(e: CustomEvent<User>) {
		showRegister = false;
		loadUsers();
		login(e.detail);
	}
</script>

<main class="login-page">
	<div class="hero">
		<h1>📅 ScheduleHelper</h1>
		<p class="subtitle">Найдите общее время для встреч с командой</p>
	</div>

	{#if $users.length > 0}
		<div class="users-section">
			<h2>Выберите свой профиль</h2>
			<div class="users-grid">
				{#each $users as user (user.id)}
					<button
						class="user-card"
						on:click={() => login(user)}
						on:keydown={(e) => e.key === "Enter" && login(user)}
					>
						<Profile {user} size="md" showName={true} />
					</button>
				{/each}
			</div>
		</div>
	{:else}
		<div class="empty-state">
			<span class="empty-icon">👥</span>
			<p>Пока нет ни одного пользователя</p>
		</div>
	{/if}

	<div class="register-section">
		<button
			class="btn btn-primary"
			on:click={() => (showRegister = true)}
			id="register-btn"
		>
			+ Добавить профиль
		</button>
	</div>

	{#if showRegister}
		<RegisterForm
			on:registered={handleRegistered}
			on:back={() => (showRegister = false)}
		/>
	{/if}
</main>

<style>
	.login-page {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 3rem 1.5rem;
	}
	.hero {
		text-align: center;
		margin-bottom: 3rem;
	}
	h1 {
		font-size: 2.5rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
		background: linear-gradient(135deg, var(--color-primary), #a78bfa);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}
	.subtitle {
		color: var(--text-muted);
		font-size: 1.1rem;
	}
	.users-section {
		width: 100%;
		max-width: 720px;
		margin-bottom: 2rem;
	}
	h2 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		margin-bottom: 1rem;
	}
	.users-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}
	.user-card {
		background: var(--bg-card);
		border: 1.5px solid var(--border);
		border-radius: var(--radius);
		padding: 0.8rem 1.2rem;
		cursor: pointer;
		transition: all 0.18s ease;
		box-shadow: var(--shadow);
	}
	.user-card:hover {
		border-color: var(--color-primary);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}
	.user-card:focus-visible {
		outline: 2px solid var(--color-primary);
		outline-offset: 2px;
	}
	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-muted);
		margin-bottom: 2rem;
	}
	.empty-icon {
		font-size: 3rem;
		display: block;
		margin-bottom: 0.8rem;
	}
	.register-section {
		margin-top: 1rem;
	}
</style>
