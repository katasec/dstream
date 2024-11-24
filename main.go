package main

func main() {

	dStream := NewServer()
	dStream.Start()

	//config := config.NewConfig()

	// duration := time.Duration(time.Second * 60)
	// locker, err := cdc.NewBlobLocker(config.Locks.ConnectionString, config.Locks.ContainerName, "testlock", duration)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// locker.AcquireLock()
	// locker.ReleaseLock()

}
