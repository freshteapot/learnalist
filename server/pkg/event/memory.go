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

func (b *memoryBus) Subscribe(topic string, key string, fn interface{}) {
	listener := eventlogPubSubListener{
		topic: topic,
		key:   key,
		fn:    fn,
	}
	b.listeners = append(b.listeners, listener)
}

func (b *memoryBus) Publish(topic string, moment Eventlog) {
	for _, listener := range b.listeners {
		if listener.topic != topic {
			return
		}

		type HandlerType func(entry Eventlog)
		if f, ok := listener.fn.(func(entry Eventlog)); ok {
			HandlerType(f)(moment)
		}
	}
}

func (b *memoryBus) Start(topic string) {
}

func (b *memoryBus) Unsubscribe(topic string, key string) {
	for index, listener := range b.listeners {
		if listener.topic == topic && listener.key == key {
			b.listeners = append(b.listeners[:index], b.listeners[index+1:]...)
			break
		}
	}
}
