import type {
  Identity,
  Board,
  Thread,
  Message,
  ThreadDetail,
  CreateThreadPayload,
  CreateMessagePayload,
  TranslationRecord,
  ReportRecord,
  ReportReason,
  ModerationStatus,
  ModerationAction,
  ModerationQueueItem,
  APIErrorResponse,
} from '../types';

const API_BASE = import.meta.env.VITE_API_BASE || '/api/v0.1';

export function getCountryEmoji(countryCode?: string | null): string {
  if (!countryCode) return '';
  const code = countryCode.trim().toUpperCase();
  if (code.length !== 2) return '';
  const codePoints = code
    .split('')
    .map((char) => 127397 + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
}

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

export const SYSTEM_BOARDS: Board[] = [
  { id: 'brd_technology', slug: 'technology', display_name: 'Technology', description: 'General discussions about technology.' },
  { id: 'brd_programming', slug: 'programming', display_name: 'Programming', description: 'Software engineering, languages, tooling and architecture.' },
  { id: 'brd_ai', slug: 'ai', display_name: 'Artificial Intelligence', description: 'AI, machine learning, neural models and autonomous systems.' },
  { id: 'brd_science', slug: 'science', display_name: 'Science', description: 'Natural sciences, physics, biology, and scientific discoveries.' },
  { id: 'brd_design', slug: 'design', display_name: 'Design', description: 'Product design, UX, typography and visual systems.' },
  { id: 'brd_philosophy', slug: 'philosophy', display_name: 'Philosophy', description: 'Ethics, metaphysics, logic, and existential thought.' },
  { id: 'brd_politics', slug: 'politics', display_name: 'Politics', description: 'Political theory, governance, and public policy.' },
  { id: 'brd_history', slug: 'history', display_name: 'History', description: 'Historical events, eras, and historiography.' },
  { id: 'brd_books', slug: 'books', display_name: 'Books', description: 'Literature, prose, and reading.' },
  { id: 'brd_movies', slug: 'movies', display_name: 'Movies', description: 'Cinema, film theory, and filmmaking.' },
  { id: 'brd_music', slug: 'music', display_name: 'Music', description: 'Acoustics, composition, genres, and audio.' },
  { id: 'brd_gaming', slug: 'gaming', display_name: 'Gaming', description: 'Game design, mechanics, and interactive media.' },
  { id: 'brd_cybersecurity', slug: 'cybersecurity', display_name: 'Cybersecurity', description: 'Security, cryptography, and privacy engineering.' },
  { id: 'brd_mathematics', slug: 'mathematics', display_name: 'Mathematics', description: 'Pure and applied mathematics, proof, and computation.' },
  { id: 'brd_engineering', slug: 'engineering', display_name: 'Engineering', description: 'Systems, hardware, and physical engineering.' },
  { id: 'brd_economics', slug: 'economics', display_name: 'Economics', description: 'Markets, incentive design, and economic theory.' },
  { id: 'brd_psychology', slug: 'psychology', display_name: 'Psychology', description: 'Cognition, behavior, and mental processes.' },
];

const MOCK_THREADS: Thread[] = [];

export async function fetchBoards(): Promise<Board[]> {
  try {
    const res = await fetch(`${API_BASE}/boards`);
    const data = await handleResponse<{ boards: Board[] }>(res);
    return data.boards || SYSTEM_BOARDS;
  } catch {
    return SYSTEM_BOARDS;
  }
}

export async function fetchBoardBySlug(slug: string): Promise<Board> {
  try {
    const res = await fetch(`${API_BASE}/boards/${slug}`);
    return await handleResponse<Board>(res);
  } catch {
    const b = SYSTEM_BOARDS.find((x) => x.slug === slug);
    if (b) return b;
    throw new Error('Board not found');
  }
}

export async function fetchNewConversationName(): Promise<Identity> {
  try {
    const res = await fetch(`${API_BASE}/identities`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    });
    return await handleResponse<Identity>(res);
  } catch {
    const names = ['River', 'Echo', 'Ash', 'Stone', 'Willow', 'Cedar', 'Falcon', 'North', 'Quartz', 'Juniper'];
    const chosen = names[Math.floor(Math.random() * names.length)];
    return { conversation_name: chosen };
  }
}

export function getThreadParticipantToken(threadId: string): string | null {
  try {
    return sessionStorage.getItem(`chorus_ptk_${threadId}`);
  } catch {
    return null;
  }
}

export function setThreadParticipantToken(threadId: string, token: string): void {
  try {
    if (token) {
      sessionStorage.setItem(`chorus_ptk_${threadId}`, token);
    }
  } catch {
    // Storage unavailable or blocked
  }
}

export async function fetchThreads(boardSlug?: string): Promise<Thread[]> {
  try {
    const url = boardSlug && boardSlug !== 'all'
      ? `${API_BASE}/threads?board=${encodeURIComponent(boardSlug)}`
      : `${API_BASE}/threads`;
    const res = await fetch(url);
    const data = await handleResponse<{ threads: Thread[] }>(res);
    return data.threads || [];
  } catch (err) {
    console.warn('Backend API unreachable, using local fallback threads:', err);
    if (boardSlug && boardSlug !== 'all') {
      return MOCK_THREADS.filter((t) => t.board_slug === boardSlug);
    }
    return MOCK_THREADS;
  }
}

export async function fetchThreadDetail(threadId: string): Promise<ThreadDetail> {
  try {
    const res = await fetch(`${API_BASE}/threads/${threadId}`);
    return await handleResponse<ThreadDetail>(res);
  } catch (err) {
    console.warn('Backend API unreachable, using fallback thread detail:', err);
    const thread = MOCK_THREADS.find((t) => t.id === threadId) || MOCK_THREADS[0];
    return {
      thread,
      messages: [
        {
          id: 'msg-1',
          thread_id: thread.id,
          content: thread.body || thread.preview || '',
          conversation_name: thread.conversation_name,
          country: thread.country,
          created_at: thread.created_at,
        },
      ],
    };
  }
}

export async function createThread(payload: CreateThreadPayload): Promise<Thread> {
  try {
    const res = await fetch(`${API_BASE}/threads`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const thread = await handleResponse<Thread>(res);
    if (thread && thread.id && thread.participant_token) {
      setThreadParticipantToken(thread.id, thread.participant_token);
    }
    return thread;
  } catch (err) {
    console.warn('Backend API unreachable, creating local mock thread:', err);
    const boardSlug = payload.board_slug || (payload.topic ? payload.topic.toLowerCase() : 'technology');
    const matchedBoard = SYSTEM_BOARDS.find((b) => b.slug === boardSlug);
    const newTh: Thread = {
      id: `thd_${Date.now()}`,
      topic: matchedBoard ? matchedBoard.display_name : payload.topic || 'Technology',
      board_slug: boardSlug,
      board_display_name: matchedBoard ? matchedBoard.display_name : payload.topic || 'Technology',
      title: payload.title,
      body: payload.body,
      preview: (payload.body || '').slice(0, 120),
      conversation_name: payload.conversation_name || 'Anonymous',
      country: payload.show_country ? 'TR' : null,
      message_count: 0,
      participant_count: 1,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    MOCK_THREADS.unshift(newTh);
    return newTh;
  }
}

export async function createMessage(
  threadId: string,
  payload: CreateMessagePayload
): Promise<Message> {
  const ptk = payload.participant_token || getThreadParticipantToken(threadId) || undefined;
  const fullPayload = {
    ...payload,
    participant_token: ptk,
  };
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (ptk) {
    headers['X-Participant-Token'] = ptk;
  }

  try {
    const res = await fetch(`${API_BASE}/threads/${threadId}/messages`, {
      method: 'POST',
      headers,
      body: JSON.stringify(fullPayload),
    });
    const message = await handleResponse<Message>(res);
    if (message && message.participant_token) {
      setThreadParticipantToken(threadId, message.participant_token);
    }
    return message;
  } catch (err) {
    console.warn('Backend API unreachable, creating local mock message:', err);
    return {
      id: `msg_${Date.now()}`,
      thread_id: threadId,
      content: payload.body,
      conversation_name: payload.conversation_name || 'Anonymous',
      country: payload.show_country ? 'TR' : null,
      created_at: new Date().toISOString(),
    };
  }
}

export async function translateMessage(
  threadId: string,
  messageId: string,
  targetLang?: string,
  textToTranslate?: string
): Promise<TranslationRecord> {
  const userLang = typeof navigator !== 'undefined' && navigator.language
    ? navigator.language.split('-')[0].toLowerCase()
    : 'tr';

  let lang = targetLang || userLang;

  try {
    const res = await fetch(`${API_BASE}/threads/${threadId}/messages/${messageId}/translate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target_lang: lang }),
    });
    return await handleResponse<TranslationRecord>(res);
  } catch {
    if (textToTranslate) {
      try {
        let gtxUrl = `https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=${lang}&dt=t&q=${encodeURIComponent(textToTranslate)}`;
        let gtxRes = await fetch(gtxUrl);
        let data = await gtxRes.json();
        let detectedSource = data && data[2] ? data[2] : '';

        // If source language matches target language, translate to the alternate language (e.g. tr <-> en)
        if (detectedSource && detectedSource === lang) {
          const altLang = lang === 'tr' ? 'en' : 'tr';
          gtxUrl = `https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=${altLang}&dt=t&q=${encodeURIComponent(textToTranslate)}`;
          gtxRes = await fetch(gtxUrl);
          data = await gtxRes.json();
          lang = altLang;
        }

        let translatedText = '';
        if (Array.isArray(data) && Array.isArray(data[0])) {
          translatedText = data[0].map((part: any) => (part && part[0] ? part[0] : '')).join('');
        }

        if (translatedText) {
          return {
            message_id: messageId,
            target_lang: lang,
            translated_text: translatedText,
            provider: 'Google Translate',
          };
        }
      } catch (err) {
        console.warn('Free client translation failed:', err);
      }
    }
    return {
      message_id: messageId,
      target_lang: lang,
      translated_text: textToTranslate || 'Translation unavailable.',
      provider: 'mock',
    };
  }
}

export async function reportMessage(
  threadId: string,
  messageId: string,
  reason: ReportReason,
  details?: string
): Promise<ReportRecord> {
  try {
    const res = await fetch(`${API_BASE}/threads/${threadId}/messages/${messageId}/report`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ reason, details }),
    });
    return await handleResponse<ReportRecord>(res);
  } catch {
    const report: ReportRecord = {
      id: `rep_${Date.now()}`,
      thread_id: threadId,
      message_id: messageId,
      reason,
      details,
      created_at: new Date().toISOString(),
    };
    MOCK_MODERATION_ITEMS.unshift({
      report,
      message: {
        id: messageId,
        thread_id: threadId,
        conversation_name: 'Reported User',
        country: 'US',
        content: details || `Reported message content (${reason})`,
        created_at: new Date().toISOString(),
      },
      current_status: 'pending',
      history: [],
    });
    return report;
  }
}

const MOCK_MODERATION_ITEMS: ModerationQueueItem[] = [];

export async function loginAdmin(token: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/moderation/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
    credentials: 'same-origin',
  });
  await handleResponse<{ status: string }>(res);
  return true;
}

export async function logoutAdmin(): Promise<boolean> {
  const res = await fetch(`${API_BASE}/moderation/logout`, {
    method: 'POST',
    credentials: 'same-origin',
  });
  await handleResponse<{ status: string }>(res);
  return true;
}

export async function checkAdminSession(): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/moderation/session`, { credentials: 'same-origin' });
    const data = await handleResponse<{ authenticated: boolean }>(res);
    return data.authenticated;
  } catch {
    return false;
  }
}

export async function fetchModerationQueue(licenseKey?: string): Promise<ModerationQueueItem[]> {
  const key = licenseKey || localStorage.getItem('chorus_admin_license_key') || '';
  const headers: Record<string, string> = {};
  if (key) {
    headers['Authorization'] = `Bearer ${key}`;
    headers['X-Admin-License-Key'] = key;
  }

  const res = await fetch(`${API_BASE}/moderation/reports`, { headers, credentials: 'same-origin' });
  const data = await handleResponse<{ reports: ModerationQueueItem[] }>(res);
  return data.reports || [];
}

export async function fetchModerationReportDetail(reportId: string, licenseKey?: string): Promise<ModerationQueueItem> {
  const key = licenseKey || localStorage.getItem('chorus_admin_license_key') || '';
  const headers: Record<string, string> = {};
  if (key) {
    headers['Authorization'] = `Bearer ${key}`;
    headers['X-Admin-License-Key'] = key;
  }

  const res = await fetch(`${API_BASE}/moderation/reports/${reportId}`, { headers, credentials: 'same-origin' });
  return await handleResponse<ModerationQueueItem>(res);
}

export async function submitModerationAction(
  reportId: string,
  status: ModerationStatus,
  note?: string,
  licenseKey?: string
): Promise<ModerationAction> {
  const key = licenseKey || localStorage.getItem('chorus_admin_license_key') || '';
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (key) {
    headers['Authorization'] = `Bearer ${key}`;
    headers['X-Admin-License-Key'] = key;
  }

  const res = await fetch(`${API_BASE}/moderation/reports/${reportId}/action`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ status, note }),
    credentials: 'same-origin',
  });
  return await handleResponse<ModerationAction>(res);
}

