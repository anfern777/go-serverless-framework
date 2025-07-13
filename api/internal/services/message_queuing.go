package services

type MessageQueuing[T any] interface {
	SendMessage(*T) error
}
