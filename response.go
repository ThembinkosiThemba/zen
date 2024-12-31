package zen

// This package has neccessay response functions and helpers for structuring response
// data in api responses

// AppCode represents the status code of an application operation.
type AppCode int

// Constants representing different application status codes.
// These codes are used to indicate the result or state of an operation.
const (
	OK      AppCode = iota // OK represents a successful operation or result.
	Failure                // Failure represents a failed operation or result.
)

// Map is a shorthand for map[string]interface{} with additional helper methods
type M map[string]interface{}

// Response represents a standard data response structure
type Response struct {
	Status  int         `json:"status"`
	Success AppCode     `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Response creates a successful data response
func (c *Context) Success(status int, data interface{}, message string) {
	response := Response{
		Status:  status,
		Data:    data,
		Success: OK,
		Message: message,
	}
	c.JSON(status, response)
}

// Error creates an error response
func (c *Context) Error(status int, message string, details ...interface{}) {
	var response Response
	if len(details) > 0 {
		response = Response{
			Status:  status,
			Success: Failure,
			Data:    details[0],
			Message: message,
		}
	} else {
		response = Response{
			Status:  status,
			Success: Failure,
			Message: message,
		}
	}
	c.JSON(status, response)
}
