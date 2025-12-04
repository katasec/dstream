Please note that this is a design discussion so no code changes. Once we are happy with the design we can begin implementation.

Whenever we need to test dstream it needs some fuindamentals setup that I do manually everytime. I would like to setup go tests that facilitagte this for me if that makes sense. Normally I need to start an MS SQL Server. I use this script:


# Remove container if it exists
#docker ps -a | grep "sql2019" | cut -d" " -f1 | xargs docker rm -f
 
# Run new container
docker rm -f sqlserver || true
docker run --platform linux/amd64 -e 'ACCEPT_EULA=Y'\
            --name sql \
            -e 'MSSQL_SA_PASSWORD=Passw0rd123'\
            -p 1433:1433 --name sqlserver \
            -v ./mssql.conf:/var/opt/mssql/mssql.conf \
            -d mcr.microsoft.com/mssql/server:2019-latest

I use this file for the config as mssql.conf in the local system:

[sqlagent]
enabled = true

For a start - I'll provide more details - but want you to understand the above first

