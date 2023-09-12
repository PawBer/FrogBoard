package models

import (
	"github.com/PawBer/FrogBoard/internal/passwords"
	"github.com/doug-martin/goqu/v9"
)

type UserPermission int

const (
	Admin UserPermission = iota
	Moderator
)

type WrongPasswordError struct{}

func (e WrongPasswordError) Error() string {
	return "password doesn't match"
}

type User struct {
	Username    string
	Permission  UserPermission
	DisplayName string
}

type UserModel struct {
	DbConn *goqu.Database
}

func RegisterUser(username string) error {

	return nil
}

func (um *UserModel) Login(username, password string) (User, error) {
	query, params, _ := goqu.From("users").Select("password_hash", "display_name", "permission").Where(goqu.Ex{
		"username": username,
	}).ToSQL()

	var passwordHash, displayName string
	var permission UserPermission
	err := um.DbConn.QueryRow(query, params...).Scan(&passwordHash, &displayName, &permission)
	if err != nil {
		return User{}, err
	}

	valid := passwords.VerifyPassword(password, passwordHash)
	if !valid {
		return User{}, WrongPasswordError{}
	}

	return User{Username: username, Permission: permission, DisplayName: displayName}, nil
}
