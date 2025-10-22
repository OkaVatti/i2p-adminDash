// web/src/lib/api.ts
export const API_BASE = (import.meta.env.VITE_API_BASE as string) || 'http://127.0.0.1:8080';

export type SetupRequest = {
  username: string;
  email: string;
  sha3sha3: string;
};

export type LoginRequest = {
  username: string;
  sha3: string;
};

export type TokenResponse = {
  token?: string;
  error?: string;
  [k: string]: any;
};

export async function login(username: string, sha3hex: string): Promise<TokenResponse> {
  const res = await fetch(`${API_BASE}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, sha3: sha3hex } as LoginRequest)
  });
  return res.json();
}

export async function setup(username: string, email: string, sha3hex: string): Promise<any> {
  const res = await fetch(`${API_BASE}/api/setup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, sha3sha3: sha3hex } as SetupRequest)
  });
  return res.json();
}
