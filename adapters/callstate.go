package adapters

/*
	Golang Says Interface name should name with er
 */
type CallStateAdapter interface {
	Get(key string) ([]byte, error)
	Del(key string) error
	Set(key string, state []byte, expire ...int) error
	GetMembersScore(key string) (map[string]int64, error)
	IncrKeyMemberScore(key string, member string, score int) (int64, error)
	DelKeyMember(key string, member string) error
	SetRecordingJob(state []byte) error
	KeyExist(key string) (bool, error)
	AddSetMember(key string, member string, expired ...int) error
}
