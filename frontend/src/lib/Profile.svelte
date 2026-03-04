<script lang="ts">
	import type { User } from '../store';

	export let user: User;
	export let size: 'small' | 'large' = 'small';
	export let onClick: (() => void) | undefined = undefined;

	$: circleSize = size === 'small' ? '40px' : '80px';
	$: fontSize = size === 'small' ? '20px' : '40px';
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="profile-container" class:clickable={!!onClick} on:click={onClick}>
	<div 
		class="avatar-circle" 
		style="width: {circleSize}; height: {circleSize}; font-size: {fontSize}; background: linear-gradient(135deg, {user.color}88, {user.color});"
	>
		{user.emoji}
	</div>
	<span class="username" style="color: {user.color};">{user.username}</span>
</div>

<style>
	.profile-container {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px;
		border-radius: 50px;
		transition: transform 0.2s, background 0.2s;
	}
	.clickable {
		cursor: pointer;
	}
	.clickable:hover {
		transform: scale(1.02);
		background: rgba(0, 0, 0, 0.05);
	}
	.avatar-circle {
		display: flex;
		justify-content: center;
		align-items: center;
		border-radius: 50%;
		box-shadow: 0 4px 10px rgba(0,0,0,0.1);
	}
	.username {
		font-weight: 600;
		font-size: 1.1em;
		text-shadow: 
			-1px -1px 0 #fff,  
			 1px -1px 0 #fff,
			-1px  1px 0 #fff,
			 1px  1px 0 #fff;
	}
	@media (prefers-color-scheme: dark) {
		.username {
			text-shadow: 
				-1px -1px 0 #000,  
				 1px -1px 0 #000,
				-1px  1px 0 #000,
				 1px  1px 0 #000;
		}
		.clickable:hover {
			background: rgba(255, 255, 255, 0.05);
		}
	}
</style>