package tui

// NavStack manages a stack of views for navigation.
type NavStack struct {
	stack []View
}

// NewNavStack creates a new empty navigation stack.
func NewNavStack() *NavStack {
	return &NavStack{}
}

// Push adds a view to the top of the stack.
func (n *NavStack) Push(v View) {
	n.stack = append(n.stack, v)
}

// Pop removes and returns the top view. Returns nil if empty.
func (n *NavStack) Pop() View {
	if len(n.stack) == 0 {
		return nil
	}
	top := n.stack[len(n.stack)-1]
	n.stack = n.stack[:len(n.stack)-1]
	return top
}

// Current returns the top view without removing it. Returns nil if empty.
func (n *NavStack) Current() View {
	if len(n.stack) == 0 {
		return nil
	}
	return n.stack[len(n.stack)-1]
}

// Replace swaps the top view with a new one.
func (n *NavStack) Replace(v View) {
	if len(n.stack) == 0 {
		n.stack = append(n.stack, v)
		return
	}
	n.stack[len(n.stack)-1] = v
}

// Len returns the number of views in the stack.
func (n *NavStack) Len() int {
	return len(n.stack)
}
