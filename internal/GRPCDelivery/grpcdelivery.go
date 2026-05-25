package grpcdelivery

// grpcServer := v1.NewGRPCServer(newstore)
// 	s:= grpc.NewServer()
// 	v1.RegisterBotStatsServer(s, grpcServer)
// 	lis, err:= net.Listen("tcp", "localhost:50051")
// 	if err != nil {
//     	log.Fatalf("не удалось запустить gRPC слушателя: %v", err)
// 	}
// 	go func ()  {
// 		if err:= s.Serve(lis); err !=nil{
// 			log.Fatalf("не удалось запустить gRPC слушателя: %v", err)
// 		}	
// 	}()
// 	go func() {
// 		<-ctx.Done()
//     	s.GracefulStop()
// 	}()