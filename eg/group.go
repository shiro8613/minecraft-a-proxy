package eg

type EGroup struct {
	start chan struct{}
	eCh   chan error
}

func New() *EGroup {
	return &EGroup{
		start: make(chan struct{}, 1),
		eCh:   make(chan error, 1),
	}
}

func (eg *EGroup) Go(f func() error) {
	go func() {
		<-eg.start
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()

		eg.eCh <- f()
	}()
}

func (eg *EGroup) Wait() error {
	close(eg.start)
	err := <-eg.eCh
	close(eg.eCh)
	return err
}
