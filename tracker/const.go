package tracker

import "time"

const (
	stateLeader    = 0 //领导
	stateFollower  = 1 //选民
	stateCandidate = 2 //初始继承者
)

// WorkTime SyncData 服务器维护时间
const (
	WorkTime   = time.Second * 10
	LeaderTime = time.Second * 10 * 60
)
