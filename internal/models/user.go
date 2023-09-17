package models

import (
	"math/rand"

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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type User struct {
	Username    string
	DisplayName string
	Permission  UserPermission
}

type UserModel struct {
	DbConn *goqu.Database
}

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (um *UserModel) RegisterUser(username string, displayName string, permission UserPermission) (string, error) {
	password := randStr(10)
	passwordHash, err := passwords.GenerateHash(password)
	if err != nil {
		return "", err
	}

	query, params, _ := goqu.Insert("users").Rows(goqu.Record{
		"username":      username,
		"display_name":  displayName,
		"permission":    uint(permission),
		"password_hash": passwordHash,
	}).ToSQL()

	_, err = um.DbConn.Exec(query+" ON CONFLICT DO NOTHING", params...)
	if err != nil {
		return "", err
	}

	return password, nil
}

func (um *UserModel) ResetUserPassword(username string) (string, error) {
	password := randStr(10)
	passwordHash, err := passwords.GenerateHash(password)
	if err != nil {
		return "", err
	}

	sql, params, _ := goqu.Update("users").Set(goqu.Record{
		"password_hash": passwordHash,
	}).Where(goqu.Ex{"username": username}).ToSQL()

	_, err = um.DbConn.Exec(sql, params...)
	if err != nil {
		return "", err
	}

	return password, nil
}

func (um *UserModel) ChangeUserPassword(username, password string) error {
	passwordHash, err := passwords.GenerateHash(password)
	if err != nil {
		return err
	}

	sql, params, _ := goqu.Update("users").Set(goqu.Record{
		"password_hash": passwordHash,
	}).Where(goqu.Ex{"username": username}).ToSQL()

	_, err = um.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (um *UserModel) GetUsers() ([]User, error) {
	var users []User

	query, params, _ := goqu.From("users").Select("username", "display_name", "permission").ToSQL()

	rows, err := um.DbConn.Query(query, params...)
	if err != nil {
		return nil, err
	}

	var user User
	var permission int
	for rows.Next() {
		err := rows.Scan(&user.Username, &user.DisplayName, &permission)
		if err != nil {
			return nil, err
		}

		user.Permission = UserPermission(permission)

		users = append(users, user)
	}

	return users, nil
}

func (um *UserModel) GetUser(username string) (User, error) {
	var user User

	query, params, _ := goqu.From("users").Select("username", "display_name", "permission").Where(goqu.Ex{
		"username": username,
	}).ToSQL()

	row := um.DbConn.QueryRow(query, params...)

	var permission int
	err := row.Scan(&user.Username, &user.DisplayName, &permission)
	if err != nil {
		return User{}, err
	}

	user.Permission = UserPermission(permission)

	return user, nil
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

func (um *UserModel) Update(user User) error {
	sql, params, _ := goqu.Update("users").Set(goqu.Record{
		"display_name": user.DisplayName,
		"permission":   uint(user.Permission),
	}).Where(goqu.Ex{"username": user.Username}).ToSQL()

	_, err := um.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (um *UserModel) SetPassword(username, newPassword string) error {
	passwordHash, err := passwords.GenerateHash(newPassword)
	if err != nil {
		return err
	}

	sql, params, _ := goqu.Update("users").Set(goqu.Record{
		"password_hash": passwordHash,
	}).Where(goqu.Ex{"username": username}).ToSQL()

	_, err = um.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

func (um *UserModel) Delete(username string) error {
	sql, params, _ := goqu.Delete("users").Where(goqu.Ex{"username": username}).ToSQL()

	_, err := um.DbConn.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}
