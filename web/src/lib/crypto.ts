// web/src/lib/crypto.ts
import { sha3_512 } from '@noble/hashes/sha3.js';
import { bytesToHex, utf8ToBytes } from '@noble/hashes/utils.js';

/**
 * Hash a UTF-8 string with SHA3-512
 * @param input - The input string to hash
 * @returns The hexadecimal hash string
 */
export function hashSHA3_512(input: string): string {
	const bytes = utf8ToBytes(input);
	const hash = sha3_512(bytes);
	return bytesToHex(hash);
}

/**
 * Hash raw bytes with SHA3-512
 * @param input - The input bytes to hash
 * @returns The hexadecimal hash string
 */
export function hashSHA3_512Bytes(input: Uint8Array): string {
	const hash = sha3_512(input);
	return bytesToHex(hash);
}

/**
 * Validate password meets minimum requirements
 * @param password - The password to validate
 * @returns Object with isValid boolean and error message if invalid
 */
export function validatePassword(password: string): { isValid: boolean; error?: string } {
	if (password.length < 32) {
		return { isValid: false, error: 'Password must be at least 32 characters long' };
	}

	// Check for uppercase
	if (!/[A-Z]/.test(password)) {
		return { isValid: false, error: 'Password must contain at least one uppercase letter' };
	}

	// Check for lowercase
	if (!/[a-z]/.test(password)) {
		return { isValid: false, error: 'Password must contain at least one lowercase letter' };
	}

	// Check for digit
	if (!/[0-9]/.test(password)) {
		return { isValid: false, error: 'Password must contain at least one digit' };
	}

	// Check for special character
	if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
		return { isValid: false, error: 'Password must contain at least one special character' };
	}

	return { isValid: true };
}

/**
 * Validate email format (I2P email)
 * @param email - The email to validate
 * @returns true if valid I2P email format
 */
export function validateI2PEmail(email: string): boolean {
	const i2pEmailRegex = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.i2p$/;
	return i2pEmailRegex.test(email);
}
