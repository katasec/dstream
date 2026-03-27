SET NOCOUNT ON;

IF DB_ID('TestDB') IS NULL
BEGIN
    CREATE DATABASE TestDB;
END;
GO

USE TestDB;
GO

IF NOT EXISTS (
    SELECT 1
FROM sys.databases
WHERE name = 'TestDB' AND is_cdc_enabled = 1
)
BEGIN
    EXEC sys.sp_cdc_enable_db;
END;
GO

IF OBJECT_ID('dbo.Persons', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.Persons
    (
        ID INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
        Name NVARCHAR(100) NOT NULL,
        Description NVARCHAR(255) NULL
    );
END;
GO

IF OBJECT_ID('dbo.Cars', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.Cars
    (
        ID INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
        Name NVARCHAR(100) NOT NULL,
        Description NVARCHAR(255) NULL
    );
END;
GO

IF NOT EXISTS (
    SELECT 1
FROM cdc.change_tables
WHERE source_object_id = OBJECT_ID('dbo.Persons')
)
BEGIN
    EXEC sys.sp_cdc_enable_table
        @source_schema = 'dbo',
        @source_name = 'Persons',
        @role_name = NULL,
        @supports_net_changes = 0;
END;
GO

IF NOT EXISTS (
    SELECT 1
FROM cdc.change_tables
WHERE source_object_id = OBJECT_ID('dbo.Cars')
)
BEGIN
    EXEC sys.sp_cdc_enable_table
        @source_schema = 'dbo',
        @source_name = 'Cars',
        @role_name = NULL,
        @supports_net_changes = 0;
END;
GO
