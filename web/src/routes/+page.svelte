<script lang="ts">
  import { onMount } from 'svelte';
  import { sha3_512 } from 'js-sha3';
  import { login, setup } from '$lib/api';
  import { token } from '$lib/stores';
  import { goto } from '$app/navigation';

  let username = '';
  let password = '';
  let email = '';
  let message = '';

  function utf8Hex(input: string): string {
    const enc = new TextEncoder().encode(input);
    return sha3_512(enc);
  }

  async function doHashAndLogin() {
    message = '';
    const hex = utf8Hex(password);
    try {
      const res = await login(username, hex);
      if (res.token) {
        token.set(res.token);
        message = 'Login success';
        // redirect to dashboard
        await goto('/dashboard');
        return;
      }
      message = res.error || JSON.stringify(res);
    } catch (err) {
      message = String(err);
    }
  }

  async function doSetup() {
    message = '';
    if (password.length < 32) {
      message = 'Passphrase must be at least 32 characters for setup';
      return;
    }
    const hex = utf8Hex(password);
    try {
      const res = await setup(username, email, hex);
      message = JSON.stringify(res);
    } catch (err) {
      message = String(err);
    }
  }
</script>

<style>
  .container { max-width: 48rem; margin: 0 auto; padding: 1rem; }
</style>

<div class="container">
  <div class="header"><h1 class="text-2xl">I2P Admin Dashboard — Login / Setup</h1></div>

  <div class="mt-6">
    <label>Username</label>
    <input type="text" bind:value={username} class="border p-2 w-full" />
    <label class="mt-2">Password / Passphrase (min 32 chars for setup)</label>
    <input type="password" bind:value={password} class="border p-2 w-full" />
    <label class="mt-2">I2P Email (for setup)</label>
    <input type="email" bind:value={email} class="border p-2 w-full" />
    <div class="mt-4 flex gap-2">
      <button on:click={doHashAndLogin} class="px-4 py-2 bg-blue-600 text-white rounded">Login</button>
      <button on:click={doSetup} class="px-4 py-2 bg-green-600 text-white rounded">Initial Setup</button>
    </div>
    <p class="mt-4 text-sm text-red-600">{message}</p>
  </div>
</div>
