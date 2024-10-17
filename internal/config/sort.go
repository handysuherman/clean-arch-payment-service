package config

type sortPriorities struct {
	kp []int
}

func (sp *sortPriorities) Len() int { return len(sp.kp) }

func (sp *sortPriorities) Swap(i, j int) {
	sp.kp[i], sp.kp[j] = sp.kp[j], sp.kp[i]
}

func (sp *sortPriorities) Less(i, j int) bool {
	return sp.kp[i] < sp.kp[j]
}

type sortObservers struct {
	observers []ObserverWithPriority
}

func (so *sortObservers) Len() int { return len(so.observers) }

func (so *sortObservers) Swap(i, j int) {
	so.observers[i], so.observers[j] = so.observers[j], so.observers[i]
}

func (so *sortObservers) Less(i, j int) bool {
	return so.observers[i].Priority < so.observers[j].Priority
}
