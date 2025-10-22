<script lang="ts">
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';
  import { token, theme } from '$lib/stores';

  let ws: WebSocket | null = null;
  let logs = '';
  let currentTheme = 'Default';

  function connectWS() {
    const t = get(token);
    if (!t) {
      logs += 'No token, please login\n';
      return;
    }
    const base = (import.meta.env.VITE_WS_BASE as string) || 'ws://127.0.0.1:8080';
    const url = `${base}/ws?token=${encodeURIComponent(t)}`;
    ws = new WebSocket(url);
    ws.onopen = () => (logs += 'ws open\n');
    ws.onmessage = (ev: MessageEvent) => {
      try {
        const data = JSON.parse(ev.data);
        logs += JSON.stringify(data, null, 2) + '\n';
      } catch (e) {
        logs += ev.data + '\n';
      }
    };
    ws.onclose = () => (logs += 'ws closed\n');
  }

  // Accept Event or string
  function changeTheme(arg: Event | string) {
    let val: string | undefined;
    if (typeof arg === 'string') val = arg;
    else {
      const t = arg as Event & { target: HTMLSelectElement };
      val = t.target?.value;
    }
    if (!val) return;
    currentTheme = val;
    theme.set(val);
  }

  onMount(() => {
    connectWS();
    const unsub = theme.subscribe((v) => (currentTheme = v));
    return () => unsub();
  });
</script>

<div class="container">
  <div class="header">
    <h1 class="text-2xl">Admin Dashboard</h1>
    <div class="mt-2">
      Theme:
      <select on:change={changeTheme} bind:value={currentTheme} class="border p-2">
        <option>Default</option>
        <option>Light</option>
        <option>Dark</option>
        <option>LightPlus</option>
        <option>DarkPlus</option>
        <option>Midnight</option>
        <option>OLED</option>
        <option>Platinum</option>
      </select>
    </div>
  </div>

  <div class="mt-6 grid grid-cols-2 gap-4">
    <div class="p-4 border">
      <h2 class="font-bold">Router Stats</h2>
      <pre>{logs}</pre>
    </div>
    <div class="p-4 border">
      <h2 class="font-bold">Control</h2>
      <button class="px-3 py-2 bg-gray-700 text-white rounded" on:click={() => alert('Stub: i2psnark controls')}>I2PSnark: List</button>
    </div>
  </div>
</div>
