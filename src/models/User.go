package models

import (
	"html"
	"log"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint32    `gorm:"primary_key;auto_increment" json:"id"`
	Nickname  string    `gorm:"size:255;not null;unique" json:"nickname"`
	Email     string    `gorm:"size:100;not null;unique" json:"email"`
	Password  string    `gorm:"size:100;not null;" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func MakeHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (user *User) ProcessBeforeSave() error {
	hashedPassword, err := MakeHash(user.Password)
	if nil != err {
		return err
	}

	user.Password = string(hashedPassword)
	return nil
}

func (user *User) Initialize() {
	user.ID = 0
	user.Nickname = html.EscapeString(strings.TrimSpace(user.Nickname))
	user.Email = html.EscapeString(strings.TrimSpace(user.Email))
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
}

func (user *User) Validate() error {
	return nil
}

/*
	@Transaction
*/
func (user *User) SaveUser(db *gorm.DB) (*User, error) {

	tx := db.Begin()
	defer func() {
		if r := recover(); nil != r {
			tx.Rollback()
		}
	}()

	if err := tx.Error; nil != err {
		return user, err
	}

	if err := tx.Create(&user).Error; nil != err {
		tx.Rollback()
		return user, err
	}

	return user, tx.Commit().Error
}

func (user *User) FindAllUsers(db *gorm.DB) (*[]User, error) {

	users := []User{}
	if err := db.Debug().Model(&User{}).Limit(100).Find(&users).Error; nil != err {
		return &[]User{}, err
	}

	return &users, nil
}

/*
	@Transaction
*/
func (user *User) UpdateUser(db *gorm.DB, uid uint32) (*User, error) {
	err := user.ProcessBeforeSave()
	if err != nil {
		log.Fatal(err)
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); nil != r {
			tx.Rollback()
		}
	}()

	if err := tx.Error; nil != err {
		return user, err
	}

	tx.Model(&User{}).Where("id = ?", uid).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":  user.Password,
			"nickname":  user.Nickname,
			"email":     user.Email,
			"update_at": time.Now(),
		})

	if nil != tx.Error {
		tx.Rollback()
		return user, err
	}

	if err := db.Debug().Model(&User{}).Where("id = ?", uid).Take(&user).Error; nil != err {
		tx.Rollback()
		return user, err
	}

	return user, nil
}

/*
	@Transaction
*/
func (user *User) DeleteUser(db *gorm.DB, uid uint32) (int64, error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); nil != r {
			tx.Rollback()
		}
	}()

	db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).Delete(&User{})

	if db.Error != nil {
		tx.Rollback()
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
