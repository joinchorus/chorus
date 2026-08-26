package gitstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"chorus/internal/domain"
	"chorus/internal/thread"
)

type threadRepository struct {
	mu    sync.RWMutex
	store *GitStore
}

// NewThreadRepository returns a Git-backed thread repository instance.
func NewThreadRepository(store *GitStore) (thread.Repository, error) {
	ctx := context.Background()
	if err := store.Init(ctx); err != nil {
		return nil, err
	}
	if err := store.RecoverAndVerify(ctx); err != nil {
		return nil, err
	}
	return &threadRepository{store: store}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (r *threadRepository) SaveThread(ctx context.Context, t *thread.Thread) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	boardSlug := t.BoardSlug
	if boardSlug == "" {
		boardSlug = "general"
	}

	relThreadDir := filepath.Join("boards", boardSlug, "threads", t.ID)
	fullThreadDir := filepath.Join(r.store.rootPath, relThreadDir)

	if err := os.MkdirAll(fullThreadDir, 0755); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	relThreadFile := filepath.Join(relThreadDir, "thread.json")
	fullThreadFile := filepath.Join(r.store.rootPath, relThreadFile)

	if _, err := os.Stat(fullThreadFile); err == nil {
		return domain.ErrAlreadyExists
	}

	// Sanitize thread metadata before writing to disk - NEVER write raw participant token to thread.json
	diskThread := *t
	diskThread.ParticipantToken = ""

	if err := WriteJSONFile(fullThreadFile, &diskThread); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	commitMsg := fmt.Sprintf("thread: create %s in %s", t.ID, boardSlug)
	if err := r.store.AddAndCommit(ctx, relThreadFile, commitMsg); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	// Synchronize board index
	_ = r.store.UpdateThreadInIndex(ctx, boardSlug, &diskThread, nil)
	return nil
}

func (r *threadRepository) FindThreadByID(ctx context.Context, id string) (*thread.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	relThreadDir, _, err := r.store.FindThreadRelDir(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	fullThreadFile := filepath.Join(r.store.rootPath, relThreadDir, "thread.json")
	t, err := ReadJSONFile[thread.Thread](fullThreadFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	// Ensure token is stripped on read
	if t != nil {
		t.ParticipantToken = ""
	}

	return t, nil
}

func (r *threadRepository) ListThreads(ctx context.Context, boardSlug ...string) ([]*thread.Thread, error) {
	return r.ListThreadsPaginated(ctx, 0, 100, boardSlug...)
}

func (r *threadRepository) ListThreadsPaginated(ctx context.Context, offset, limit int, boardSlug ...string) ([]*thread.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50
	}

	var allItems []ThreadIndexItem

	// Filter by specific board if provided
	if len(boardSlug) > 0 && boardSlug[0] != "" && boardSlug[0] != "all" {
		targetBoard := boardSlug[0]
		idx, err := r.store.ReadBoardIndex(targetBoard)
		if err == nil && idx != nil {
			allItems = idx.Threads
		}
	} else {
		// Read across all board directories
		boardSlugs, err := r.store.ListAllBoardSlugs()
		if err == nil {
			for _, slug := range boardSlugs {
				idx, err := r.store.ReadBoardIndex(slug)
				if err == nil && idx != nil {
					allItems = append(allItems, idx.Threads...)
				}
			}
		}
		// Sort aggregated items by LastMessageAt / CreatedAt descending
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].LastMessageAt.After(allItems[j].LastMessageAt)
		})
	}

	if offset >= len(allItems) {
		return []*thread.Thread{}, nil
	}

	end := offset + limit
	if end > len(allItems) {
		end = len(allItems)
	}

	slicedItems := allItems[offset:end]
	result := make([]*thread.Thread, 0, len(slicedItems))

	for _, item := range slicedItems {
		var fullThreadFile string
		if item.BoardSlug != "" {
			fullThreadFile = filepath.Join(r.store.rootPath, "boards", item.BoardSlug, "threads", item.ThreadID, "thread.json")
		}
		if fullThreadFile == "" || !fileExists(fullThreadFile) {
			relDir, _, err := r.store.FindThreadRelDir(item.ThreadID)
			if err == nil {
				fullThreadFile = filepath.Join(r.store.rootPath, relDir, "thread.json")
			}
		}
		if fullThreadFile != "" {
			t, err := ReadJSONFile[thread.Thread](fullThreadFile)
			if err == nil && t != nil {
				t.ParticipantToken = ""
				result = append(result, t)
			}
		}
	}

	return result, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (r *threadRepository) SaveMessage(ctx context.Context, m *thread.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	relThreadDir, boardSlug, err := r.store.FindThreadRelDir(m.ThreadID)
	if err != nil {
		return domain.ErrNotFound
	}

	// Sanitize message before writing to disk - NEVER write raw participant token to messages.ndjson
	diskMsg := *m
	diskMsg.ParticipantToken = ""

	fullMsgFile := filepath.Join(r.store.rootPath, relThreadDir, "messages.ndjson")
	if err := AppendNDJSONLine(fullMsgFile, &diskMsg); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	commitMsg := fmt.Sprintf("message: append %s", m.ID)
	relMsgFile := filepath.Join(relThreadDir, "messages.ndjson")
	if err := r.store.AddAndCommit(ctx, relMsgFile, commitMsg); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	// Update index with new message timestamp & count
	_ = r.store.IncrementMessageInIndex(ctx, boardSlug, m.ThreadID, m.CreatedAt)
	return nil
}

func (r *threadRepository) ListMessagesByThreadID(ctx context.Context, threadID string) ([]*thread.Message, error) {
	return r.ListMessagesPaginated(ctx, threadID, 0, 1000)
}

func (r *threadRepository) ListMessagesPaginated(ctx context.Context, threadID string, offset, limit int) ([]*thread.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	relThreadDir, _, err := r.store.FindThreadRelDir(threadID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	fullMsgFile := filepath.Join(r.store.rootPath, relThreadDir, "messages.ndjson")
	msgs, err := ReadNDJSONLinesPaginated[thread.Message](fullMsgFile, offset, limit)
	if err != nil {
		return nil, err
	}

	// Moderation Redaction: Check if any messages in this thread have been removed by moderation
	removedMsgMap := r.getRemovedMessageIDsLocked(relThreadDir)
	for _, m := range msgs {
		if removedMsgMap[m.ID] {
			m.Content = "[This message was removed by moderation]"
			m.IsRemoved = true
		}
		// Do not leak internal participant token on public message lists
		m.ParticipantToken = ""
	}

	return msgs, nil
}

func (r *threadRepository) getRemovedMessageIDsLocked(relThreadDir string) map[string]bool {
	removed := make(map[string]bool)
	modDir := filepath.Join(r.store.rootPath, relThreadDir, "moderation")
	entries, err := os.ReadDir(modDir)
	if err != nil {
		return removed
	}

	type modAct struct {
		MessageID string `json:"message_id"`
		Status    string `json:"status"`
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		fullPath := filepath.Join(modDir, entry.Name())
		act, err := ReadJSONFile[modAct](fullPath)
		if err == nil && act != nil {
			if act.Status == "removed" {
				removed[act.MessageID] = true
			} else if act.Status == "reviewed" || act.Status == "dismissed" {
				// If marked reviewed or dismissed after removal, unmark
				delete(removed, act.MessageID)
			}
		}
	}
	return removed
}

func (r *threadRepository) SaveParticipant(ctx context.Context, threadID string, p *thread.Participant) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	relThreadDir, _, err := r.store.FindThreadRelDir(threadID)
	if err != nil {
		return domain.ErrNotFound
	}

	fullPartFile := filepath.Join(r.store.rootPath, relThreadDir, "participants.json")
	participantsMap := make(map[string]*thread.Participant)

	if existing, err := ReadJSONFile[map[string]*thread.Participant](fullPartFile); err == nil && existing != nil {
		participantsMap = *existing
	}

	// Compute cryptographic SHA-256 hash of token - NEVER store raw token on disk
	tokenHash := hashToken(p.Token)
	storedRecord := &thread.Participant{
		TokenHash:        tokenHash,
		ConversationName: p.ConversationName,
		AuthorID:         p.AuthorID,
		CreatedAt:        p.CreatedAt,
	}

	participantsMap[tokenHash] = storedRecord

	if err := WriteJSONFile(fullPartFile, participantsMap); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	relPartFile := filepath.Join(relThreadDir, "participants.json")
	commitMsg := fmt.Sprintf("participant: register %s in %s", tokenHash[:8], threadID)
	_ = r.store.AddAndCommit(ctx, relPartFile, commitMsg)

	return nil
}

func (r *threadRepository) FindParticipantByToken(ctx context.Context, threadID, token string) (*thread.Participant, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	relThreadDir, _, err := r.store.FindThreadRelDir(threadID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	fullPartFile := filepath.Join(r.store.rootPath, relThreadDir, "participants.json")
	participantsMap, err := ReadJSONFile[map[string]*thread.Participant](fullPartFile)
	if err != nil || participantsMap == nil {
		return nil, domain.ErrNotFound
	}

	tokenHash := hashToken(token)
	p, ok := (*participantsMap)[tokenHash]
	if !ok || p == nil {
		// Backward compatibility: check if raw unhashed token was stored previously
		p, ok = (*participantsMap)[token]
	}
	if !ok || p == nil {
		return nil, domain.ErrNotFound
	}

	res := *p
	res.Token = token
	return &res, nil
}
