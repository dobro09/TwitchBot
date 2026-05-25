package store

import (
	"context"
	"sort"
	"twbot/internal/model"
)

type InMemoryStore struct{
	im []model.Message
}

func NewInMemoryStore()(*InMemoryStore, error){
	return &InMemoryStore{im: []model.Message{}}, nil
}

func (s *InMemoryStore) InsertMessage(ctx context.Context, msg model.Message) error{
	s.im = append(s.im, msg)
	return nil
}

func (s *InMemoryStore) TopUsers(ctx context.Context, channel string, limit int) ([]model.UserStat, error){
	countmap := make(map[string]int)//map[userid]count
	namemap := make(map[string]string)//map[userid]username
	for _, val := range s.im{
		if val.Channel == channel{
			countmap[val.UserID]++
			namemap[val.UserID]= val.UserName
		}
	}
	result := make([]model.UserStat, 0)
	for k, v := range countmap{
		var res model.UserStat
		res.UserID = k
		res.MessageCount = v
		res.UserName = namemap[res.UserID]
		result = append(result, res)
	}
	sort.Slice(result, func(i, j int) bool {return result[i].MessageCount > result[j].MessageCount})
	if len(result) < limit{
		return result, nil
	}
	return result[:limit], nil
}