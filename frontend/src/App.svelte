<script lang="ts">
  import { currentUser, currentView } from './store';
  import Login from './lib/Login.svelte';
  import Register from './lib/Register.svelte';
  import Header from './lib/Header.svelte';
  import PersonalCalendar from './lib/PersonalCalendar.svelte';
  
  let showRegister = false;
</script>

<main>
  {#if $currentUser}
    <Header />
    <div class="content">
      {#if $currentView === 'personal'}
        <PersonalCalendar />
      {:else}
        <h1>Общий календарь (Coming Soon)</h1>
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