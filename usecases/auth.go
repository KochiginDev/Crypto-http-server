package usecases

import (
	"context"
	"crypthttp/domain"
)

type Auth interface {
	//Register
	//При пустом имени или пароле - ErrInvalidInput
	//Если имя занято - ErrAlreadyExists
	Register(ctx context.Context, username string, password string) (token string, err error)
	//Login
	//При пустых полях и тд - ErrInvalidInput
	Login(ctx context.Context, username string, password string) (token string, err error)
	//CheckToken
	//неизвестный токен - ErrInvalidCredentials
	CheckToken(ctx context.Context, token string) (user domain.User, err error)
}
