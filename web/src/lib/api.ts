import type {
  Identity,
  Thread,
  Message,
  ThreadDetail,
  CreateThreadPayload,
  CreateMessagePayload,
  TranslationRecord,
  ReportRecord,
  ReportReason,
  ModerationQueueItem,
  ModerationAction,
  ModerationStatus,
  APIErrorResponse,
} from '../types';

const API_BASE = import.meta.env.VITE_API_BASE || '/api/v0.1';

export function formatDate(dateStr?: string | null): string {
  if (!dateStr) return '';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return dateStr;
    return d.toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateStr || '';
  }
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorMsg = `HTTP Error ${res.status}`;
    try {
      const data: APIErrorResponse = await res.json();
      if (data.error && data.error.message) {
        errorMsg = data.error.message;
      }
    } catch {
      // Body is not JSON
    }
    throw new Error(errorMsg);
  }

  const contentType = res.headers.get('content-type') || '';
  if (!contentType.includes('application/json')) {
    throw new Error('Backend API unreachable or proxy unconfigured (received HTML response)');
  }

  return res.json();
}

export async function fetchNewConversationName(): Promise<Identity> {
  const res = await fetch(`${API_BASE}/identities`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({}),
  });
  return handleResponse<Identity>(res);
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

export async function translateMessage(
  threadId: string,
  messageId: string,
  targetLang: string = 'en'
): Promise<TranslationRecord> {
  const res = await fetch(`${API_BASE}/threads/${threadId}/messages/${messageId}/translate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ target_lang: targetLang }),
  });
  return handleResponse<TranslationRecord>(res);
}

export async function reportMessage(
  threadId: string,
  messageId: string,
  reason: ReportReason,
  details?: string
): Promise<ReportRecord> {
  const res = await fetch(`${API_BASE}/threads/${threadId}/messages/${messageId}/report`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ reason, details }),
  });
  return handleResponse<ReportRecord>(res);
}

export async function fetchModerationQueue(): Promise<ModerationQueueItem[]> {
  const res = await fetch(`${API_BASE}/moderation/reports`);
  const data = await handleResponse<{ reports: ModerationQueueItem[] }>(res);
  return data.reports || [];
}

export async function fetchModerationDetail(reportId: string): Promise<ModerationQueueItem> {
  const res = await fetch(`${API_BASE}/moderation/reports/${reportId}`);
  return handleResponse<ModerationQueueItem>(res);
}

export async function submitModerationAction(
  reportId: string,
  status: ModerationStatus,
  note?: string
): Promise<ModerationAction> {
  const res = await fetch(`${API_BASE}/moderation/reports/${reportId}/action`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status, note }),
  });
  return handleResponse<ModerationAction>(res);
}
