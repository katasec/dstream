# Test Database Setup

This package provides functionality to create and manage a test SQL Server database for DStream testing. It automatically sets up the database with Change Data Capture (CDC) enabled and creates the necessary tables for testing.

## Prerequisites

1. SQL Server instance with CDC capability
2. Environment variable setup:
   ```bash
   export DSTREAM_DB_CONNECTION_STRING="server=your-server;user id=your-user;password=your-password;database=master"
   ```

## What This Test Does

1. **Database Creation**
   - Creates a new database called `TestDB` if it doesn't exist
   - Enables CDC on the database

2. **Table Setup**
   - Creates two tables with CDC enabled:
     
     ### Persons Table
     ```sql
     CREATE TABLE [dbo].[Persons](
         [ID] [int] IDENTITY(1,1) NOT NULL,
         [FirstName] [varchar](100) NULL,
         [LastName] [varchar](100) NULL,
         PRIMARY KEY CLUSTERED ([ID] ASC)
     )
     ```

     ### Cars Table
     ```sql
     CREATE TABLE [dbo].[Cars](
         [CarID] [int] IDENTITY(1,1) NOT NULL,
         [BrandName] [nvarchar](50) NOT NULL,
         [Color] [nvarchar](30) NOT NULL,
         PRIMARY KEY CLUSTERED ([CarID] ASC)
     )
     ```

3. **Test Data**
   - Inserts sample records into both tables
   - Sample data includes:
     - Persons: John Doe, Jane Smith, Bob Johnson
     - Cars: Toyota (Red), Honda (Blue), Ford (Black)

## Running the Tests

```bash
# Run the test (this will create and setup the database)
go test ./test/testdb -v

# Run specific test
go test ./test/testdb -run TestDatabaseSetup -v
```

## Important Notes

1. The test database will persist between test runs to allow for repeated testing
2. CDC is automatically enabled on both tables
3. Each test run will:
   - Skip database creation if it already exists
   - Skip CDC enablement if already enabled
   - Skip table creation if they exist
   - Insert fresh test data each time

## Cleanup

If you need to remove the test database, you can uncomment the cleanup code in `setup_test.go`:
```go
defer testDB.Drop() // Clean up after test
```

## API Usage

```go
// Create a new test database
testDB, err := testdb.NewTestDB()

// Insert test data
err = testDB.InsertTestData()

// Reset test data (clear tables but keep structure)
err = testDB.Reset()

// Close connection
err = testDB.Close()

// Drop database (if needed)
err = testDB.Drop()
```
