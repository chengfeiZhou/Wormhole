package structs

type Message interface {
	Unmarshal(b []byte) error
	Marshal() ([]byte, error)
}
type TransMessage func([]byte) (Message, error)
