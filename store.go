package simgo

import (
	"sync"
)

type Store struct {
	sim   EventGenerator
	gets  []*StoreEvent
	puts  []*StoreEvent
	items []interface{}
	m     sync.Mutex
}

type StoreEvent struct {
	*Event
	Item interface{}
}

type EventGenerator interface {
	Event() *Event
}

func NewStore(sim EventGenerator) *Store {
	return &Store{sim: sim}
}

func (st *Store) Size() int {
	return len(st.items)
}

func (st *Store) Get() *StoreEvent {
	st.m.Lock()
	defer st.m.Unlock()

	ev := st.newStoreEvent()
	st.gets = append(st.gets, ev)

	st.triggerGets(true)

	return ev
}

func (st *Store) Put(item interface{}) *StoreEvent {
	st.m.Lock()
	defer st.m.Unlock()

	ev := st.newStoreEvent()

	st.puts = append(st.puts, ev)
	st.items = append(st.items, item)

	st.triggerPuts(true)

	return ev
}

func (st *Store) newStoreEvent() *StoreEvent {
	return &StoreEvent{Event: st.sim.Event()}
}

func (st *Store) triggerGets(triggerPuts bool) {
	for {
		triggered := false

		for len(st.gets) > 0 && len(st.items) > 0 {
			get := st.gets[0]
			st.gets = st.gets[1:]
			item := st.items[0]
			st.items = st.items[1:]

			if get.Aborted() {
				continue
			}

			get.Item = item
			get.Trigger()
			triggered = true
		}

		if triggered && triggerPuts {
			st.triggerPuts(false)
		} else {
			break
		}
	}
}

func (st *Store) triggerPuts(triggerGets bool) {
	for {
		triggered := false

		for len(st.puts) > 0 {
			put := st.puts[0]
			st.puts = st.puts[1:]

			if put.Aborted() {
				continue
			}

			put.Trigger()
			triggered = true
		}

		if triggered && triggerGets {
			st.triggerGets(false)
		} else {
			break
		}
	}
}
