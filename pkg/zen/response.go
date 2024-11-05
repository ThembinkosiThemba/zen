package zen

// Map is a shorthand for map[string]interface{} with additional helper methods
type Map map[string]interface{}

// Response represents a standard data response structure
type Response struct {
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}

// ApiError represents an error response structure
type ApiError struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// M creates a new Map (shorthand for creating maps)
func M(values ...interface{}) Map {
	m := Map{}
	if len(values)%2 != 0 {
		return m
	}
	for i := 0; i < len(values); i += 2 {
		if key, ok := values[i].(string); ok {
			m[key] = values[i+1]
		}
	}
	return m
}

// R creates a successful data response
func R(v interface{}) Response {
	return Response{
		Data:    v,
		Success: true,
	}
}

// Err creates an error response
func Err(message string, code int) ApiError {
	return ApiError{
		Message: message,
		Code:    code,
	}
}

// ErrWithDetails creates an error response with additional details
func ErrWithDetails(message string, code int, details interface{}) ApiError {
	return ApiError{
		Message: message,
		Code:    code,
		Details: details,
	}
}

// Helper methods for Map
func (m Map) Set(key string, value interface{}) Map {
	m[key] = value
	return m
}

func (m Map) Get(key string) interface{} {
	return m[key]
}

func (m Map) GetString(key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func (m Map) GetInt(key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

// Helper methods for Response
func (r Response) WithMeta(meta interface{}) Response {
	return Response{
		Data:    r.Data,
		Success: r.Success,
	}
}

// Helper methods for ApiError
func (e ApiError) WithDetails(details interface{}) ApiError {
	e.Details = details
	return e
}
