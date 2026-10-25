package zen

// M is a shortcut for map[string]any.
type M map[string]any

// Response is the envelope written by Context.Success and Context.Error.
type Response struct {
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// Success writes a successful JSON response envelope.
func (c *Context) Success(status int, data any, message string) {
	c.JSON(status, Response{Status: status, Success: true, Data: data, Message: message})
}

// Error writes a failed JSON response envelope. An optional detail value is
// placed in the data field.
func (c *Context) Error(status int, message string, details ...any) {
	response := Response{Status: status, Success: false, Message: message}
	if len(details) > 0 {
		response.Data = details[0]
	}
	c.JSON(status, response)
}
