package event

type memoryBus struct {
	listeners []eventlogPubSubListener
}

func NewMemoryBus() EventlogPubSub {
	return &memoryBus{}
}

func (b *memoryBus) Close() {
	b.listeners = make([]eventlogPubSubListener, 0)

}

func (b *memoryBus) Subscribe(key string, fn interface{}) {
	listener := eventlogPubSubListener{
		key: key,
		fn:  fn,
	}
	b.listeners = append(b.listeners, listener)
}

func (b *memoryBus) Publish(moment Eventlog) {
	for _, listener := range b.listeners {
		type HandlerType func(entry Eventlog)
		if f, ok := listener.fn.(func(entry Eventlog)); ok {
			HandlerType(f)(moment)
		}
	}
}

func (b *memoryBus) Start() {
}

func (b *memoryBus) Unsubscribe(key string) {
	for index, listener := range b.listeners {
		if listener.key == key {
			b.listeners = append(b.listeners[:index], b.listeners[index+1:]...)
			break
		}
	}
}
