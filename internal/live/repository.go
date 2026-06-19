package live

type Repository interface {
	Create(stream *Stream) error
	GetByID(id string) (*Stream, error)
	GetByStreamKey(streamKey string) (*Stream, error)
	List() ([]*Stream, error)
	Update(stream *Stream) error
}
