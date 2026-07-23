import React, { useState } from 'react';
import type { Message as MessageType, TranslationRecord, ReportReason } from '../types';
import { translateMessage, reportMessage, formatDate, getCountryEmoji } from '../lib/api';
import { Button } from './ui/Button';
import { Badge } from './ui/Badge';

interface MessageProps {
  message: MessageType;
}

const REPORT_OPTIONS: { label: string; value: ReportReason }[] = [
  { label: 'Spam', value: 'spam' },
  { label: 'Harassment', value: 'harassment' },
  { label: 'Illegal content', value: 'illegal' },
  { label: 'Violence', value: 'violence' },
  { label: 'Copyright', value: 'copyright' },
  { label: 'Other', value: 'other' },
];

export const Message: React.FC<MessageProps> = ({ message }) => {
  const [translation, setTranslation] = useState<TranslationRecord | null>(null);
  const [isTranslating, setIsTranslating] = useState(false);
  const [showTranslated, setShowTranslated] = useState(false);
  const [transError, setTransError] = useState<string | null>(null);

  // Report state
  const [showReportForm, setShowReportForm] = useState(false);
  const [selectedReason, setSelectedReason] = useState<ReportReason>('spam');
  const [isSubmittingReport, setIsSubmittingReport] = useState(false);
  const [reportSubmitted, setReportSubmitted] = useState(false);
  const [reportError, setReportError] = useState<string | null>(null);

  const formattedDate = formatDate(message.created_at);
  const authorName = message.conversation_name || 'Anonymous';

  const handleTranslateClick = async () => {
    if (showTranslated) {
      setShowTranslated(false);
      return;
    }

    if (translation) {
      setShowTranslated(true);
      return;
    }

    setIsTranslating(true);
    setTransError(null);

    try {
      const res = await translateMessage(message.thread_id, message.id, 'en');
      setTranslation(res);
      setShowTranslated(true);
    } catch (err) {
      setTransError(err instanceof Error ? err.message : 'Translation failed');
    } finally {
      setIsTranslating(false);
    }
  };

  const handleReportSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmittingReport(true);
    setReportError(null);

    try {
      await reportMessage(message.thread_id, message.id, selectedReason);
      setReportSubmitted(true);
      setShowReportForm(false);
    } catch (err) {
      setReportError(err instanceof Error ? err.message : 'Report failed');
    } finally {
      setIsSubmittingReport(false);
    }
  };

  return (
    <article className="message-item">
      <header className="message-header">
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <Badge variant="name" className="message-author">{authorName}</Badge>
          {message.country && (
            <span title={`Country: ${message.country}`} aria-label={`Country: ${message.country}`}>
              {getCountryEmoji(message.country)}
            </span>
          )}
          <span>&bull;</span>
          <span className="message-timestamp">{formattedDate}</span>
        </div>

        {/* Action buttons */}
        <div className="message-actions">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleTranslateClick}
            disabled={isTranslating}
            title="Translate message to English"
          >
            {isTranslating
              ? 'Translating...'
              : showTranslated
              ? 'Show original'
              : translation
              ? 'Show translation'
              : 'Translate'}
          </Button>

          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowReportForm(!showReportForm)}
            disabled={reportSubmitted}
          >
            {reportSubmitted ? 'Reported' : 'Report'}
          </Button>
        </div>
      </header>

      {/* Message Content */}
      <div className="message-body">
        {showTranslated && translation ? translation.translated_text : message.content}
      </div>

      {showTranslated && translation && (
        <div style={{ marginTop: '0.5rem', fontSize: '0.75rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>
          Translated via {translation.provider} backend provider
        </div>
      )}

      {/* Report Form Selector */}
      {showReportForm && !reportSubmitted && (
        <form onSubmit={handleReportSubmit} className="composer-box" style={{ marginTop: '1rem' }}>
          <div style={{ fontWeight: 600, marginBottom: '0.5rem', fontSize: '0.875rem' }}>
            Report this message
          </div>
          <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '1rem' }}>
            {REPORT_OPTIONS.map((opt) => (
              <label key={opt.value} className="checkbox-label">
                <input
                  type="radio"
                  name="reportReason"
                  value={opt.value}
                  checked={selectedReason === opt.value}
                  onChange={() => setSelectedReason(opt.value)}
                />
                {opt.label}
              </label>
            ))}
          </div>

          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <Button type="submit" size="sm" disabled={isSubmittingReport}>
              {isSubmittingReport ? 'Submitting...' : 'Submit Report'}
            </Button>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => setShowReportForm(false)}
            >
              Cancel
            </Button>
          </div>

          {reportError && <div className="form-error">{reportError}</div>}
        </form>
      )}

      {reportSubmitted && (
        <div style={{ marginTop: '0.5rem', fontSize: '0.8125rem', color: 'var(--accent-primary)' }}>
          Report submitted. Thank you.
        </div>
      )}

      {transError && <div className="form-error">{transError}</div>}
    </article>
  );
};
