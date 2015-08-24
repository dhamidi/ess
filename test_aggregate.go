package ess

// testAggregate is a dummy aggregate used for testing.  This type is
// included in the main package, because it is referenced by
// EventStoreTest.
type testAggregate struct {
	id     string
	events EventPublisher
	error  error

	onEvent   func(event *Event)
	onCommand func(*testAggregate)
}

func newTestAggregateFromCommand(command *Command) Aggregate {
	return newTestAggregate(command.Get("id").String())
}

func newTestAggregate(id string) *testAggregate {
	return &testAggregate{id: id}
}

func (self *testAggregate) FailWith(err error) *testAggregate {
	self.error = err
	return self
}

func (self *testAggregate) Id() string {
	return self.id
}

func (self *testAggregate) HandleEvent(e *Event) {
	if self.onEvent != nil {
		self.onEvent(e)
	}
}

func (self *testAggregate) HandleCommand(command *Command) error {
	if self.onCommand != nil {
		self.onCommand(self)
	}
	return self.error
}

func (self *testAggregate) PublishWith(publisher EventPublisher) Aggregate {
	self.events = publisher
	return self
}
