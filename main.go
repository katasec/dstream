package main

func main() {

	dStream := NewServer()
	dStream.Start()

	// log.Println(config.Ingester.Tables[0].Name)
	// log.Println(config.Ingester.Tables[0].PollInterval)
	// log.Println(config.Ingester.Tables[0].MaxPollInterval)

	// log.Println(config.Ingester.Tables[1].Name)
	// log.Println(config.Ingester.Tables[1].PollInterval)
	// log.Println(config.Ingester.Tables[1].MaxPollInterval)

	// log.Println(config.Publisher.Output.ConnectionString)
	// log.Println(config.Ingester.Topic.ConnectionString)
}
