package conductor

type Policy[T any] interface {
	Decide() (T, bool)
}

type TaggedPolicy[T any] interface {
	Decide(tag string) (T, bool)
}
