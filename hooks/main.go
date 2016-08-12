package hooks

import "sort"

var Main *Hooks

func init() {
	Main = New()
}

func New() *Hooks {
	return &Hooks{}
}

type Hooks struct {
	actions           map[string]map[int][]func(...interface{})
	actionsPriorities map[string][]int
}

func (this *Hooks) AddAction(actionName string, f func(...interface{}), priority int) {
	if actionName != "" && f != nil {
		if this.actions == nil {
			this.actions = map[string]map[int][]func(...interface{}){}
		}
		if this.actionsPriorities == nil {
			this.actionsPriorities = map[string][]int{}
		}
		ok := false
		if _, ok = this.actions[actionName]; !ok {
			this.actions[actionName] = map[int][]func(...interface{}){}
			this.actionsPriorities[actionName] = []int{}
		}
		if _, ok = this.actions[actionName][priority]; !ok {
			this.actions[actionName][priority] = []func(...interface{}){}
			this.actionsPriorities[actionName] = append(this.actionsPriorities[actionName], priority)
		}
		this.actions[actionName][priority] = append(this.actions[actionName][priority], f)
	}
}

func (this *Hooks) DoAction(actionName string, params ...interface{}) {
	if actionName != "" {
		if actionPriorities, ok := this.actionsPriorities[actionName]; ok {
			if _, ok = this.actions[actionName]; ok {
				sort.Ints(actionPriorities)
				for _, actionPriority := range actionPriorities {
					if _, ok = this.actions[actionName][actionPriority]; ok {
						for _, actionFunc := range this.actions[actionName][actionPriority] {
							if actionFunc != nil {
								actionFunc(params...)
							}
						}
					}
				}
			}
		}
	}
}

func (this *Hooks) RemoveAction(actionName string) {
	if this.actions != nil {
		if _, ok := this.actions[actionName]; !ok {
			delete(this.actions, actionName)
		}
	}
}
