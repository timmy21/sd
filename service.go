package sd

type Node struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

type Event struct {
	Nodes []Node
	Err   error
}

func (e Event) Copy() Event {
	nodes := make([]Node, len(e.Nodes))
	for i, node := range e.Nodes {
		nodes[i] = node
	}
	return Event{
		Nodes: nodes,
		Err:   e.Err,
	}
}

type Service interface {
	Subscribe(chan<- Event)
	Unsubscribe(chan<- Event)
	Stop() error
}
