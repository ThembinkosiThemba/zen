# Zen Response Helpers Documentation

Zen provides built-in response helpers for creating consistent API responses across your application. It includes a lightweight map type and standardized response methods for both success and error cases.

## Table of Contents

- [Map Type (zen.M)](#map-type-zenm)
- [Response Structure](#response-structure)
- [Response Methods](#response-methods)
  - [Success Response (c.Success)](#success-response-csuccess)
  - [Error Response (c.Error)](#error-response-cerror)
- [AppCode Constants](#appcode-constants)
- [Complete Example](#complete-example)
- [Best Practices](#best-practices)

  - [Consistent Status Codes](#consistent-status-codes)
  - [Clear Messages](#clear-messages)
  - [Error Details](#error-details)
  - [Early Returns](#early-returns)
  - [Structured Data](#structured-data)

## Map Type (zen.M)

`M` is a convenient type alias for `map[string]interface{}` that allows for clean map creation.

```go
// Create a map directly
data := zen.M{
    "name": "John",
    "age": 30,
    "roles": []string{"admin", "user"},
}
```

## Response Structure

All API responses follow this JSON structure:

```json
{
  "status": 200, // HTTP status code
  "success": 0, // AppCode: 0 for OK, 1 for Failure
  "data": {}, // Response data (optional)
  "message": "Success" // Response message
}
```

## Response Methods

### Success Response (c.Success)

Used for sending successful responses with data.

```go
func (c *Context) Success(status int, data interface{}, message string)

// Example usage
app.GET("/users", func(c *zen.Context) {
    users := []User{...}
    c.Success(http.StatusOK, users, "Users retrieved successfully")
})

// Example with map data
app.GET("/profile", func(c *zen.Context) {
    c.Success(http.StatusOK, zen.M{
        "user": user,
        "settings": settings,
        "lastLogin": time.Now(),
    }, "Profile retrieved successfully")
})
```

### Error Response (c.Error)

Used for sending error responses with optional details.

```go
func (c *Context) Error(status int, message string, details ...interface{})

// Basic error
app.GET("/user/:id", func(c *zen.Context) {
    c.Error(http.StatusNotFound, "User not found")
})

// Error with details
app.POST("/users", func(c *zen.Context) {
    if err := validateUser(user); err != nil {
        c.Error(http.StatusBadRequest, "Validation failed", err)
        return
    }
})
```

## AppCode Constants

```go
const (
    OK AppCode = iota      // 0: Success
    Failure                // 1: Failure
)
```

## Complete Example

```go
app.POST("/users", func(c *zen.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.Error(http.StatusBadRequest, "Invalid request body", err)
        return
    }

    newUser, err := db.CreateUser(user)
    if err != nil {
        c.Error(http.StatusInternalServerError, "Failed to create user", err)
        return
    }

    c.Success(http.StatusCreated, zen.M{
        "id": newUser.ID,
        "created_at": newUser.CreatedAt,
        "status": "active",
    }, "User created successfully")
})
```

## Best Practices

1. **Consistent Status Codes**

   ```go
   // Use standard HTTP status codes
   c.Success(http.StatusOK, data, "Success")       // 200
   c.Success(http.StatusCreated, data, "Created")  // 201
   c.Error(http.StatusBadRequest, "Bad request")   // 400
   c.Error(http.StatusNotFound, "Not found")       // 404
   ```

2. **Clear Messages**

   ```go
   // Good
   c.Success(http.StatusOK, users, "Users retrieved successfully")

   // Bad
   c.Success(http.StatusOK, users, "Success")
   ```

3. **Error Details**

   ```go
   if err := validateInput(input); err != nil {
       c.Error(http.StatusBadRequest, "Validation failed", zen.M{
           "field": "email",
           "error": err.Error(),
       })
       return
   }
   ```

4. **Early Returns**

   ```go
   if err != nil {
       c.Error(http.StatusInternalServerError, "Database error", err)
       return  // Important: always return after sending an error response
   }
   ```

5. **Structured Data**
   ```go
   c.Success(http.StatusOK, zen.M{
       "user": user,
       "stats": zen.M{
           "visits": 100,
           "lastSeen": time.Now(),
       },
   }, "Profile data retrieved")
   ```

For additional details and updates, visit the [GitHub repository](https://github.com/ThembinkosiThemba/zen).
