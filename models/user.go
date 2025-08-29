package models

import (
	"errors"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/forms"

	"golang.org/x/crypto/bcrypt"
)

type UserLoginResponse struct {
	User    LegacyUser `json:"user"`
	Token   string     `json:"token"`
	Message string     `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// LegacyUser for backward compatibility with existing auth system
type LegacyUser struct {
	ID        int64  `db:"id, primarykey, autoincrement" json:"id"`
	Email     string `db:"email" json:"email"`
	Password  string `db:"password" json:"-"`
	Name      string `db:"name" json:"name"`
	UpdatedAt int64  `db:"updated_at" json:"-"`
	CreatedAt int64  `db:"created_at" json:"-"`
}

// UserModel ...
type UserModel struct{}

var authModel = new(AuthModel)

// Login ...
func (m UserModel) Login(form forms.LoginForm) (user LegacyUser, token Token, err error) {

	err = db.GetDB().SelectOne(&user, "SELECT id, email, password, name, updated_at, created_at FROM user WHERE email=LOWER(?) LIMIT 1", form.Email)

	if err != nil {
		return user, token, err
	}

	//Compare the password form and database if match
	bytePassword := []byte(form.Password)
	byteHashedPassword := []byte(user.Password)

	err = bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)

	if err != nil {
		return user, token, err
	}

	//Generate the JWT auth token
	tokenDetails, err := authModel.CreateToken(user.ID)
	if err != nil {
		return user, token, err
	}

	saveErr := authModel.CreateAuth(user.ID, tokenDetails)
	if saveErr == nil {
		token.AccessToken = tokenDetails.AccessToken
		token.RefreshToken = tokenDetails.RefreshToken
	}

	return user, token, nil
}

// Register ...
func (m UserModel) Register(form forms.RegisterForm) (user LegacyUser, err error) {
	getDb := db.GetDB()

	//Check if the user exists in database
	checkUser, err := getDb.SelectInt("SELECT count(id) FROM user WHERE email=LOWER(?) LIMIT 1", form.Email)
	if err != nil {
		return user, errors.New("something went wrong, please try again later")
	}

	if checkUser > 0 {
		return user, errors.New("email already exists")
	}

	bytePassword := []byte(form.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return user, errors.New("something went wrong, please try again later")
	}

	//Create the user and return back the user ID
	err = getDb.QueryRow("INSERT INTO user(email, password, name) VALUES(?, ?, ?) RETURNING id", form.Email, string(hashedPassword), form.Name).Scan(&user.ID)
	if err != nil {
		return user, errors.New("something went wrong, please try again later")
	}

	user.Name = form.Name
	user.Email = form.Email

	return user, err
}

// One ...
func (m UserModel) One(userID int64) (user LegacyUser, err error) {
	err = db.GetDB().SelectOne(&user, "SELECT id, email, name FROM user WHERE id=? LIMIT 1", userID)
	return user, err
}
