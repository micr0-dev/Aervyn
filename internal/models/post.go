package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Username  string    `json:"-"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"attributedTo"`
	Author    Profile   `json:"author,omitempty"`
	CreatedAt time.Time `json:"published"`
	ReplyTo   *string   `json:"inReplyTo,omitempty"`
	URL       string    `json:"url"`
	IsLocal   bool      `json:"-"`

	ReplyDepth int   `json:"-"`
	ParentPost *Post `json:"-"`

	// Interaction counts
	LikeCount  int `json:"likes"`
	BoostCount int `json:"shares"`
	ReplyCount int `json:"replies"`

	// Current user's interactions
	HasLiked   bool `json:"hasLiked"`
	HasBoosted bool `json:"hasBoosted"`
}

func GetLocalTimeline() ([]Post, error) {
	query := `
        WITH RECURSIVE thread_posts AS (
            -- Get root posts (non-replies)
            SELECT 
                p.id, 
                p.user_id, 
                u.username, 
                p.content, 
                p.created_at, 
                p.reply_to,
                0 as depth,
                p.created_at as thread_start,
                p.id as root_id,
                CAST(printf('%020d', p.id) AS TEXT) as path
            FROM posts p
            JOIN users u ON p.user_id = u.id
            WHERE p.reply_to IS NULL
            
            UNION ALL
            
            -- Get replies recursively
            SELECT 
                p.id, 
                p.user_id, 
                u.username, 
                p.content, 
                p.created_at, 
                p.reply_to,
                tp.depth + 1,
                tp.thread_start,
                tp.root_id,
                tp.path || '.' || printf('%020d', p.id)
            FROM posts p
            JOIN users u ON p.user_id = u.id
            JOIN thread_posts tp ON p.reply_to = tp.id
        )
        SELECT 
            id, user_id, username, content, created_at, reply_to, depth,
            (SELECT COUNT(*) FROM likes WHERE post_id = thread_posts.id) as like_count,
            (SELECT COUNT(*) FROM boosts WHERE post_id = thread_posts.id) as boost_count,
            (SELECT COUNT(*) FROM posts WHERE reply_to = thread_posts.id) as reply_count
        FROM thread_posts
        ORDER BY 
            thread_start DESC,
            path ASC
        LIMIT 50
    `

	return getPostsFromQuery(query)
}

type PostWithContext struct {
	Post       Post
	ParentID   *string
	RootID     string
	ThreadTime time.Time
}

func GetFollowingTimeline(userID string) ([]Post, error) {
	// FIXME: Make this much better by actually sorting and timelining the posts correctly

	// First get local posts from followed users
	localPosts, err := getLocalFollowingPosts(userID)
	if err != nil {
		return nil, err
	}

	// Get list of remote users we follow
	remoteFollows, err := getRemoteFollows(userID)
	if err != nil {
		return nil, err
	}

	// Fetch remote posts
	var remotePosts []Post
	for _, follow := range remoteFollows {
		posts, err := fetchRemoteUserPosts(follow.Actor)
		if err != nil {
			log.Printf("Error fetching posts from %s: %v", follow.Actor, err)
			continue // Skip this user if there's an error
		}
		remotePosts = append(remotePosts, posts...)
	}

	// Merge all posts
	allPosts := append(localPosts, remotePosts...)

	// Create a map of posts by ID for quick lookup
	postMap := make(map[string]Post)
	for _, post := range allPosts {
		postMap[post.ID] = post
	}

	// Create posts with context and find root posts for threads
	var postsWithContext []PostWithContext
	for _, post := range allPosts {
		ctx := PostWithContext{
			Post:       post,
			ThreadTime: post.CreatedAt,
		}

		if post.ReplyTo != nil {
			ctx.ParentID = post.ReplyTo
			// Find the root post
			currentID := *post.ReplyTo
			for i := 0; i < 10; i++ { // Limit depth to prevent infinite loops
				if parent, ok := postMap[currentID]; ok {
					if parent.ReplyTo == nil {
						ctx.RootID = parent.ID
						ctx.ThreadTime = parent.CreatedAt
						break
					}
					currentID = *parent.ReplyTo
				} else {
					break
				}
			}
		} else {
			ctx.RootID = post.ID
		}

		postsWithContext = append(postsWithContext, ctx)
	}

	// Sort posts
	sort.Slice(postsWithContext, func(i, j int) bool {
		postI := postsWithContext[i]
		postJ := postsWithContext[j]

		// If one post is a reply and its parent is nearby (within last 10 posts)
		if postI.ParentID != nil {
			parentIndex := -1
			for k := j; k < len(postsWithContext) && k < j+10; k++ {
				if postsWithContext[k].Post.ID == *postI.ParentID {
					parentIndex = k
					break
				}
			}
			if parentIndex != -1 {
				return false // Keep reply after its parent
			}
		}

		if postJ.ParentID != nil {
			parentIndex := -1
			for k := i; k < len(postsWithContext) && k < i+10; k++ {
				if postsWithContext[k].Post.ID == *postJ.ParentID {
					parentIndex = k
					break
				}
			}
			if parentIndex != -1 {
				return true // Keep reply after its parent
			}
		}

		// If posts are in the same thread and close in time
		if postI.RootID == postJ.RootID {
			timeDiff := postI.Post.CreatedAt.Sub(postJ.Post.CreatedAt)
			if timeDiff.Hours() < 24 { // Within 24 hours
				// Sort by thread position
				return postI.Post.CreatedAt.Before(postJ.Post.CreatedAt)
			}
		}

		// Default to chronological order
		return postI.Post.CreatedAt.After(postJ.Post.CreatedAt)
	})

	// Convert back to regular posts
	sortedPosts := make([]Post, len(postsWithContext))
	for i, pc := range postsWithContext {
		sortedPosts[i] = pc.Post
	}

	return sortedPosts, nil
}

func getLocalFollowingPosts(userID string) ([]Post, error) {
	query := `
        WITH RECURSIVE thread_posts AS (
            SELECT 
                p.id, 
                p.user_id, 
                u.username, 
                p.content, 
                p.created_at, 
                p.reply_to,
                0 as depth,
                p.created_at as thread_start,
                p.id as root_id,
                CAST(printf('%020d', p.id) AS TEXT) as path
            FROM posts p
            JOIN users u ON p.user_id = u.id
            JOIN followers f ON p.user_id = f.actor
            WHERE p.reply_to IS NULL
            AND f.user_id = ?
            AND f.accepted = true
            
            UNION ALL
            
            SELECT 
                p.id, 
                p.user_id, 
                u.username, 
                p.content, 
                p.created_at, 
                p.reply_to,
                tp.depth + 1,
                tp.thread_start,
                tp.root_id,
                tp.path || '.' || printf('%020d', p.id)
            FROM posts p
            JOIN users u ON p.user_id = u.id
            JOIN thread_posts tp ON p.reply_to = tp.id
        )
        SELECT 
            id, user_id, username, content, created_at, reply_to, depth,
            (SELECT COUNT(*) FROM likes WHERE post_id = thread_posts.id) as like_count,
            (SELECT COUNT(*) FROM boosts WHERE post_id = thread_posts.id) as boost_count,
            (SELECT COUNT(*) FROM posts WHERE reply_to = thread_posts.id) as reply_count
        FROM thread_posts
        ORDER BY 
            thread_start DESC,
            path ASC
    `
	return getPostsFromQuery(query, userID)
}

func getRemoteFollows(userID string) ([]Follower, error) {
	rows, err := db.Query(`
        SELECT id, user_id, actor, accepted, created_at
        FROM followers
        WHERE user_id = ?
        AND accepted = true
        AND actor LIKE 'http%'
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var follows []Follower
	for rows.Next() {
		var f Follower
		err := rows.Scan(&f.ID, &f.UserID, &f.Actor, &f.Accepted, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		follows = append(follows, f)
	}
	return follows, nil
}

func fetchRemoteUserPosts(actorURI string) ([]Post, error) {
	// First fetch the user's outbox URL
	profile, err := FetchRemoteProfile(actorURI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}

	// Make request to outbox
	req, err := http.NewRequest("GET", profile.OutboxURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/activity+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the outbox response
	var outbox struct {
		First string            `json:"first"`
		Items []json.RawMessage `json:"orderedItems"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&outbox); err != nil {
		return nil, err
	}

	// If we got a "first" URL, fetch that page
	if outbox.First != "" && len(outbox.Items) == 0 {
		req, err = http.NewRequest("GET", outbox.First, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/activity+json")

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&outbox); err != nil {
			return nil, err
		}
	}

	var posts []Post
	replyMap := make(map[string]*Post) // Map to store posts by ID for reply linking

	for _, item := range outbox.Items {
		var activity struct {
			Type   string          `json:"type"`
			Object json.RawMessage `json:"object"`
		}
		if err := json.Unmarshal(item, &activity); err != nil {
			continue
		}

		var postContent struct {
			ID           string    `json:"id"`
			Type         string    `json:"type"`
			Content      string    `json:"content"`
			Published    time.Time `json:"published"`
			InReplyTo    *string   `json:"inReplyTo"`
			AttributedTo string    `json:"attributedTo"`
		}
		if activity.Type == "Create" {
			if err := json.Unmarshal(activity.Object, &postContent); err != nil {
				continue
			}
		} else {
			if err := json.Unmarshal(item, &postContent); err != nil {
				continue
			}
		}

		if postContent.Type != "Note" {
			continue
		}

		post := Post{
			ID:        postContent.ID,
			Content:   postContent.Content,
			CreatedAt: postContent.Published,
			Author:    *profile,
			IsLocal:   false,
			ReplyTo:   postContent.InReplyTo,
		}

		posts = append(posts, post)
		replyMap[post.ID] = &posts[len(posts)-1]
	}

	return posts, nil
}

func getPostsFromQuery(query string, args ...interface{}) ([]Post, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		var replyTo sql.NullString
		err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Author.Username,
			&p.Content,
			&p.CreatedAt,
			&replyTo,
			&p.ReplyDepth,
			&p.LikeCount,
			&p.BoostCount,
			&p.ReplyCount,
		)
		if err != nil {
			return nil, err
		}

		if replyTo.Valid {
			p.ReplyTo = &replyTo.String
		}

		author, err := GetProfileByID(p.AuthorID)
		if err != nil {
			return nil, err
		}
		p.Author = *author

		posts = append(posts, p)
	}

	for i := range posts {
		if posts[i].ReplyTo != nil {
			err = posts[i].LoadParentPost()
			if err != nil {
				return nil, err
			}
		}
	}

	return posts, nil
}

func GetPost(id string) (*Post, error) {
	query := `
        WITH RECURSIVE thread AS (
            -- Get root post
            SELECT 
                p.id, p.user_id, u.username, p.content, 
                p.created_at, p.reply_to,
                CASE WHEN p.reply_to IS NULL THEN 0
                     ELSE (
                         WITH RECURSIVE reply_depth AS (
                             SELECT id, reply_to, 1 as depth
                             FROM posts
                             WHERE id = p.reply_to
                             UNION ALL
                             SELECT p2.id, p2.reply_to, rd.depth + 1
                             FROM posts p2
                             JOIN reply_depth rd ON p2.id = rd.reply_to
                         )
                         SELECT MAX(depth)
                         FROM reply_depth
                     )
                END as depth
            FROM posts p
            JOIN users u ON p.user_id = u.id
            WHERE p.id = ?
        )
        SELECT 
            id, user_id, username, content, created_at, reply_to, depth,
            (SELECT COUNT(*) FROM likes WHERE post_id = thread.id) as like_count,
            (SELECT COUNT(*) FROM boosts WHERE post_id = thread.id) as boost_count,
            (SELECT COUNT(*) FROM posts WHERE reply_to = thread.id) as reply_count
        FROM thread
    `

	var p Post
	var replyTo sql.NullString

	err := db.QueryRow(query, id).Scan(
		&p.ID,
		&p.UserID,
		&p.Username,
		&p.Content,
		&p.CreatedAt,
		&replyTo,
		&p.ReplyDepth,
		&p.LikeCount,
		&p.BoostCount,
		&p.ReplyCount,
	)
	if err != nil {
		return nil, err
	}

	if replyTo.Valid {
		p.ReplyTo = &replyTo.String
		err := p.LoadParentPost()
		if err != nil {
			log.Printf("Error loading parent post: %v", err)
		}
	}

	return &p, nil
}

func GetPostsByUserID(userID string) ([]Post, error) {
	query := `
        WITH RECURSIVE thread_posts AS (
    -- Get root posts (non-replies) from the user
    SELECT 
        p.id, 
        p.user_id, 
        u.username, 
        p.content, 
        p.created_at, 
        p.reply_to,
        0 as depth,
        p.created_at as thread_start,
        p.id as root_id,
        CAST(printf('%020d', p.id) AS TEXT) as path
    FROM posts p
    JOIN users u ON p.user_id = u.id
    WHERE p.reply_to IS NULL 
    AND p.user_id = ?  -- Add this condition for user's root posts
    
    UNION ALL
    
    -- Get replies recursively (including replies to this user's posts)
    SELECT 
        p.id, 
        p.user_id, 
        u.username, 
        p.content, 
        p.created_at, 
        p.reply_to,
        tp.depth + 1,
        tp.thread_start,
        tp.root_id,
        tp.path || '.' || printf('%020d', p.id)
    FROM posts p
    JOIN users u ON p.user_id = u.id
    JOIN thread_posts tp ON p.reply_to = tp.id
)
SELECT 
    id, user_id, username, content, created_at, reply_to, depth,
    (SELECT COUNT(*) FROM likes WHERE post_id = thread_posts.id) as like_count,
    (SELECT COUNT(*) FROM boosts WHERE post_id = thread_posts.id) as boost_count,
    (SELECT COUNT(*) FROM posts WHERE reply_to = thread_posts.id) as reply_count
FROM thread_posts
WHERE 
    user_id = ? OR  -- Show user's posts
    root_id IN (    -- And complete threads of posts they replied to
        SELECT id FROM posts WHERE user_id = ?
    )
ORDER BY 
    thread_start DESC, -- Order threads by root post time
    path ASC          -- Maintain reply hierarchy within thread
LIMIT 100
    `

	rows, err := db.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		var replyTo sql.NullString
		err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Author.Username,
			&p.Content,
			&p.CreatedAt,
			&replyTo,
			&p.ReplyDepth,
			&p.LikeCount,
			&p.BoostCount,
			&p.ReplyCount,
		)
		if err != nil {
			return nil, err
		}

		if replyTo.Valid {
			p.ReplyTo = &replyTo.String
		}

		p.IsLocal = true
		posts = append(posts, p)
	}

	for i := range posts {
		if posts[i].ReplyTo != nil {
			err = posts[i].LoadParentPost()
			if err != nil {
				return nil, err
			}
		}
	}

	return posts, nil
}

func CreatePost(content string, userID string) (*Post, error) {
	id := uuid.New().String()
	now := time.Now()

	// First create the post
	_, err := db.Exec(
		"INSERT INTO posts (id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
		id, userID, content, now,
	)
	if err != nil {
		return nil, err
	}

	// Then get the username for the response
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		return nil, err
	}

	return &Post{
		ID:        id,
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: now,
	}, nil
}

func CreateReply(content string, replyTo string, userID string) (*Post, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.Exec(
		"INSERT INTO posts (id, user_id, content, created_at, reply_to) VALUES (?, ?, ?, ?, ?)",
		id, userID, content, now, replyTo,
	)
	if err != nil {
		return nil, err
	}

	post, err := GetPost(id)
	if err != nil {
		return nil, err
	}

	err = post.LoadParentPost()
	if err != nil {
		return nil, err
	}

	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		return nil, err
	}

	post.Username = username

	return post, nil
}

func LikePost(postID, userID string) error {
	id := uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO likes (id, post_id, user_id) VALUES (?, ?, ?)",
		id, postID, userID,
	)
	return err
}

func UnlikePost(postID, userID string) error {
	_, err := db.Exec(
		"DELETE FROM likes WHERE post_id = ? AND user_id = ?",
		postID, userID,
	)
	return err
}

func BoostPost(postID, userID string) error {
	id := uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO boosts (id, post_id, user_id) VALUES (?, ?, ?)",
		id, postID, userID,
	)
	return err
}

func UnboostPost(postID, userID string) error {
	_, err := db.Exec(
		"DELETE FROM boosts WHERE post_id = ? AND user_id = ?",
		postID, userID,
	)
	return err
}

func (p *Post) LoadUserInteractions(userID string) error {
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?)",
		p.ID, userID,
	).Scan(&p.HasLiked)
	if err != nil {
		return err
	}

	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM boosts WHERE post_id = ? AND user_id = ?)",
		p.ID, userID,
	).Scan(&p.HasBoosted)
	return err
}

func (p *Post) LoadParentPost() error {
	if p.ReplyTo == nil {
		return nil
	}

	parent, err := GetPost(*p.ReplyTo)
	if err != nil {
		return err
	}

	p.ParentPost = parent
	return nil
}
