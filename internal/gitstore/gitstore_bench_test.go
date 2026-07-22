package gitstore_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"chorus/internal/gitstore"
	"chorus/internal/thread"
)

func BenchmarkCreateThread(b *testing.B) {
	tempDir := b.TempDir()
	store := gitstore.NewGitStore(tempDir)
	repo, err := gitstore.NewThreadRepository(store)
	if err != nil {
		b.Fatalf("failed initializing thread repo: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		th := &thread.Thread{
			ID:        fmt.Sprintf("thd_bench_%d", i),
			Title:     fmt.Sprintf("Benchmark Thread %d", i),
			AuthorID:  "usr_bench",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := repo.SaveThread(ctx, th); err != nil {
			b.Fatalf("SaveThread failed at iteration %d: %v", i, err)
		}
	}
}

func BenchmarkAppendMessage(b *testing.B) {
	tempDir := b.TempDir()
	store := gitstore.NewGitStore(tempDir)
	repo, err := gitstore.NewThreadRepository(store)
	if err != nil {
		b.Fatalf("failed initializing thread repo: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC()

	// Seed one thread to append messages to
	th := &thread.Thread{
		ID:        "thd_append_bench",
		Title:     "Append Benchmark Thread",
		AuthorID:  "usr_bench",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.SaveThread(ctx, th); err != nil {
		b.Fatalf("failed seeding thread: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := &thread.Message{
			ID:        fmt.Sprintf("msg_bench_%d", i),
			ThreadID:  "thd_append_bench",
			AuthorID:  "usr_bench",
			Content:   fmt.Sprintf("Benchmark message content line %d", i),
			CreatedAt: now,
		}
		if err := repo.SaveMessage(ctx, msg); err != nil {
			b.Fatalf("SaveMessage failed at iteration %d: %v", i, err)
		}
	}
}

func BenchmarkGetThread(b *testing.B) {
	tempDir := b.TempDir()
	store := gitstore.NewGitStore(tempDir)
	repo, err := gitstore.NewThreadRepository(store)
	if err != nil {
		b.Fatalf("failed initializing thread repo: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC()
	th := &thread.Thread{
		ID:        "thd_get_bench",
		Title:     "Get Thread Benchmark",
		AuthorID:  "usr_bench",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.SaveThread(ctx, th); err != nil {
		b.Fatalf("failed seeding thread: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.FindThreadByID(ctx, "thd_get_bench")
		if err != nil {
			b.Fatalf("FindThreadByID failed: %v", err)
		}
	}
}

func BenchmarkListThreads(b *testing.B) {
	tempDir := b.TempDir()
	store := gitstore.NewGitStore(tempDir)
	repo, err := gitstore.NewThreadRepository(store)
	if err != nil {
		b.Fatalf("failed initializing thread repo: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC()

	// Seed 20 threads into index
	for i := 0; i < 20; i++ {
		th := &thread.Thread{
			ID:        fmt.Sprintf("thd_seed_%d", i),
			Title:     fmt.Sprintf("Seed Thread %d", i),
			AuthorID:  "usr_bench",
			CreatedAt: now,
			UpdatedAt: now,
		}
		_ = repo.SaveThread(ctx, th)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		threads, err := repo.ListThreads(ctx)
		if err != nil || len(threads) == 0 {
			b.Fatalf("ListThreads failed: %v", err)
		}
	}
}

func BenchmarkListMessages(b *testing.B) {
	tempDir := b.TempDir()
	store := gitstore.NewGitStore(tempDir)
	repo, err := gitstore.NewThreadRepository(store)
	if err != nil {
		b.Fatalf("failed initializing thread repo: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC()

	th := &thread.Thread{
		ID:        "thd_list_msg_bench",
		Title:     "List Messages Benchmark",
		AuthorID:  "usr_bench",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = repo.SaveThread(ctx, th)

	// Seed 100 messages into NDJSON file
	for i := 0; i < 100; i++ {
		msg := &thread.Message{
			ID:        fmt.Sprintf("msg_seed_%d", i),
			ThreadID:  "thd_list_msg_bench",
			AuthorID:  "usr_bench",
			Content:   fmt.Sprintf("Seeded message content line %d", i),
			CreatedAt: now,
		}
		_ = repo.SaveMessage(ctx, msg)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msgs, err := repo.ListMessagesByThreadID(ctx, "thd_list_msg_bench")
		if err != nil || len(msgs) == 0 {
			b.Fatalf("ListMessagesByThreadID failed: %v", err)
		}
	}
}
