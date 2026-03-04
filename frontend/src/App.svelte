<script lang="ts">
	import { onMount } from "svelte";
	import {
		currentUser,
		theme,
		applyTheme,
		connectSSE,
		loadUsers,
	} from "./stores";
	import { api } from "./api";
	import LoginScreen from "./components/LoginScreen.svelte";
	import Header from "./components/Header.svelte";
	import PersonalCalendar from "./components/PersonalCalendar.svelte";
	import SharedCalendar from "./components/SharedCalendar.svelte";
	import { activeTab } from "./stores";
	import type { User } from "./api";

	onMount(async () => {
		// Restore session from localStorage
		const savedUserId = localStorage.getItem("userId");
		if (savedUserId) {
			try {
				const user = await api.getProfile(Number(savedUserId));
				currentUser.set(user);
				theme.set(user.theme || "system");
				applyTheme(user.theme || "system");
			} catch {
				localStorage.removeItem("userId");
			}
		}

		// Apply initial theme (before login)
		if (!$currentUser) {
			applyTheme("system");
		}

		await loadUsers();
		connectSSE();

		// Listen for system theme changes
		window
			.matchMedia("(prefers-color-scheme: dark)")
			.addEventListener("change", () => {
				if ($theme === "system") applyTheme("system");
			});
	});

	function handleLogin(e: CustomEvent<User>) {
		const user = e.detail;
		currentUser.set(user);
		theme.set(user.theme || "system");
		applyTheme(user.theme || "system");
		localStorage.setItem("userId", String(user.id));
	}

	function handleLogout() {
		currentUser.set(null);
		localStorage.removeItem("userId");
		applyTheme("system");
	}

	function handleProfileUpdated(e: CustomEvent<User>) {
		const user = e.detail;
		currentUser.set(user);
		theme.set(user.theme || "system");
		applyTheme(user.theme || "system");
	}
</script>

{#if !$currentUser}
	<LoginScreen on:login={handleLogin} />
{:else}
	<div class="app-shell">
		<Header
			on:logout={handleLogout}
			on:profileUpdated={handleProfileUpdated}
		/>
		<main class="main-content">
			{#if $activeTab === "shared"}
				<SharedCalendar />
			{:else}
				<PersonalCalendar />
			{/if}
		</main>
	</div>
{/if}

<style>
	.app-shell {
		display: flex;
		flex-direction: column;
		height: 100vh;
		overflow: hidden;
	}
	.main-content {
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}
</style>
