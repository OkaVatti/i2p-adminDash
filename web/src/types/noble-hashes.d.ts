declare module '@noble/hashes/sha3' {
  export function sha3_512(data: Uint8Array): Uint8Array;
}

declare module '@noble/hashes/utils' {
  export function bytesToHex(bytes: Uint8Array): string;
  export function hexToBytes(hex: string): Uint8Array;
  export function utf8ToBytes(str: string): Uint8Array;
}