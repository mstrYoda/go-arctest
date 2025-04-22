package customer

import _ "github.com/mstrYoda/go-arctest/examples/example_project/utils"

type Customer struct {
	ID   string
	Name string
}

func (c *Customer) GetName() string {
	return c.Name
}
