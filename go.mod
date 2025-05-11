module github.com/katasec/dstream

go 1.24.2

require (
	github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus v1.8.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/hashicorp/hcl/v2 v2.23.0
	github.com/katasec/dstream-ingester-mssql v0.0.0-00010101000000-000000000000
	github.com/microsoft/go-mssqldb v1.8.0
	github.com/spf13/cobra v1.9.1
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.0 // indirect
	github.com/Azure/go-amqp v1.4.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/zclconf/go-cty v1.16.2 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
)

replace github.com/katasec/dstream-ingester-mssql => ../dstream-ingester-mssql
