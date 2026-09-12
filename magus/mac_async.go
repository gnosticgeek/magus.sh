package main

// asyncGeneration identifies one logical run of asynchronous work. Commands
// copy the value into their result messages and Update accepts a result only
// while it still matches the owning subsystem. This prevents canceled or slow
// work from publishing state after the user has navigated elsewhere.
type asyncGeneration uint64

// next invalidates every result issued by an earlier logical run and returns
// the identity that a newly scheduled command should carry.
func (g *asyncGeneration) next() asyncGeneration {
	(*g)++
	return *g
}

func (g asyncGeneration) current(candidate asyncGeneration) bool {
	return candidate != 0 && candidate == g
}
