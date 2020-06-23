package sd

type Instance struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

type Event struct {
	Instances []Instance
	Err       error
}

func (e Event) Copy() Event {
	instances := make([]Instance, len(e.Instances))
	for i, inst := range e.Instances {
		instances[i] = inst
	}
	return Event{
		Instances: instances,
		Err:       e.Err,
	}
}

type Service interface {
	Subscribe(chan<- Event)
	Unsubscribe(chan<- Event)
	Stop() error
}
