package v1

import (
	"context"
	"twbot/internal/store"
)

type GRPCServer struct{
	store store.MessageStore
	UnimplementedBotStatsServer
}

func (s *GRPCServer) GetTopUsers(ctx context.Context, req *GetRequest)(*GetResponse, error){
	res, err:=s.store.TopUsers(ctx, req.Channel, int(req.Limit))
	if err != nil {
        return nil, err
    }
	result := make([]*UserStat, 0, len(res))
	for _, val := range res{
		result=append(result, &UserStat{
			UserId: val.UserID,
			UserName: val.UserName,
			MessageCount: int32(val.MessageCount)})
	}
	return &GetResponse{Userstat: result}, nil
}

func NewGRPCServer(store store.MessageStore) *GRPCServer {
    return &GRPCServer{store: store}
}