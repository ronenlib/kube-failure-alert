package handler

// Handler handles controller events
type Handler interface {
	Handle(obj interface{}) error
}

// TmpHandler will be updated
type TmpHandler struct{}

// Handle will be updated
func (h TmpHandler) Handle(obj interface{}) error {
	println("notify")
	return nil
}
