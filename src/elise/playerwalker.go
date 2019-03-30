package elise

import (
	"github.com/yuhanfang/riot/apiclient"
)

// Walker is an interface to use for types that can
// walk through players
type Walker interface {
	GetGameIDList(AccountID int64) []int64
}

// PlayerWalker walks for players
type PlayerWalker struct {
	client *apiclient.Client
}

// GetGameIDList walks to get list of games
func (pw *PlayerWalker) GetGameIDList(AccountID int64) []int64 {
	return []int64{0}
}
