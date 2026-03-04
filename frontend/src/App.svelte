<script lang="ts">
  import { currentUser } from './store';
  import Login from './lib/Login.svelte';
  import Register from './lib/Register.svelte';
  import Header from './lib/Header.svelte';
  
  let showRegister = false;
</script>

<main>
  {#if $currentUser}
    <Header />
    <div class="content">
      <!-- Main Content will go here based on currentView -->
      <h1>Welcome, {$currentUser.username}!</h1>
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
    padding: 20px;
  }
</style>