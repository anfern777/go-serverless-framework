package services

type AuthProviderAccessParams struct {
	UserpoolID *string
	Email      *string
}

type AuthProvider interface {
	CreateUser(authAccessParams AuthProviderAccessParams) (any, error)
	DeleteUser(authAccessParams AuthProviderAccessParams) error
}
