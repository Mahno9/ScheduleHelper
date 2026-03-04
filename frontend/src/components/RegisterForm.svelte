<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { api } from '../api'
  import type { User } from '../api'

  const dispatch = createEventDispatcher<{ registered: User; back: void }>()

  const EMOJIS = [
    '😊','😎','🥳','🤩','🦊','🐱','🐶','🐼','🦁','🐯',
    '🦋','🦄','🐸','🐧','🦅','🐺','🦝','🐨','🦔','🐻',
    '🌟','⭐','🔥','💎','🚀','🎯','🎨','🎭','🎪','🎸',
    '🌙','☀️','🌈','❄️','🌺','🌸','🍀','🌊','⚡','🎋',
    '💜','💙','💚','❤️','🧡','💛','🤍','🖤','💕','🫀',
  ]

  const COLORS = [
    '#4A90D9','#E74C3C','#2ECC71','#9B59B6','#F39C12',
    '#1ABC9C','#E91E63','#3F51B5','#00BCD4','#8BC34A',
    '#FF5722','#607D8B','#795548','#FFC107','#9C27B0',
    '#03A9F4','#4CAF50','#FF9800','#F44336','#673AB7',
  ]

  let name = ''
  let selectedEmoji = EMOJIS[0]
  let selectedColor = COLORS[0]
  let loading = false
  let error = ''
  let showEmojiPicker = false
  let tz = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  let autoTZ = true

  function pickEmoji(e: string) {
    selectedEmoji = e
    showEmojiPicker = false
  }

  async function submit() {
    error = ''
    const n = name.trim()
    if (!n) { error = 'Введите имя'; return }
    if (n.length > 40) { error = 'Имя слишком длинное'; return }
    loading = true
    try {
      const user = await api.register({ name: n, emoji: selectedEmoji, color: selectedColor, timezone: tz, auto_timezone: autoTZ })
      dispatch('registered', user)
    } catch (e: any) {
      error = e.message === 'name already taken' ? 'Имя уже занято' : (e.message || 'Ошибка регистрации')
    } finally {
      loading = false
    }
  }
</script>

<div class="modal-overlay" on:click|self={() => dispatch('back')}>
  <div class="modal">
    <button class="back-btn" on:click={() => dispatch('back')}>← Назад</button>
    <h2>Создать профиль</h2>

    <!-- Preview -->
    <div class="preview" style="--color: {selectedColor}">
      <div class="avatar-preview" style="background: radial-gradient(circle at 35% 35%, {selectedColor}66, {selectedColor}, {selectedColor}cc)">
        <span>{selectedEmoji}</span>
      </div>
      <span class="name-preview" style="color: {selectedColor}">{name || 'Ваше имя'}</span>
    </div>

    <!-- Emoji picker -->
    <div class="field">
      <label class="label">Аватар</label>
      <button class="emoji-trigger" on:click={() => showEmojiPicker = !showEmojiPicker}>
        <span class="big-emoji">{selectedEmoji}</span>
        <span>Выбрать эмоджи</span>
      </button>
      {#if showEmojiPicker}
        <div class="emoji-grid" role="grid">
          {#each EMOJIS as e}
            <button
              class="emoji-btn"
              class:active={e === selectedEmoji}
              on:click={() => pickEmoji(e)}
              title={e}
            >{e}</button>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Color picker -->
    <div class="field">
      <label class="label">Цвет</label>
      <div class="color-grid">
        {#each COLORS as c}
          <button
            class="color-dot"
            class:selected={c === selectedColor}
            style="background: {c}"
            on:click={() => selectedColor = c}
            title={c}
          ></button>
        {/each}
      </div>
    </div>

    <!-- Name -->
    <div class="field">
      <label class="label" for="reg-name">Имя пользователя</label>
      <input id="reg-name" type="text" bind:value={name} placeholder="Уникальное имя" maxlength="40" />
      {#if error}<p class="error-msg">{error}</p>{/if}
    </div>

    <button class="btn btn-primary" style="width:100%; justify-content: center; margin-top: 1rem" on:click={submit} disabled={loading}>
      {#if loading}<div class="spinner"></div>{:else}Зарегистрироваться{/if}
    </button>
  </div>
</div>

<style>
  h2 { margin-bottom: 1.5rem; font-size: 1.3rem; }
  .back-btn {
    background: none;
    color: var(--text-muted);
    font-size: 0.9rem;
    padding: 0;
    margin-bottom: 1rem;
    cursor: pointer;
  }
  .back-btn:hover { color: var(--text); }
  .preview {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    padding: 1rem;
    background: var(--bg-input);
    border-radius: var(--radius-sm);
    margin-bottom: 1.5rem;
  }
  .avatar-preview {
    width: 56px; height: 56px;
    border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    box-shadow: 0 2px 8px rgba(0,0,0,0.2);
    font-size: 1.8rem;
    flex-shrink: 0;
  }
  .name-preview {
    font-size: 1.1rem;
    font-weight: 600;
    text-shadow: -1px -1px 0 rgba(0,0,0,0.4), 1px 1px 0 rgba(0,0,0,0.4);
    word-break: break-word;
  }
  .field { margin-bottom: 1.2rem; }
  .emoji-trigger {
    display: flex; align-items: center; gap: 0.6rem;
    background: var(--bg-input);
    border: 1.5px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 0.5rem 0.8rem;
    font-size: 0.9rem;
    color: var(--text);
    width: 100%;
    cursor: pointer;
    transition: border-color 0.2s;
  }
  .emoji-trigger:hover { border-color: var(--color-primary); }
  .big-emoji { font-size: 1.4rem; line-height: 1; }
  .emoji-grid {
    display: grid;
    grid-template-columns: repeat(10, 1fr);
    gap: 4px;
    margin-top: 0.5rem;
    padding: 0.6rem;
    background: var(--bg-input);
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    max-height: 160px;
    overflow-y: auto;
  }
  .emoji-btn {
    font-size: 1.3rem;
    padding: 4px;
    border-radius: 6px;
    background: none;
    transition: background 0.12s, transform 0.12s;
    cursor: pointer;
    line-height: 1;
  }
  .emoji-btn:hover { background: var(--bg-card); transform: scale(1.15); }
  .emoji-btn.active { background: var(--color-primary); }
  .color-grid {
    display: flex; flex-wrap: wrap; gap: 8px;
  }
  .color-dot {
    width: 28px; height: 28px;
    border-radius: 50%;
    border: 2.5px solid transparent;
    cursor: pointer;
    transition: transform 0.15s, border-color 0.15s;
  }
  .color-dot:hover { transform: scale(1.15); }
  .color-dot.selected { border-color: white; box-shadow: 0 0 0 2px rgba(0,0,0,0.3); transform: scale(1.1); }
</style>
