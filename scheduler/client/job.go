package bus

type Job interface {
	Execute(params string) (string, error)
}
