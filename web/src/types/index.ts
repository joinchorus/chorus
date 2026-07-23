export interface Identity {
  id?: string;
  conversation_name: string;
  country?: string | null;
  created_at?: string;
}

export interface Thread {
  id: string;
  title: string;
  topic?: string;
  body?: string;
  preview?: string;
  author_id?: string;
  conversation_name: string;
  country?: string | null;
  created_at: string;
  updated_at: string;
  participant_count?: number;
  message_count?: number;
  is_pinned?: boolean;
}

export interface Message {
  id: string;
  thread_id: string;
  author_id?: string;
  conversation_name: string;
  country?: string | null;
  content: string;
  created_at: string;
  translated_content?: string;
}

export interface ThreadDetail {
  thread: Thread;
  messages: Message[];
}

export interface CreateThreadPayload {
  title: string;
  topic?: string;
  body?: string;
  show_country?: boolean;
  conversation_name?: string;
}

export interface CreateMessagePayload {
  body: string;
  show_country?: boolean;
  conversation_name?: string;
}

export interface TranslationRecord {
  message_id: string;
  target_lang: string;
  translated_text: string;
  provider: string;
}

export type ReportReason = 'spam' | 'harassment' | 'illegal' | 'violence' | 'copyright' | 'other';

export interface ReportRecord {
  id: string;
  thread_id: string;
  message_id: string;
  reason: ReportReason;
  details?: string;
  created_at: string;
}

export type ModerationStatus = 'pending' | 'reviewed' | 'dismissed' | 'removed';

export interface ModerationAction {
  id: string;
  report_id: string;
  thread_id: string;
  message_id: string;
  status: ModerationStatus;
  note?: string;
  created_at: string;
}

export interface ModerationQueueItem {
  report: ReportRecord;
  message?: Message | null;
  current_status: ModerationStatus;
  history: ModerationAction[];
}

export interface APIErrorResponse {
  error: {
    code: string;
    message: string;
  };
}
