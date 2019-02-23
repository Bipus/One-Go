package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/crypto/bcrypt"
)

var (
	//ErrNorFound is returned when a resource cannot be found
	// in the database.
	ErrNorFaund = errors.New("models: resource not found")
	//ErrInvalidI is returned when  an invalid ID is provided
	// to a mathod like Delete.
	ErrInvalidID = errors.New("models: ID provided was invalid, must be > 0")
)

const userPwPepper = "secret-random-string" // любую страку написать для усложнения паролей

func NewUserService(connectionInfo string) (*UserService, error) {
	//Соединение с базой данных  !!ВАЖНО ?parseTime=true  добисывать в конце если надо чтобы выводило время
	db, err := gorm.Open("mysql", connectionInfo) //"root:alfadog1@/bipusdb?charset=utf8&parseTime=True&loc=Local"

	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &UserService{
		db: db,
	}, nil
}

type UserService struct {
	db *gorm.DB
}

//ByID will look up by the id provided
//if  the user  if found ,we will return a nill error
//if the user if not found ,we will return ErrRecordNotFound
//if these is another error ,we will return an error with
//more information about what went wrong
//
// as a general rule, any error but ErrNotFound should
//probably result in a 500 error
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ? ", id)
	err := first(db, &user)
	return &user, err
}

// ByEmail looks up a user with the given email address and
//return that user
//if  the user  if found ,we will return a nill error
//if the user if not found ,we will return ErrRecordNotFound
//if these is another error ,we will return an error with
//more information about what went wrong
//
// as a general rule, any error but ErrNotFound should
//probably result in a 500 error
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err

}

//first  will query using  the provided gorm.BD and it will
//get the first item returned and place it into dst. if
// nothing is found in the query , it will return ErrNotFound
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNorFaund
	}
	return err
}

//Create will create the provided user and backfill data
// like the ID, CreatedAt and UpdateAt fileds.
func (us *UserService) Create(user *User) error {
	pwBytes := []byte(user.Password + userPwPepper)                              // этим услажнили просто пароль
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost) // используется для хеширования пароля
	if err != nil {
		return err
	}
	user.PsswordHash = string(hashedBytes)
	user.Password = ""
	return us.db.Create(user).Error
}

//Update will update the provided user with all of the database
//in the  provaided user object
func (us *UserService) Update(user *User) error { return us.db.Save(user).Error }

//Delete will delete the user with the proveided ID
func (us *UserService) Delete(id *uint) error {
	if *id == 0 {
		return ErrInvalidID
	}
	user := User{
		Model: gorm.Model{
			ID: *id,
		},
	}

	return us.db.Delete(&user).Error
}

//Фнкция для closes the Service with database
func (us *UserService) Close() error { return us.db.Close() }

//DestructiveReset drops the user table and rebuilds it
func (us *UserService) DestructiveReset() error { // удалит таблицу если существует
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()

}

//AutoMigrate will attempt to autonatically migrate the
//user table
//Добовляет в базу данных нехватающих полей
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

type User struct {
	gorm.Model
	Name        string
	Email       string `gorm:"not null;unique_index"`
	Password    string `gorm: "-"`
	PsswordHash string `gorm:not null`
}
