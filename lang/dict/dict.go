package dict

type IResponse interface {
	GetKana() string
}

type IDictionary[R IResponse] interface {
	Search(key string) (R, error)
}
