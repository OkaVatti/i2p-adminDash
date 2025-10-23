<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { get } from 'svelte/store';
	import { token, theme } from '$lib/stores';
	import { goto } from '$app/navigation';

	let ws: WebSocket | null = null;
	let logs = '';
	let currentTheme = 'Default';
	let isConnected = false;
	let reconnectTimer: number | null = null;

	function connectWS() {
		const t = get(token);
		if (!t) {
			logs += '[ERROR] No token found. Redirecting to login...\n';
			setTimeout(() => goto('/'), 2000);
			return;
		}

		const base = (import.meta.env.VITE_WS_BASE as string) || 'ws://127.0.0.1:8080';
		const url = `${base}/ws?token=${encodeURIComponent(t)}`;

		try {
			ws = new WebSocket(url);

			ws.onopen = () => {
				logs += `[${new Date().toISOString()}] WebSocket connection established\n`;
				isConnected = true;
				if (reconnectTimer) {
					clearTimeout(reconnectTimer);
					reconnectTimer = null;
				}
			};

			ws.onmessage = (ev: MessageEvent) => {
				try {
					const data = JSON.parse(ev.data);
					logs += `[${new Date().toISOString()}] Received data:\n${JSON.stringify(data, null, 2)}\n\n`;
				} catch (e) {
					logs += `[${new Date().toISOString()}] ${ev.data}\n`;
				}
			};

			ws.onerror = (error) => {
				logs += `[${new Date().toISOString()}] WebSocket error occurred\n`;
				isConnected = false;
			};

			ws.onclose = () => {
				logs += `[${new Date().toISOString()}] WebSocket connection closed\n`;
				isConnected = false;

				// Attempt to reconnect after 5 seconds
				reconnectTimer = setTimeout(() => {
					logs += '[INFO] Attempting to reconnect...\n';
					connectWS();
				}, 5000) as unknown as number;
			};
		} catch (error) {
			logs += `[ERROR] Failed to create WebSocket: ${error}\n`;
		}
	}

	function changeTheme(arg: Event | string) {
		let val: string | undefined;
		if (typeof arg === 'string') {
			val = arg;
		} else {
			const t = arg as Event & { target: HTMLSelectElement };
			val = t.target?.value;
		}
		if (!val) return;
		currentTheme = val;
		theme.set(val);
	}

	function clearLogs() {
		logs = '';
	}

	function logout() {
		if (ws) {
			ws.close();
			ws = null;
		}
		token.set(null);
		goto('/');
	}

	onMount(() => {
		// Check if user is authenticated
		const t = get(token);
		if (!t) {
			goto('/');
			return;
		}

		connectWS();
		const unsub = theme.subscribe((v) => (currentTheme = v));

		return () => {
			unsub();
		};
	});

	onDestroy(() => {
		if (ws) {
			ws.close();
			ws = null;
		}
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
		}
	});
</script>

<div class="container">
	<div class="header">
		<div class="header-left">
			<h1>I2P Admin Dashboard</h1>
			<div
				class="connection-status"
				class:connected={isConnected}
				class:disconnected={!isConnected}
			>
				<span class="status-dot" class:connected={isConnected} class:disconnected={!isConnected}
				></span>
				{isConnected ? 'Connected' : 'Disconnected'}
			</div>
		</div>

		<div class="header-right">
			<div class="theme-selector">
				<label for="theme-select">Theme:</label>
				<select id="theme-select" on:change={changeTheme} bind:value={currentTheme}>
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

			<button class="btn-logout" on:click={logout}> Logout </button>
		</div>
	</div>

	<div class="dashboard-grid">
		<div class="panel">
			<div class="panel-header">
				<h2>Router Stats & Logs</h2>
				<button class="btn-clear" on:click={clearLogs}>Clear</button>
			</div>
			<div class="logs-container">
				{logs || 'Waiting for data...'}
			</div>
		</div>

		<div class="panel">
			<h2>Control Panel</h2>
			<div class="control-buttons">
				<button class="btn-control" on:click={() => alert('Feature: List I2PSnark torrents')}>
					📊 I2PSnark: List Torrents
				</button>
				<button class="btn-control" on:click={() => alert('Feature: Start torrent')}>
					▶️ I2PSnark: Start Torrent
				</button>
				<button class="btn-control" on:click={() => alert('Feature: Stop torrent')}>
					⏸️ I2PSnark: Stop Torrent
				</button>
				<button class="btn-control" on:click={() => alert('Feature: View tunnel stats')}>
					🚇 View Tunnel Statistics
				</button>
				<button class="btn-control" on:click={() => alert('Feature: View hidden service stats')}>
					🔒 View Hidden Service Stats
				</button>
			</div>
		</div>
	</div>
</div>

<style>
	.container {
		max-width: 80rem;
		margin: 0 auto;
		padding: 1rem;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		border-bottom: 1px solid #e5e7eb;
		margin-bottom: 1.5rem;
	}

	.header-left h1 {
		font-size: 1.5rem;
		font-weight: bold;
		margin-bottom: 0.5rem;
	}

	.connection-status {
		display: inline-flex;
		align-items: center;
		font-size: 0.875rem;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
	}

	.connection-status.connected {
		background-color: #d1fae5;
		color: #065f46;
	}

	.connection-status.disconnected {
		background-color: #fee2e2;
		color: #991b1b;
	}

	.status-dot {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		margin-right: 0.5rem;
	}

	.status-dot.connected {
		background-color: #10b981;
	}

	.status-dot.disconnected {
		background-color: #ef4444;
	}

	.header-right {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.theme-selector {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.theme-selector label {
		font-weight: 500;
	}

	.theme-selector select {
		padding: 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 0.25rem;
	}

	.btn-logout {
		padding: 0.5rem 1rem;
		background-color: #ef4444;
		color: white;
		border: none;
		border-radius: 0.25rem;
		cursor: pointer;
		font-weight: 500;
	}

	.btn-logout:hover {
		background-color: #dc2626;
	}

	.dashboard-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1.5rem;
		margin-top: 1.5rem;
	}

	.panel {
		padding: 1.5rem;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		background-color: white;
	}

	.panel-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.panel h2 {
		font-size: 1.25rem;
		font-weight: 600;
	}

	.btn-clear {
		padding: 0.25rem 0.75rem;
		background-color: #6b7280;
		color: white;
		border: none;
		border-radius: 0.25rem;
		cursor: pointer;
		font-size: 0.875rem;
	}

	.btn-clear:hover {
		background-color: #4b5563;
	}

	.logs-container {
		max-height: 400px;
		overflow-y: auto;
		background-color: #1f2937;
		color: #f9fafb;
		padding: 1rem;
		border-radius: 0.25rem;
		font-family: 'Courier New', monospace;
		font-size: 0.875rem;
		white-space: pre-wrap;
		word-break: break-all;
	}

	.control-buttons {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.btn-control {
		padding: 0.75rem;
		background-color: #3b82f6;
		color: white;
		border: none;
		border-radius: 0.25rem;
		cursor: pointer;
		font-weight: 500;
		text-align: left;
	}

	.btn-control:hover {
		background-color: #2563eb;
	}

	@media (max-width: 768px) {
		.dashboard-grid {
			grid-template-columns: 1fr;
		}

		.header {
			flex-direction: column;
			gap: 1rem;
		}

		.header-right {
			width: 100%;
			flex-direction: column;
		}
	}
</style>
