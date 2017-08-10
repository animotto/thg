package thg

type Thg struct {
	Config
	SessionId string
	Profile ProfileData
}

type Config struct {
	IdPlayer uint32
	Password string
	Address string
	ReqUrl string
	HashSalt string
	AppVersion uint16
}

type AuthData struct {
	SessionId string
	RawData string
}

type NetMaintData struct {
	Nodes []NodeData
	Profile ProfileData
	Programs []ProgramData
	Queue []QueueData
	LogAttacks []LogAttackData
	RawData string
}

type NodeData struct {
	Id uint32
	Ntype byte
	Level byte
	UpgradeTimer int32
}

type ProgramData struct {
	Id uint32
	Ptype byte
	Level byte
	Amount uint32
	UpgradeTimer int32
}

type ProfileData struct {
	Id uint32
	Name string
	Gold uint32
	Bitcoins uint32
	Credits uint32
	Builders byte
	Rank uint32
	Experience uint32
	Country byte
}

type QueueData struct {
	Ptype byte
	Amount uint32
}

type LogAttackData struct {
	ReplayId uint32
	DateTime string
	AttackerId uint32
	TargetId uint32
	AttackerName string
	TargetName string
	AttackerCountry byte
	TargetCountry byte
	Gold uint32
	Bitcoins uint32
	Rank int32
}

type WorldData struct {
	Players []ProfileData
	Bonuses []BonusData
	Goals []GoalData
	RawData string
}

type BonusData struct {
	Id uint32
	Amount uint32
}

type GoalData struct {
	Id uint32
	Gtype byte
}

type GoalTypes struct {
	Gtypes map[byte]*GoalType
	RawData string
}

type GoalType struct {
	Name string
	Max byte
}

type ChatData struct {
	Messages []ChatMessage
	RawData string
}

type ChatMessage struct {
	DateTime string
	Name string
	Message string
	Id uint32
	Experience uint32
	Country uint16
}
