const STORAGE_KEY = "wacalls.apiKey";
const SIGNOUT = "wacalls-signout";

export const getApiKey = (): string => localStorage.getItem(STORAGE_KEY) ?? "";

export const setApiKey = (key: string): void => {
  const trimmed = key.trim();
  if (trimmed) localStorage.setItem(STORAGE_KEY, trimmed);
  else localStorage.removeItem(STORAGE_KEY);
};

export const signOut = (): void => {
  setApiKey("");
  window.dispatchEvent(new Event(SIGNOUT));
};

export const onSignOut = (fn: () => void): (() => void) => {
  window.addEventListener(SIGNOUT, fn);
  return () => window.removeEventListener(SIGNOUT, fn);
};
