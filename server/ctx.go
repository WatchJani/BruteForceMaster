package server

type Ctx struct {
	Response
	header map[string]string
}

func (c *Ctx) GetHeader() map[string]string {
	return c.header
}
