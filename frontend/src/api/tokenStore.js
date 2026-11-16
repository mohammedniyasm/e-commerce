let session = null;
const listeners = new Set();

function decodePayload(accessToken) {
  try {
    const payload = accessToken.split('.')[1];
    const normalizedPayload = payload.replace(/-/g, '+').replace(/_/g, '/');
    const json = decodeURIComponent(
      atob(normalizedPayload)
        .split('')
        .map((character) => `%${`00${character.charCodeAt(0).toString(16)}`.slice(-2)}`)
        .join(''),
    );

    return JSON.parse(json);
  } catch {
    return null;
  }
}

function notify() {
  listeners.forEach((listener) => listener(session));
}

export function setAccessToken(accessToken) {
  const payload = decodePayload(accessToken);

  session = {
    accessToken,
    role: payload?.role ?? null,
    userId: payload?.sub ?? null,
    expiresAt: payload?.exp ? payload.exp * 1000 : null,
  };
  notify();
}

export function clearAccessToken() {
  session = null;
  notify();
}

export function getAccessToken() {
  return session?.accessToken ?? null;
}

export function getSession() {
  return session;
}

export function subscribeToSession(listener) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}
