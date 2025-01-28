package models

import (
	"database/sql"
	"log"
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

func GetPosts() ([]Post, error) {
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
            thread_start DESC, -- Order threads by root post time
            path ASC          -- Maintain reply hierarchy within thread
        LIMIT 100
    `
	rows, err := db.Query(query)
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
		}

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
