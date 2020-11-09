package event

func GetBus() EventlogPubSub {
	return bus
}

func SetBus(newBus EventlogPubSub) {
	bus = newBus
}
