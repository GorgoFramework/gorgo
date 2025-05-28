# PostgreSQL Example

This example demonstrates how to use the SQL plugin with PostgreSQL database in Gorgo Framework.

## Prerequisites

1. **PostgreSQL Server**: Make sure PostgreSQL is installed and running
2. **Go 1.23+**: Ensure you have Go installed

## Setup

### 1. Database Setup

Create a database and table for testing:

```sql
-- Connect to PostgreSQL
psql -U postgres

-- Create database (if needed)
CREATE DATABASE postgres;

-- Connect to the database
\c postgres;

-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO users (username, email) VALUES 
    ('john_doe', 'john@example.com'),
    ('jane_smith', 'jane@example.com'),
    ('bob_wilson', 'bob@example.com');
```

### 2. Configuration

Update the `config/app.toml` file with your PostgreSQL connection details:

```toml
[app]
name = "postgres_example"
version = "0.0.4"
debug = true

[server]
host = "localhost"
port = 8080

[plugins.sql]
host = "localhost"
port = 5432
user = "postgres"
password = "your_password"
db = "postgres"
```

### 3. Install Dependencies

```bash
go mod tidy
```

## Running the Example

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### GET /
Returns API documentation and available endpoints.

**Response:**
```json
{
  "message": "PostgreSQL Example API",
  "endpoints": [
    "GET / - This documentation",
    "GET /users - Get all users",
    "GET /user/:id - Get user by ID"
  ],
  "note": "Make sure to create a 'users' table with id, username, and email columns"
}
```

### GET /users
Returns all users from the database.

**Response:**
```json
{
  "users": [
    {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com"
    },
    {
      "id": 2,
      "username": "jane_smith",
      "email": "jane@example.com"
    }
  ]
}
```

### GET /user/:id
Returns a specific user by ID.

**Example:** `GET /user/1`

**Response:**
```json
{
  "id": "1",
  "username": "john_doe",
  "email": "john@example.com"
}
```

**Error Response (User not found):**
```json
{
  "error": "User not found"
}
```

## Testing

You can test the API using curl:

```bash
# Get all users
curl http://localhost:8080/users

# Get specific user
curl http://localhost:8080/user/1

# Get API documentation
curl http://localhost:8080/
```

## Features Demonstrated

- **SQL Plugin Integration**: Shows how to add and configure the SQL plugin
- **Connection Pooling**: Uses pgxpool for efficient database connections
- **Error Handling**: Proper error handling for database operations
- **URL Parameters**: Demonstrates parameter extraction from URLs
- **JSON Responses**: Returns structured JSON responses
- **HTTP Status Codes**: Proper HTTP status codes for different scenarios

## Troubleshooting

### Connection Issues

1. **Connection refused**: Make sure PostgreSQL is running
2. **Authentication failed**: Check username/password in config
3. **Database not found**: Ensure the database exists
4. **Permission denied**: Check user permissions

### Common Errors

- **"Database service not available"**: The SQL plugin failed to initialize
- **"User not found"**: No user exists with the specified ID
- **"Database error"**: General database operation error

## Next Steps

This example can be extended with:

- User creation (POST /users)
- User updates (PUT /user/:id)
- User deletion (DELETE /user/:id)
- Authentication and authorization
- Input validation
- Pagination for user lists
- Database migrations 