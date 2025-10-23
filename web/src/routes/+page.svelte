<script lang="ts">
	import { onMount } from 'svelte';
	import { hashSHA3_512, validatePassword, validateI2PEmail } from '$lib/crypto';
	import { login, setup } from '$lib/api';
	import { token } from '$lib/stores';
	import { goto } from '$app/navigation';

	let username = '';
	let password = '';
	let email = '';
	let message = '';
	let isLoading = false;

	async function doHashAndLogin() {
		message = '';
		isLoading = true;

		try {
			if (!username || !password) {
				message = 'Username and password are required';
				return;
			}

			const hex = hashSHA3_512(password);
			const res = await login(username, hex);

			if (res.token) {
				token.set(res.token);
				message = 'Login successful! Redirecting...';
				setTimeout(async () => {
					await goto('/dashboard');
				}, 500);
				return;
			}

			message = res.error || 'Login failed: ' + JSON.stringify(res);
		} catch (err) {
			message = 'Error: ' + String(err);
		} finally {
			isLoading = false;
		}
	}

	async function doSetup() {
		message = '';
		isLoading = true;

		try {
			if (!username || !password || !email) {
				message = 'All fields are required for setup';
				return;
			}

			// Validate password
			const passwordValidation = validatePassword(password);
			if (!passwordValidation.isValid) {
				message = passwordValidation.error || 'Password validation failed';
				return;
			}

			// Validate I2P email
			if (!validateI2PEmail(email)) {
				message = 'Invalid I2P email format. Must end with .i2p';
				return;
			}

			const hex = hashSHA3_512(password);
			const res = await setup(username, email, hex);

			if (res.status === 'setup_complete') {
				message = 'Setup complete! You can now login.';
				// Clear password field
				password = '';
			} else if (res.error === 'already_setup') {
				message = 'Setup already completed. Please use login.';
			} else {
				message = res.error || 'Setup failed: ' + JSON.stringify(res);
			}
		} catch (err) {
			message = 'Error: ' + String(err);
		} finally {
			isLoading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && !isLoading) {
			doHashAndLogin();
		}
	}
</script>

<div class="container">
	<div class="header">
		<h1 class="text-2xl font-bold">I2P Admin Dashboard — Login / Setup</h1>
	</div>

	<div class="mt-6">
		<div class="form-group">
			<label for="username">Username</label>
			<input
				id="username"
				type="text"
				bind:value={username}
				on:keypress={handleKeyPress}
				placeholder="Enter your username"
				disabled={isLoading}
			/>
		</div>

		<div class="form-group">
			<label for="password">Password / Passphrase</label>
			<input
				id="password"
				type="password"
				bind:value={password}
				on:keypress={handleKeyPress}
				placeholder="Enter your password"
				disabled={isLoading}
			/>
			<div class="password-requirements">
				<p>For setup, password must meet these requirements:</p>
				<ul>
					<li>Minimum 32 characters</li>
					<li>At least one uppercase letter</li>
					<li>At least one lowercase letter</li>
					<li>At least one digit</li>
					<li>At least one special character</li>
				</ul>
			</div>
		</div>

		<div class="form-group">
			<label for="email">I2P Email (for setup only)</label>
			<input
				id="email"
				type="email"
				bind:value={email}
				on:keypress={handleKeyPress}
				placeholder="username@mail.i2p"
				disabled={isLoading}
			/>
		</div>

		<div class="button-group">
			<button class="btn-login" on:click={doHashAndLogin} disabled={isLoading}>
				{isLoading ? 'Processing...' : 'Login'}
			</button>
			<button class="btn-setup" on:click={doSetup} disabled={isLoading}>
				{isLoading ? 'Processing...' : 'Initial Setup'}
			</button>
		</div>

		{#if message}
			<div
				class="message"
				class:error={message.includes('Error') || message.includes('failed')}
				class:success={message.includes('success') || message.includes('complete')}
			>
				{message}
			</div>
		{/if}
	</div>
</div>

<style>
	.container {
		max-width: 48rem;
		margin: 0 auto;
		padding: 1rem;
	}

	.form-group {
		margin-bottom: 1rem;
	}

	label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
	}

	input {
		width: 100%;
		padding: 0.5rem;
		border: 1px solid #ccc;
		border-radius: 0.25rem;
	}

	input:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.button-group {
		display: flex;
		gap: 0.5rem;
		margin-top: 1.5rem;
	}

	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.25rem;
		cursor: pointer;
		font-weight: 500;
		transition: opacity 0.2s;
	}

	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	button:not(:disabled):hover {
		opacity: 0.9;
	}

	.btn-login {
		background-color: #2563eb;
		color: white;
	}

	.btn-setup {
		background-color: #16a34a;
		color: white;
	}

	.message {
		margin-top: 1rem;
		padding: 0.75rem;
		border-radius: 0.25rem;
		font-size: 0.875rem;
	}

	.message.error {
		background-color: #fee2e2;
		color: #991b1b;
		border: 1px solid #fca5a5;
	}

	.message.success {
		background-color: #d1fae5;
		color: #065f46;
		border: 1px solid #6ee7b7;
	}

	.password-requirements {
		font-size: 0.75rem;
		color: #6b7280;
		margin-top: 0.25rem;
	}

	.password-requirements ul {
		margin: 0.25rem 0 0 1.25rem;
		padding: 0;
	}
</style>
