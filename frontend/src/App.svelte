<script lang="ts">
  import { onMount } from 'svelte';
  import { currentUser, currentView, setupSSE } from './store';
  import Login from './lib/Login.svelte';
  import Register from './lib/Register.svelte';
  import Header from './lib/Header.svelte';
  import PersonalCalendar from './lib/PersonalCalendar.svelte';
  import CommonCalendar from './lib/CommonCalendar.svelte';
  
  let showRegister = false;

  onMount(() => {
    setupSSE();
  });
</script>

<main>
  {#if $currentUser}
    <Header />
    <div class="content">
      {#if $currentView === 'personal'}
        <PersonalCalendar />
      {:else}
        <CommonCalendar />
      {/if}
    </div>
  {:else if showRegister}
    <Register onCancel={() => showRegister = false} />
  {:else}
    <Login onRegister={() => showRegister = true} />
  {/if}
</main>

<style>
  main {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }
  .content {
    flex: 1;
    padding: 0;
    display: flex;
    flex-direction: column;
  }
</style>