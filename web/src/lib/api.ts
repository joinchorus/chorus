import type {
  Thread,
  Message,
  ThreadDetail,
  CreateThreadPayload,
  CreateMessagePayload,
  APIErrorResponse,
} from '../types';

const API_BASE = '/api/v1';

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorMsg = 'An unexpected error occurred.';
    try {
      const data: APIErrorResponse = await res.json();
      if (data.error && data.error.message) {
        errorMsg = data.error.message;
      }
    } catch {
      // Use default error string if body decoding fails
    }
    throw new Error(errorMsg);
  }
  return res.json();
}

export async function fetchThreads(): Promise<Thread[]> {
  const res = await fetch(`${API_BASE}/threads`);
  const data = await handleResponse<{ threads: Thread[] }>(res);
  return data.threads || [];
}

export async function fetchThreadDetail(threadId: string): Promise<ThreadDetail> {
  const res = await fetch(`${API_BASE}/threads/${threadId}`);
  return handleResponse<ThreadDetail>(res);
}

export async function createThread(payload: CreateThreadPayload): Promise<Thread> {
  const res = await fetch(`${API_BASE}/threads`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<Thread>(res);
}

export async function createMessage(
  threadId: string,
  payload: CreateMessagePayload
): Promise<Message> {
  const res = await fetch(`${API_BASE}/threads/${threadId}/messages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<Message>(res);
}
