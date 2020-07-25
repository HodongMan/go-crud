package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Board struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	Title     string    `gorm:"size:255;not null;unique" json:"title"`
	Content   string    `gorm:"size:255;not null;" json:"content"`
	Author    User      `json:"author"`
	AuthorID  uint32    `gorm:"not null" json:"author_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (board *Board) Initalize() {
	board.ID = 0
	board.Title = html.EscapeString(strings.TrimSpace(board.Title))
	board.Content = html.EscapeString(strings.TrimSpace(board.Content))
	board.Author = User{}
	board.CreatedAt = time.Now()
	board.UpdatedAt = time.Now()
}

func (board *Board) Validate() error {

	if "" == board.Title {
		return errors.New("Required Title")
	}
	if "" == board.Content {
		return errors.New("Required Content")
	}
	if board.AuthorID < 1 {
		return errors.New("Required Author")
	}

	return nil
}

/*
	@Transaction
*/
func (board *Board) SaveBoard(db *gorm.DB) (*Board, error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); nil != r {
			tx.Rollback()
		}
	}()

	if err := tx.Error; nil != err {
		return board, err
	}

	if err := tx.Model(&Board{}).Create(&board).Error; nil != err {
		tx.Rollback()
		return &Board{}, err
	}

	if 0 != board.ID {
		if err := tx.Debug().Model(&User{}).Where("id = ?", board.AuthorID).Take(&board.Author).Error; nil != err {
			tx.Rollback()
			return &Board{}, err
		}
	}

	return board, nil
}

func (board *Board) FindAllPosts(db *gorm.DB) (*[]Board, error) {
	var err error
	boards := []Board{}
	err = db.Debug().Model(&Board{}).Limit(100).Find(&boards).Error
	if err != nil {
		return &[]Board{}, err
	}

	return &boards, nil
}
