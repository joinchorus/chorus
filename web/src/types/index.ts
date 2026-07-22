export interface Identity {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface Thread {
  id: string;
  title: string;
  author_id: string;
  country?: string | null;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  thread_id: string;
  author_id: string;
  country?: string | null;
  content: string;
  created_at: string;
}

export interface ThreadDetail {
  thread: Thread;
  messages: Message[];
}

export interface CreateThreadPayload {
  title: string;
  body?: string;
  show_country?: boolean;
}

export interface CreateMessagePayload {
  body: string;
  show_country?: boolean;
}

export interface APIErrorResponse {
  error: {
    code: string;
    message: string;
  };
}
