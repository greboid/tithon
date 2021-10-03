import { writable } from 'svelte/store';

export const loggedIn = writable(false);
export const connected = writable(false);
export const serverList = writable(new Map())
export const messages = writable([])
export const selectedNetwork = writable("")
export const selectedChannel = writable("")