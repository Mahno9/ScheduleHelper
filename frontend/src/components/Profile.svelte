<script lang="ts">
  import type { User } from '../api'

  export let user: User
  export let size: 'sm' | 'md' | 'lg' = 'md'
  export let showName: boolean = true
  export let clickable: boolean = false

  const sizeMap = { sm: 36, md: 48, lg: 64 }
  const fontMap = { sm: '1.2rem', md: '1.6rem', lg: '2rem' }
  const nameSizeMap = { sm: '0.85rem', md: '0.95rem', lg: '1.1rem' }

  $: px = sizeMap[size]
  $: emojiSize = fontMap[size]
  $: nameSize = nameSizeMap[size]

  // Gradient using the user color
  $: gradient = `radial-gradient(circle at 35% 35%, ${lighten(user.color, 40)}, ${user.color} 70%, ${darken(user.color, 20)})`

  // Compute contrast outline color based on color brightness
  $: outline = getBrightness(user.color) > 128 ? 'rgba(0,0,0,0.5)' : 'rgba(255,255,255,0.8)'

  function hexToRgb(hex: string) {
    const h = hex.replace('#', '')
    const n = parseInt(h.length === 3 ? h.split('').map(c => c + c).join('') : h, 16)
    return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 }
  }

  function getBrightness(hex: string) {
    try {
      const { r, g, b } = hexToRgb(hex)
      return (r * 299 + g * 587 + b * 114) / 1000
    } catch { return 128 }
  }

  function lighten(hex: string, amt: number) {
    try {
      const { r, g, b } = hexToRgb(hex)
      return `rgb(${Math.min(255, r + amt)}, ${Math.min(255, g + amt)}, ${Math.min(255, b + amt)})`
    } catch { return hex }
  }

  function darken(hex: string, amt: number) {
    try {
      const { r, g, b } = hexToRgb(hex)
      return `rgb(${Math.max(0, r - amt)}, ${Math.max(0, g - amt)}, ${Math.max(0, b - amt)})`
    } catch { return hex }
  }
</script>

<div
  class="profile"
  class:clickable
  style="--px: {px}px; --emoji-size: {emojiSize}; --name-size: {nameSize}; --color: {user.color}; --outline: {outline}"
  on:click
  on:keydown={(e) => e.key === 'Enter' && clickable}
  role={clickable ? 'button' : undefined}
  tabindex={clickable ? 0 : undefined}
>
  <div class="avatar" style="background: {gradient}">
    <span class="emoji" aria-label="avatar">{user.emoji}</span>
  </div>
  {#if showName}
    <span class="name contrast-text" style="color: {user.color}; --outline: {outline}">
      {user.name}
    </span>
  {/if}
</div>

<style>
  .profile {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    user-select: none;
  }
  .profile.clickable {
    cursor: pointer;
    border-radius: var(--radius-sm);
    padding: 0.4rem 0.6rem;
    transition: background 0.15s;
  }
  .profile.clickable:hover {
    background: var(--bg-input);
  }
  .profile.clickable:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }
  .avatar {
    width: var(--px);
    height: var(--px);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    box-shadow: 0 2px 8px rgba(0,0,0,0.2);
    transition: transform 0.15s;
  }
  .profile.clickable:hover .avatar {
    transform: scale(1.05);
  }
  .emoji {
    font-size: var(--emoji-size);
    line-height: 1;
  }
  .name {
    font-size: var(--name-size);
    font-weight: 600;
    letter-spacing: -0.01em;
  }
</style>
