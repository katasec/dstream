SET NOCOUNT ON;

USE master;
GO

IF DB_ID('TestDB') IS NOT NULL
BEGIN
    ALTER DATABASE TestDB SET SINGLE_USER WITH ROLLBACK IMMEDIATE;
    DROP DATABASE TestDB;
END;
GO

CREATE DATABASE TestDB;
GO

USE TestDB;
GO

EXEC sys.sp_cdc_enable_db;
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
