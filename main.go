package main

func main() {

	dStream := NewServer()
	dStream.Start()

	// config := config.NewConfig()

	// duration := time.Duration(time.Second * 60)
	// locker, err := lockers.NewBlobLocker(config.Locks.ConnectionString, config.Locks.ContainerName, "deleteme", duration)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// locker.GetLockedTables()
	// // locker.AcquireLock()
	// // locker.ReleaseLock()

}
