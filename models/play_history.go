package models

import "time"

type PlayHistoryAggregate struct {
	Id 				[]int
	Name 			string
	Path			string
	PlayTimeTotal 	int
	PlayCountTotal 	int
	FirstPlayedTime time.Time
	LastPlayedTime 	time.Time
}

type PlayHistoryGranular struct {
	PlayTime	int
	StartTime	int
	UpdateTime	int
}

type PlayHistorySearchFilter struct {
	DisplayName	string
	SqlFilter	string
	FilterType	int
	PlayTime	int
}
