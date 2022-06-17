package internal

type JsonResponse[T any] struct {
	Error string
	Data  T
}