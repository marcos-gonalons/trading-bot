package api

type Data struct {
	orders    []*Order
	positions []*Position
	state     *State
}

// SetOrders ...
func (s *Data) SetOrders(orders []*Order) {
	s.orders = orders
}

// GetOrders ...
func (s *Data) GetOrders() []*Order {
	return s.orders
}

// SetPositions ...
func (s *Data) SetPositions(positions []*Position) {
	s.positions = positions
}

// GetPositions ...
func (s *Data) GetPositions() []*Position {
	return s.positions
}

// SetState ...
func (s *Data) SetState(state *State) {
	s.state = state
}

// GetState ...
func (s *Data) GetState() *State {
	return s.state
}
