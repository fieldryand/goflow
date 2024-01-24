package main

//func main() {
//	storeOptions := postgresql.Options{
//		ConnectionURL:      "postgres://postgres:example@0.0.0.0:5432/postgres?sslmode=disable",
//		TableName:          "Item",
//		MaxOpenConnections: 100,
//		Codec:              encoding.JSON,
//	}
//
//	client, err := postgresql.NewClient(storeOptions)
//	if err != nil {
//		panic(err)
//	}
//	defer client.Close()
//
//	options := goflow.Options{
//		Streaming:    true,
//		ShowExamples: true,
//		WithSeconds:  true,
//	}
//	gf := goflow.New(options)
//	gf.AttachStore(client)
//	gf.Use(goflow.DefaultLogger())
//	gf.Run(":8181")
//}
