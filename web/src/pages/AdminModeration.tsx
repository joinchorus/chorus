import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  ShieldAlert,
  CheckCircle2,
  XCircle,
  Trash2,
  Clock,
  Filter,
  MessageSquare,
  ExternalLink,
  ChevronDown,
  ChevronUp,
  AlertTriangle,
  History,
} from 'lucide-react';
import type { ModerationQueueItem, ModerationStatus, ReportReason } from '../types';
import { fetchModerationQueue, submitModerationAction, formatDate, getCountryEmoji } from '../lib/api';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import { Skeleton } from '../components/ui/Skeleton';

export const AdminModeration: React.FC = () => {
  const [items, setItems] = useState<ModerationQueueItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [reasonFilter, setReasonFilter] = useState<string>('all');

  // Active action state
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [submittingId, setSubmittingId] = useState<string | null>(null);
  const [actionSuccess, setActionSuccess] = useState<string | null>(null);

  // Expanded history logs
  const [expandedHistory, setExpandedHistory] = useState<Record<string, boolean>>({});

  useEffect(() => {
    loadQueue();
  }, []);

  const loadQueue = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchModerationQueue();
      setItems(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load moderation queue');
    } finally {
      setLoading(false);
    }
  };

  const handleAction = async (reportId: string, status: ModerationStatus) => {
    setSubmittingId(reportId);
    setActionSuccess(null);
    const note = notes[reportId] || '';

    try {
      await submitModerationAction(reportId, status, note);
      setActionSuccess(`Report #${reportId} updated to ${status}`);
      // Refresh local queue state
      setItems((prev) =>
        prev.map((item) => {
          if (item.report.id === reportId) {
            const newHistory = [
              ...item.history,
              {
                id: `mod_${Date.now()}`,
                report_id: reportId,
                thread_id: item.report.thread_id,
                message_id: item.report.message_id,
                status,
                note,
                created_at: new Date().toISOString(),
              },
            ];
            return {
              ...item,
              current_status: status,
              history: newHistory,
            };
          }
          return item;
        })
      );
      // Clear note
      setNotes((prev) => ({ ...prev, [reportId]: '' }));
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setSubmittingId(null);
    }
  };

  const toggleHistory = (reportId: string) => {
    setExpandedHistory((prev) => ({ ...prev, [reportId]: !prev[reportId] }));
  };

  // Filtered Items
  const filteredItems = items.filter((item) => {
    if (statusFilter !== 'all' && item.current_status !== statusFilter) {
      return false;
    }
    if (reasonFilter !== 'all' && item.report.reason !== reasonFilter) {
      return false;
    }
    return true;
  });

  // Calculate statistics
  const totalCount = items.length;
  const pendingCount = items.filter((i) => i.current_status === 'pending').length;
  const reviewedCount = items.filter((i) => i.current_status === 'reviewed').length;
  const dismissedCount = items.filter((i) => i.current_status === 'dismissed').length;
  const removedCount = items.filter((i) => i.current_status === 'removed').length;

  const getReasonLabel = (reason: ReportReason) => {
    switch (reason) {
      case 'spam':
        return 'Spam';
      case 'harassment':
        return 'Harassment';
      case 'illegal':
        return 'Illegal Content';
      case 'violence':
        return 'Violence';
      case 'copyright':
        return 'Copyright';
      case 'other':
        return 'Other';
      default:
        return reason;
    }
  };

  return (
    <div className="admin-moderation-page">
      {/* Header Banner */}
      <header className="admin-header" style={{ marginBottom: '2rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.5rem' }}>
          <div
            style={{
              display: 'inline-flex',
              padding: '0.5rem',
              borderRadius: '0.5rem',
              backgroundColor: 'rgba(239, 68, 68, 0.1)',
              color: '#ef4444',
            }}
          >
            <ShieldAlert size={28} />
          </div>
          <div>
            <h1 style={{ fontSize: '1.75rem', fontWeight: 700, margin: 0, color: 'var(--text-primary)' }}>
              Moderation Dashboard
            </h1>
            <p style={{ margin: 0, color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
              Review user reports, inspect flagged messages, and enforce platform guidelines.
            </p>
          </div>
        </div>
      </header>

      {/* Success Notification */}
      {actionSuccess && (
        <div
          style={{
            padding: '0.75rem 1rem',
            marginBottom: '1.5rem',
            borderRadius: '0.5rem',
            backgroundColor: 'rgba(34, 197, 94, 0.15)',
            border: '1px solid rgba(34, 197, 94, 0.3)',
            color: '#22c55e',
            fontSize: '0.875rem',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <CheckCircle2 size={18} />
          <span>{actionSuccess}</span>
        </div>
      )}

      {/* Stats Cards */}
      <section
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
          gap: '1rem',
          marginBottom: '2rem',
        }}
      >
        <div
          style={{
            padding: '1.25rem',
            borderRadius: '0.75rem',
            backgroundColor: 'var(--bg-secondary)',
            border: '1px solid var(--border-color)',
          }}
        >
          <div style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>
            Total Queue
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: 'var(--text-primary)' }}>
            {totalCount}
          </div>
        </div>

        <div
          style={{
            padding: '1.25rem',
            borderRadius: '0.75rem',
            backgroundColor: 'var(--bg-secondary)',
            border: '1px solid rgba(234, 179, 8, 0.3)',
          }}
        >
          <div style={{ fontSize: '0.8125rem', color: '#eab308', marginBottom: '0.25rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <Clock size={14} /> Pending Review
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#eab308' }}>
            {pendingCount}
          </div>
        </div>

        <div
          style={{
            padding: '1.25rem',
            borderRadius: '0.75rem',
            backgroundColor: 'var(--bg-secondary)',
            border: '1px solid rgba(59, 130, 246, 0.3)',
          }}
        >
          <div style={{ fontSize: '0.8125rem', color: '#3b82f6', marginBottom: '0.25rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <CheckCircle2 size={14} /> Reviewed
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#3b82f6' }}>
            {reviewedCount}
          </div>
        </div>

        <div
          style={{
            padding: '1.25rem',
            borderRadius: '0.75rem',
            backgroundColor: 'var(--bg-secondary)',
            border: '1px solid rgba(239, 68, 68, 0.3)',
          }}
        >
          <div style={{ fontSize: '0.8125rem', color: '#ef4444', marginBottom: '0.25rem', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <Trash2 size={14} /> Content Removed
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 700, color: '#ef4444' }}>
            {removedCount}
          </div>
        </div>
      </section>

      {/* Filter Controls Bar */}
      <section
        style={{
          display: 'flex',
          flexWrap: 'wrap',
          gap: '1rem',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '1.5rem',
          padding: '1rem',
          backgroundColor: 'var(--bg-secondary)',
          borderRadius: '0.75rem',
          border: '1px solid var(--border-color)',
        }}
      >
        {/* Status Tabs */}
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
          {[
            { key: 'all', label: `All (${totalCount})` },
            { key: 'pending', label: `Pending (${pendingCount})` },
            { key: 'reviewed', label: `Reviewed (${reviewedCount})` },
            { key: 'dismissed', label: `Dismissed (${dismissedCount})` },
            { key: 'removed', label: `Removed (${removedCount})` },
          ].map((tab) => (
            <button
              key={tab.key}
              onClick={() => setStatusFilter(tab.key)}
              style={{
                padding: '0.4rem 0.85rem',
                borderRadius: '0.5rem',
                fontSize: '0.8125rem',
                fontWeight: 600,
                border: 'none',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                backgroundColor:
                  statusFilter === tab.key
                    ? 'var(--accent-primary, #6366f1)'
                    : 'var(--bg-tertiary, rgba(255,255,255,0.05))',
                color: statusFilter === tab.key ? '#ffffff' : 'var(--text-secondary)',
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Reason Filter */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Filter size={16} style={{ color: 'var(--text-muted)' }} />
          <select
            value={reasonFilter}
            onChange={(e) => setReasonFilter(e.target.value)}
            style={{
              padding: '0.4rem 0.75rem',
              borderRadius: '0.5rem',
              backgroundColor: 'var(--bg-primary)',
              color: 'var(--text-primary)',
              border: '1px solid var(--border-color)',
              fontSize: '0.8125rem',
            }}
          >
            <option value="all">All Reasons</option>
            <option value="spam">Spam</option>
            <option value="harassment">Harassment</option>
            <option value="illegal">Illegal Content</option>
            <option value="violence">Violence</option>
            <option value="copyright">Copyright</option>
            <option value="other">Other</option>
          </select>
        </div>
      </section>

      {/* Main Queue Content */}
      {loading ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          <Skeleton style={{ height: '120px' }} />
          <Skeleton style={{ height: '120px' }} />
          <Skeleton style={{ height: '120px' }} />
        </div>
      ) : error ? (
        <div className="form-error" style={{ padding: '1.5rem', textAlign: 'center' }}>
          <AlertTriangle size={24} style={{ marginBottom: '0.5rem' }} />
          <div>{error}</div>
          <Button onClick={loadQueue} size="sm" style={{ marginTop: '1rem' }}>
            Retry Loading
          </Button>
        </div>
      ) : filteredItems.length === 0 ? (
        <div
          style={{
            padding: '3rem 1.5rem',
            textAlign: 'center',
            backgroundColor: 'var(--bg-secondary)',
            borderRadius: '0.75rem',
            border: '1px dashed var(--border-color)',
            color: 'var(--text-secondary)',
          }}
        >
          <CheckCircle2 size={36} style={{ color: 'var(--text-muted)', marginBottom: '0.75rem' }} />
          <div style={{ fontWeight: 600, fontSize: '1rem', marginBottom: '0.25rem' }}>
            No reports in queue
          </div>
          <div style={{ fontSize: '0.875rem' }}>
            {statusFilter !== 'all' || reasonFilter !== 'all'
              ? 'Try adjusting your filters.'
              : 'All reports have been moderated.'}
          </div>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
          {filteredItems.map((item) => {
            const report = item.report;
            const msg = item.message;
            const isSubmitting = submittingId === report.id;
            const historyList = item.history || [];
            const isHistoryExpanded = !!expandedHistory[report.id];

            return (
              <article
                key={report.id}
                style={{
                  padding: '1.25rem',
                  borderRadius: '0.75rem',
                  backgroundColor: 'var(--bg-secondary)',
                  border:
                    item.current_status === 'pending'
                      ? '1px solid rgba(234, 179, 8, 0.4)'
                      : item.current_status === 'removed'
                      ? '1px solid rgba(239, 68, 68, 0.4)'
                      : '1px solid var(--border-color)',
                  boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                }}
              >
                {/* Report Item Top Bar */}
                <header
                  style={{
                    display: 'flex',
                    flexWrap: 'wrap',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    gap: '0.75rem',
                    marginBottom: '1rem',
                    paddingBottom: '0.75rem',
                    borderBottom: '1px solid var(--border-color)',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                    <span
                      style={{
                        fontFamily: 'monospace',
                        fontWeight: 700,
                        fontSize: '0.875rem',
                        color: 'var(--text-secondary)',
                      }}
                    >
                      #{report.id}
                    </span>
                    <Badge variant="name">
                      {item.current_status.toUpperCase()}
                    </Badge>
                    <span
                      style={{
                        padding: '0.2rem 0.6rem',
                        borderRadius: '0.375rem',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        backgroundColor: 'rgba(99, 102, 241, 0.15)',
                        color: 'var(--accent-primary, #6366f1)',
                      }}
                    >
                      Reason: {getReasonLabel(report.reason)}
                    </span>
                  </div>

                  <div style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>
                    Reported {formatDate(report.created_at)}
                  </div>
                </header>

                {/* Reported Content Container */}
                <div
                  style={{
                    padding: '1rem',
                    borderRadius: '0.5rem',
                    backgroundColor: 'var(--bg-primary)',
                    border: '1px solid var(--border-color)',
                    marginBottom: '1rem',
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      marginBottom: '0.5rem',
                      fontSize: '0.8125rem',
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--text-secondary)' }}>
                      <MessageSquare size={15} />
                      <span>Flagged Message:</span>
                      <strong style={{ color: 'var(--text-primary)' }}>
                        {msg?.conversation_name || 'Anonymous User'}
                      </strong>
                      {msg?.country && <span>{getCountryEmoji(msg.country)}</span>}
                    </div>

                    {report.thread_id && (
                      <Link
                        to={`/thread/${report.thread_id}`}
                        target="_blank"
                        style={{
                          fontSize: '0.75rem',
                          color: 'var(--accent-primary)',
                          display: 'flex',
                          alignItems: 'center',
                          gap: '0.25rem',
                          textDecoration: 'none',
                        }}
                      >
                        Open Thread <ExternalLink size={12} />
                      </Link>
                    )}
                  </div>

                  {/* Message body snippet */}
                  <div
                    style={{
                      fontSize: '0.9375rem',
                      color: item.current_status === 'removed' ? 'var(--text-muted)' : 'var(--text-primary)',
                      fontStyle: item.current_status === 'removed' ? 'italic' : 'normal',
                      textDecoration: item.current_status === 'removed' ? 'line-through' : 'none',
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                    }}
                  >
                    {msg?.content || report.details || '(No content text available)'}
                  </div>

                  {report.details && report.details !== msg?.content && (
                    <div
                      style={{
                        marginTop: '0.5rem',
                        padding: '0.4rem 0.6rem',
                        borderRadius: '0.375rem',
                        backgroundColor: 'rgba(255,255,255,0.03)',
                        fontSize: '0.8125rem',
                        color: 'var(--text-secondary)',
                      }}
                    >
                      <strong>Reporter detail note:</strong> {report.details}
                    </div>
                  )}
                </div>

                {/* History Log Toggle */}
                {historyList.length > 0 && (
                  <div style={{ marginBottom: '1rem' }}>
                    <button
                      onClick={() => toggleHistory(report.id)}
                      style={{
                        background: 'none',
                        border: 'none',
                        color: 'var(--accent-primary)',
                        cursor: 'pointer',
                        fontSize: '0.8125rem',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.35rem',
                        padding: 0,
                      }}
                    >
                      <History size={14} />
                      <span>{historyList.length} Moderation Action(s) Recorded</span>
                      {isHistoryExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                    </button>

                    {isHistoryExpanded && (
                      <div
                        style={{
                          marginTop: '0.5rem',
                          padding: '0.75rem',
                          borderRadius: '0.5rem',
                          backgroundColor: 'var(--bg-primary)',
                          border: '1px dashed var(--border-color)',
                          display: 'flex',
                          flexDirection: 'column',
                          gap: '0.5rem',
                        }}
                      >
                        {historyList.map((act) => (
                          <div key={act.id} style={{ fontSize: '0.78125rem', color: 'var(--text-secondary)' }}>
                            <strong style={{ color: 'var(--text-primary)' }}>{act.status.toUpperCase()}</strong> &bull;{' '}
                            {formatDate(act.created_at)}
                            {act.note && (
                              <div style={{ color: 'var(--text-muted)', fontStyle: 'italic', marginTop: '0.1rem' }}>
                                Note: "{act.note}"
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Moderator Decision Action Box */}
                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '0.75rem',
                    paddingTop: '0.75rem',
                    borderTop: '1px solid var(--border-color)',
                  }}
                >
                  <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                    <input
                      type="text"
                      placeholder="Add moderator note (optional)..."
                      value={notes[report.id] || ''}
                      onChange={(e) =>
                        setNotes((prev) => ({ ...prev, [report.id]: e.target.value }))
                      }
                      style={{
                        flex: 1,
                        minWidth: '220px',
                        padding: '0.45rem 0.75rem',
                        borderRadius: '0.375rem',
                        backgroundColor: 'var(--bg-primary)',
                        color: 'var(--text-primary)',
                        border: '1px solid var(--border-color)',
                        fontSize: '0.8125rem',
                      }}
                    />

                    <Button
                      size="sm"
                      variant="secondary"
                      disabled={isSubmitting || item.current_status === 'reviewed'}
                      onClick={() => handleAction(report.id, 'reviewed')}
                    >
                      <CheckCircle2 size={14} style={{ marginRight: '0.35rem' }} />
                      Mark Reviewed
                    </Button>

                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={isSubmitting || item.current_status === 'dismissed'}
                      onClick={() => handleAction(report.id, 'dismissed')}
                    >
                      <XCircle size={14} style={{ marginRight: '0.35rem' }} />
                      Dismiss
                    </Button>

                    <Button
                      size="sm"
                      disabled={isSubmitting || item.current_status === 'removed'}
                      style={{
                        backgroundColor: '#ef4444',
                        color: '#ffffff',
                      }}
                      onClick={() => handleAction(report.id, 'removed')}
                    >
                      <Trash2 size={14} style={{ marginRight: '0.35rem' }} />
                      Remove Content
                    </Button>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default AdminModeration;
