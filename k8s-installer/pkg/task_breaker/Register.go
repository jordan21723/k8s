package task_breaker

type AddOnsRegister struct {
	AddOns []IAddOns
}

func (addOn *AddOnsRegister) RegisterUp(plugin IAddOns) {
	addOn.AddOns = append(addOn.AddOns, plugin)
}
