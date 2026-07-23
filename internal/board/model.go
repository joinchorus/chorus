package board

import "time"

// Board represents a system-curated context domain for organizing discussions.
// Boards are immutable system entities that provide context, not community identity.
type Board struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// SystemBoards is the immutable list of curated project boards.
var SystemBoards = []*Board{
	{ID: "brd_technology", Slug: "technology", DisplayName: "Technology", Description: "General discussions about technology."},
	{ID: "brd_programming", Slug: "programming", DisplayName: "Programming", Description: "Software engineering, languages, tooling and architecture."},
	{ID: "brd_ai", Slug: "ai", DisplayName: "Artificial Intelligence", Description: "AI, machine learning, neural models and autonomous systems."},
	{ID: "brd_science", Slug: "science", DisplayName: "Science", Description: "Natural sciences, physics, biology, and scientific discoveries."},
	{ID: "brd_design", Slug: "design", DisplayName: "Design", Description: "Product design, UX, typography and visual systems."},
	{ID: "brd_philosophy", Slug: "philosophy", DisplayName: "Philosophy", Description: "Ethics, metaphysics, logic, and existential thought."},
	{ID: "brd_politics", Slug: "politics", DisplayName: "Politics", Description: "Political theory, governance, and public policy."},
	{ID: "brd_history", Slug: "history", DisplayName: "History", Description: "Historical events, eras, and historiography."},
	{ID: "brd_books", Slug: "books", DisplayName: "Books", Description: "Literature, prose, and reading."},
	{ID: "brd_movies", Slug: "movies", DisplayName: "Movies", Description: "Cinema, film theory, and filmmaking."},
	{ID: "brd_music", Slug: "music", DisplayName: "Music", Description: "Acoustics, composition, genres, and audio."},
	{ID: "brd_gaming", Slug: "gaming", DisplayName: "Gaming", Description: "Game design, mechanics, and interactive media."},
	{ID: "brd_cybersecurity", Slug: "cybersecurity", DisplayName: "Cybersecurity", Description: "Security, cryptography, and privacy engineering."},
	{ID: "brd_mathematics", Slug: "mathematics", DisplayName: "Mathematics", Description: "Pure and applied mathematics, proof, and computation."},
	{ID: "brd_engineering", Slug: "engineering", DisplayName: "Engineering", Description: "Systems, hardware, and physical engineering."},
	{ID: "brd_economics", Slug: "economics", DisplayName: "Economics", Description: "Markets, incentive design, and economic theory."},
	{ID: "brd_psychology", Slug: "psychology", DisplayName: "Psychology", Description: "Cognition, behavior, and mental processes."},
}

// GetBoardBySlug finds a board by its unique URL slug.
func GetBoardBySlug(slug string) *Board {
	for _, b := range SystemBoards {
		if b.Slug == slug {
			return b
		}
	}
	return nil
}
