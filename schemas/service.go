package schemas

type Service interface {
	String() string
	Close() error
}
