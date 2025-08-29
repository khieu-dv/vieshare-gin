package models

import (
	"errors"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/forms"
)

// Response represents the top-level structure
type AllArticleResponse struct {
	Results []Result `json:"results"`
}

type OneArticleResponse struct {
	Data ArticleResponse `json:"data"`
}

// Result represents each result item
type Result struct {
	Data []ArticleResponse `json:"data"`
	Meta Meta              `json:"meta"`
}

// Article represents an individual article
type ArticleResponse struct {
	ID        int          `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	UpdatedAt int64        `json:"updated_at"`
	CreatedAt int64        `json:"created_at"`
	User      UserResponse `json:"user"`
}

// User represents the article's author
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Meta represents metadata for pagination or total count
type Meta struct {
	Total int `json:"total"`
}

// Article ...
type Article struct {
	ID        int64    `db:"id, primarykey, autoincrement" json:"id"`
	UserID    int64    `db:"user_id" json:"-"`
	Title     string   `db:"title" json:"title"`
	Content   string   `db:"content" json:"content"`
	UpdatedAt int64    `db:"updated_at" json:"updated_at"`
	CreatedAt int64    `db:"created_at" json:"created_at"`
	User      *JSONRaw `db:"user" json:"user"`
}

// ArticleModel ...
type ArticleModel struct{}

// Create ...
func (m ArticleModel) Create(userID int64, form forms.CreateArticleForm) (articleID int64, err error) {
	err = db.GetDB().QueryRow("INSERT INTO article(user_id, title, content) VALUES(?, ?, ?) RETURNING id", userID, form.Title, form.Content).Scan(&articleID)
	return articleID, err
}

// One ...
func (m ArticleModel) One(userID, id int64) (article Article, err error) {
	err = db.GetDB().SelectOne(&article, "SELECT a.id, a.title, a.content, a.updated_at, a.created_at, json_object('id', u.id, 'name', u.name, 'email', u.email) AS user FROM article a LEFT JOIN user u ON a.user_id = u.id WHERE a.user_id=? AND a.id=? LIMIT 1", userID, id)
	return article, err
}

// All ...
func (m ArticleModel) All(userID int64) (articles []DataList, err error) {
	_, err = db.GetDB().Select(&articles, "SELECT COALESCE(json_group_array(json_object('id', a.id, 'title', a.title, 'content', a.content, 'updated_at', a.updated_at, 'created_at', a.created_at, 'user', json_object('id', u.id, 'name', u.name, 'email', u.email))), '[]') AS data, (SELECT json_object('total', count(a.id)) FROM article AS a WHERE a.user_id=? LIMIT 1) AS meta FROM article a LEFT JOIN user u ON a.user_id = u.id WHERE a.user_id=? ORDER BY a.id DESC", userID, userID)
	return articles, err
}

// Update ...
func (m ArticleModel) Update(userID int64, id int64, form forms.CreateArticleForm) (err error) {
	//METHOD 1
	//Check the article by ID using this way
	// _, err = m.One(userID, id)
	// if err != nil {
	// 	return err
	// }

	operation, err := db.GetDB().Exec("UPDATE article SET title=?, content=? WHERE id=?", form.Title, form.Content, id)
	if err != nil {
		return err
	}

	success, _ := operation.RowsAffected()
	if success == 0 {
		return errors.New("updated 0 records")
	}

	return err
}

// Delete ...
func (m ArticleModel) Delete(userID, id int64) (err error) {

	operation, err := db.GetDB().Exec("DELETE FROM article WHERE id=?", id)
	if err != nil {
		return err
	}

	success, _ := operation.RowsAffected()
	if success == 0 {
		return errors.New("no records were deleted")
	}

	return err
}
