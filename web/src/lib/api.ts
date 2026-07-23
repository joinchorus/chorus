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

const nowStr = new Date().toISOString();

const MOCK_THREADS: Thread[] = [
  {
    id: 'thd_mock1',
    topic: 'Philosophy',
    board_slug: 'philosophy',
    board_display_name: 'Philosophy',
    title: 'Identity in the Digital Age: Why anonymity enables truth',
    preview: 'When identity is attached to reputation, speech becomes performative. Anonymous discussion forces ideas to stand on their own merit.',
    body: 'When identity is attached to reputation, speech becomes performative. Anonymous discussion forces ideas to stand on their own merit. What are the long-term consequences of permanent digital footprints on free thought?',
    conversation_name: 'River',
    country: 'DE',
    message_count: 14,
    participant_count: 6,
    created_at: new Date(Date.now() - 3600000 * 2).toISOString(),
    updated_at: nowStr,
  },
  {
    id: 'thd_mock2',
    topic: 'Programming',
    board_slug: 'programming',
    board_display_name: 'Programming',
    title: 'The shift from monolithic frameworks to zero-dependency architectures',
    preview: 'Exploring the modern resurgence of minimal dependency graphs and first-principles software architecture.',
    body: 'Exploring the modern resurgence of minimal dependency graphs and first-principles software architecture. Is complexity in web tooling self-inflicted?',
    conversation_name: 'Echo',
    country: 'US',
    message_count: 28,
    participant_count: 9,
    created_at: new Date(Date.now() - 3600000 * 5).toISOString(),
    updated_at: nowStr,
  },
  {
    id: 'thd_mock3',
    topic: 'AI',
    board_slug: 'ai',
    board_display_name: 'Artificial Intelligence',
    title: 'Evaluating autonomous agent reasoning without human bias',
    preview: 'How do we measure true reasoning capabilities in non-deterministic AI agents when benchmark datasets are leaked into training corpora?',
    body: 'How do we measure true reasoning capabilities in non-deterministic AI agents when benchmark datasets are leaked into training corpora?',
    conversation_name: 'Quartz',
    country: 'JP',
    message_count: 42,
    participant_count: 15,
    created_at: new Date(Date.now() - 3600000 * 12).toISOString(),
    updated_at: nowStr,
  },
  {
    id: 'thd_mock4',
    topic: 'Design',
    board_slug: 'design',
    board_display_name: 'Design',
    title: 'Quiet design systems: Why less interface means better thinking',
    preview: 'High-density dashboards cause fatigue. Editorial typography and calm layout allow content to take center stage.',
    body: 'High-density dashboards cause fatigue. Editorial typography and calm layout allow content to take center stage.',
    conversation_name: 'Cedar',
    country: 'SE',
    message_count: 9,
    participant_count: 4,
    created_at: new Date(Date.now() - 3600000 * 20).toISOString(),
    updated_at: nowStr,
  },
];

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

export async function fetchThreads(): Promise<Thread[]> {
  try {
    const res = await fetch(`${API_BASE}/threads`);
    const data = await handleResponse<{ threads: Thread[] }>(res);
    return data.threads || [];
  } catch (err) {
    console.warn('Backend API unreachable, using local fallback threads:', err);
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
    return await handleResponse<Thread>(res);
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
      country: 'TR',
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
  try {
    const res = await fetch(`${API_BASE}/threads/${threadId}/messages`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return await handleResponse<Message>(res);
  } catch (err) {
    console.warn('Backend API unreachable, creating local mock message:', err);
    return {
      id: `msg_${Date.now()}`,
      thread_id: threadId,
      content: payload.body,
      conversation_name: payload.conversation_name || 'Anonymous',
      country: 'TR',
      created_at: new Date().toISOString(),
    };
  }
}

export async function translateMessage(
  threadId: string,
  messageId: string,
  targetLang: string = 'en'
): Promise<TranslationRecord> {
  try {
    const res = await fetch(`${API_BASE}/threads/${threadId}/messages/${messageId}/translate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target_lang: targetLang }),
    });
    return await handleResponse<TranslationRecord>(res);
  } catch {
    return {
      message_id: messageId,
      target_lang: targetLang,
      translated_text: 'Translation unavailable in offline dev mode.',
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
    return {
      id: `rep_${Date.now()}`,
      thread_id: threadId,
      message_id: messageId,
      reason,
      details,
      created_at: new Date().toISOString(),
    };
  }
}
