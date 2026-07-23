export interface Topic {
  id: string;
  name: string;
  description: string;
  official: boolean;
  votesCount?: number;
}

export const OFFICIAL_TOPICS: Topic[] = [
  { id: 'technology', name: 'Technology', description: 'Computing, software, hardware, and digital tools.', official: true },
  { id: 'programming', name: 'Programming', description: 'Languages, architecture, frameworks, and engineering.', official: true },
  { id: 'ai', name: 'Artificial Intelligence', description: 'Machine learning, neural networks, LLMs, and robotics.', official: true },
  { id: 'science', name: 'Science', description: 'Physics, biology, astronomy, chemistry, and research.', official: true },
  { id: 'design', name: 'Design', description: 'UI/UX, typography, industrial design, and aesthetics.', official: true },
  { id: 'philosophy', name: 'Philosophy', description: 'Ethics, logic, metaphysics, and human thought.', official: true },
  { id: 'books', name: 'Books', description: 'Literature, non-fiction, essays, and reading.', official: true },
  { id: 'movies', name: 'Movies', description: 'Cinema, filmmaking, screenwriting, and animation.', official: true },
  { id: 'music', name: 'Music', description: 'Composition, genres, instruments, and audio theory.', official: true },
  { id: 'gaming', name: 'Gaming', description: 'Game design, systems, mechanics, and interactive art.', official: true },
  { id: 'history', name: 'History', description: 'Historical events, civilizations, eras, and archives.', official: true },
  { id: 'politics', name: 'Politics', description: 'Civics, governance, political economy, and policy.', official: true },
  { id: 'sports', name: 'Sports', description: 'Athletics, competition, training, and strategy.', official: true },
];

export interface ProposedTopic {
  id: string;
  name: string;
  description: string;
  approves: number;
  rejects: number;
}

export const INITIAL_PROPOSED_TOPICS: ProposedTopic[] = [
  { id: 'prop-1', name: 'Cybersecurity', description: 'Cryptography, threat models, network security, and privacy tools.', approves: 14, rejects: 2 },
  { id: 'prop-2', name: 'Architecture & Cities', description: 'Urban planning, structural engineering, building design, and public spaces.', approves: 9, rejects: 1 },
  { id: 'prop-3', name: 'Mathematics', description: 'Pure math, probability, number theory, and proofs.', approves: 18, rejects: 0 }
];
