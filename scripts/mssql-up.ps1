Write-Host "Starting MSSQL container..."

docker rm -f sqlserver 2>$null | Out-Null


"Root is $PSScriptRoot"

docker run --platform linux/amd64 `
  -e "ACCEPT_EULA=Y" `
  -e "MSSQL_SA_PASSWORD=Passw0rd123" `
  -p 1433:1433 `
  --name sqlserver `
  -v "$PSScriptRoot/mssql.conf:/var/opt/mssql/mssql.conf" `
  -d mcr.microsoft.com/mssql/server:2019-latest

Write-Host "MSSQL container started."
