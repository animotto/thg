package thg

import (
	"crypto/md5"
	"encoding/base64"
	"strings"
	"fmt"
	"net/http"
	neturl "net/url"
	"io/ioutil"
	"errors"
	"strconv"
)

const (
	cmdAuthIdPassword = "auth_id_password"
	cmdCheckConn = "check_connectivity"
	cmdGetNetMaint = "net_get_for_maintenance"
	cmdGetWorld = "player_get_world"
	cmdBonusCollect = "bonus_collect"
	cmdGoalUpdate = "goal_update"
	cmdGetGoalTypes = "goal_types_get_list"
	cmdChatDisplay = "chat_display"
	cmdNetUpdate = "net_update"
	cmdChatSend = "chat_send"
	cmdNodeCreate = "create_node_and_update_net"
	cmdNodeDelete = "node_delete_net_update"
	cmdNodeUpgrade = "upgrade_node"
	cmdNodeSetBuilders = "node_set_builders"
	cmdNodeCollect = "collect"
	cmdProgramUpgrade = "upgrade_program"
)

func New(config Config) Thg {
	return Thg{
		Config: config,
	}
}

func (thg *Thg) HttpGet(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", errors.New("status code: " + fmt.Sprint(res.StatusCode))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func hashUrl(url, salt string) string {
	offset := 10
	if len(url) < offset {
		offset = len(url)
	}

	url = url[:offset] + salt + url[offset:]
	md5Hash := md5.Sum([]byte(url))
	hash := base64.StdEncoding.EncodeToString(md5Hash[2:8])
	hash = strings.Replace(hash, "=", ".", -1)
	hash = strings.Replace(hash, "+", "-", -1)
	hash = strings.Replace(hash, "/", "_", -1)
	return hash
}

func ParseDataSc(data string) [][]string {
	var ret [][]string
	d1 := strings.Split(data, ";")
	ret = make([][]string, len(d1))
	if d1[len(d1)-1] == "" {
		d1 = d1[:len(d1)-1]
	}
	for i1, v1 := range d1 {
		d2 := strings.Split(v1, ",")
		ret[i1] = make([]string, len(d2))
		for i2, v2 := range d2 {
			ret[i1][i2] = v2
		}
	}
	return ret
}

func ParseDataDog(data string) [][][]string {
	var ret [][][]string
	d1 := strings.Split(data, "@")
	ret = make([][][]string, len(d1))
	for i1, v1 := range d1 {
		d2 := strings.Split(v1, ";")
		if d2[len(d2)-1] == "" {
			d2 = d2[:len(d2)-1]
		}
		ret[i1] = make([][]string, len(d2))
		for i2, v2 := range d2 {
			d3 := strings.Split(v2, ",")
			ret[i1][i2] = make([]string, len(d3))
			for i3, v3 := range d3 {
				ret[i1][i2][i3] = v3
			}
		}
	}
	return ret
}

func GetNormalSymbols(s string) string {
	s = strings.Replace(s, "\x01", ",", -1)
	s = strings.Replace(s, "\x02", ";", -1)
	s = strings.Replace(s, "\x03", "@", -1)
	return s
}

func (thg *Thg) addHashToUrl(url string) string {
	return url + "&cmd_id=" + hashUrl(url, thg.HashSalt)
}

func (thg *Thg) GetFullUrl(url string) string {
	req := thg.ReqUrl + "?" + neturl.PathEscape(url)
	return "https:/" + "/" + thg.Address + thg.addHashToUrl(req) // doh! my editor syntax highlight is dumb, fix this concatenate strings later!
}

func (thg *Thg) GetConfig() map[string]string {
	ret := map[string]string{
		"address": thg.Address,
		"requrl": thg.ReqUrl,
		"hashshalt": thg.HashSalt,
		"appversion": fmt.Sprint(thg.AppVersion),
		"idplayer": fmt.Sprint(thg.IdPlayer),
	}

	return ret
}

// Command: authenticate client
func (thg *Thg) Auth() (AuthData, error) {
	var ret AuthData
	req := cmdAuthIdPassword +
		"&id_player=" + fmt.Sprint(thg.IdPlayer) +
		"&password=" + thg.Password +
		"&app_version=" + fmt.Sprint(thg.AppVersion)

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return ret, err
	}

	data := ParseDataDog(res)
	ret.RawData = res
	ret.SessionId = data[0][0][3]
	thg.SessionId = ret.SessionId
	return ret, nil
}

// Command: checking connection
func (thg *Thg) CheckConn() (error) {
	req := cmdCheckConn + "=1" +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	if res != "1" {
		return errors.New(cmdCheckConn + ": invalid response")
	}

	return nil
}

// Command: get my network information
func (thg *Thg) NetMaint() (NetMaintData, error) {
	var ret NetMaintData
	req := cmdGetNetMaint + "=1" +
		"&id_player=" + fmt.Sprint(thg.IdPlayer) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return ret, err
	}

	data := ParseDataDog(res)
	ret.RawData = res

	if v, e := strconv.Atoi(data[2][0][0]); e == nil {
		ret.Profile.Id = uint32(v)
	}
	ret.Profile.Name = data[2][0][1]
	if v, e := strconv.Atoi(data[2][0][2]); e == nil {
		ret.Profile.Gold = uint32(v)
	}
	if v, e := strconv.Atoi(data[2][0][3]); e == nil {
		ret.Profile.Bitcoins = uint32(v)
	}
	if v, e := strconv.Atoi(data[2][0][4]); e == nil {
		ret.Profile.Credits = uint32(v)
	}
	if v, e := strconv.Atoi(data[2][0][5]); e == nil {
		ret.Profile.Experience = uint32(v)
	}
	if v, e := strconv.Atoi(data[2][0][9]); e == nil {
		ret.Profile.Rank = uint32(v)
	}
	if v, e := strconv.Atoi(data[2][0][10]); e == nil {
		ret.Profile.Builders = byte(v)
	}
	if v, e := strconv.Atoi(data[2][0][13]); e == nil {
	ret.Profile.Country = byte(v)
	}

	thg.Profile = ret.Profile

	ret.Nodes = make([]NodeData, len(data[0]))
	for i := range data[0] {
		if v, e := strconv.Atoi(data[0][i][0]); e == nil {
			ret.Nodes[i].Id = uint32(v)
		}
		if v, e := strconv.Atoi(data[0][i][2]); e == nil {
			ret.Nodes[i].Ntype = byte(v)
		}
		if v, e := strconv.Atoi(data[0][i][3]); e == nil {
			ret.Nodes[i].Level = byte(v)
		}
		if v, e := strconv.Atoi(data[0][i][4]); e == nil {
			ret.Nodes[i].UpgradeTimer = int32(v)
		}
	}

	ret.Programs = make([]ProgramData, len(data[3]))
	for i := range data[3] {
		if v, e := strconv.Atoi(data[3][i][0]); e == nil {
			ret.Programs[i].Id = uint32(v)
		}
		if v, e := strconv.Atoi(data[3][i][2]); e == nil {
			ret.Programs[i].Ptype = byte(v)
		}
		if v, e := strconv.Atoi(data[3][i][3]); e == nil {
			ret.Programs[i].Level = byte(v)
		}
		if v, e := strconv.Atoi(data[3][i][4]); e == nil {
			ret.Programs[i].Amount = uint32(v)
		}
		if v, e := strconv.Atoi(data[3][i][5]); e == nil {
			ret.Programs[i].UpgradeTimer = int32(v)
		}
	}

	ret.Queue = make([]QueueData, len(data[4]))
	for i := range data[4] {
		if v, e := strconv.Atoi(data[4][i][0]); e == nil {
			ret.Queue[i].Ptype = byte(v)
		}
		if v, e := strconv.Atoi(data[4][i][1]); e == nil {
			ret.Queue[i].Amount = uint32(v)
		}
	}

	ret.LogAttacks = make([]LogAttackData, len(data[9]))
	for i := range data[9] {
		if v, e := strconv.Atoi(data[9][i][0]); e == nil {
			ret.LogAttacks[i].ReplayId = uint32(v)
		}
		ret.LogAttacks[i].DateTime = data[9][i][1]
		if v, e := strconv.Atoi(data[9][i][2]); e == nil {
			ret.LogAttacks[i].AttackerId = uint32(v)
		}
		if v, e := strconv.Atoi(data[9][i][3]); e == nil {
			ret.LogAttacks[i].TargetId = uint32(v)
		}
		if v, e := strconv.Atoi(data[9][i][4]); e == nil {
			ret.LogAttacks[i].Gold = uint32(v)
		}
		if v, e := strconv.Atoi(data[9][i][5]); e == nil {
			ret.LogAttacks[i].Bitcoins = uint32(v)
		}
		ret.LogAttacks[i].AttackerName = data[9][i][9]
		ret.LogAttacks[i].TargetName = data[9][i][10]
		if v, e := strconv.Atoi(data[9][i][11]); e == nil {
			ret.LogAttacks[i].AttackerCountry = byte(v)
		}
		if v, e := strconv.Atoi(data[9][i][12]); e == nil {
			ret.LogAttacks[i].TargetCountry = byte(v)
		}
		if v, e := strconv.Atoi(data[9][i][13]); e == nil {
			ret.LogAttacks[i].Rank = int32(v)
		}
	}

	return ret, nil
}

// Command: get world information
func (thg *Thg) GetWorld() (WorldData, error) {
	var ret WorldData
	req := cmdGetWorld + "=1" +
		"&id=" + fmt.Sprint(thg.IdPlayer) +
		"&id_country=" + fmt.Sprint(thg.Profile.Country) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return ret, err
	}

	data := ParseDataDog(res)
	ret.RawData = res

	ret.Players = make([]ProfileData, len(data[0]))
	for i := range data[0] {
		if v, e := strconv.Atoi(data[0][i][0]); e == nil {
			ret.Players[i].Id = uint32(v)
		}
		ret.Players[i].Name = data[0][i][1]
	}

	ret.Bonuses = make([]BonusData, len(data[1]))
	for i := range data[1] {
		if v, e := strconv.Atoi(data[1][i][0]); e == nil {
			ret.Bonuses[i].Id = uint32(v)
		}
		if v, e := strconv.Atoi(data[1][i][2]); e == nil {
			ret.Bonuses[i].Amount = uint32(v)
		}
	}

	ret.Goals = make([]GoalData, len(data[4]))
	for i := range data[4] {
		if v, e := strconv.Atoi(data[4][i][0]); e == nil {
			ret.Goals[i].Id = uint32(v)
		}
		if v, e := strconv.Atoi(data[4][i][2]); e == nil {
			ret.Goals[i].Gtype = byte(v)
		}
	}


	return ret, nil
}

// Command: collect bonus
func (thg *Thg) BonusCollect(id uint32) error {
	req := cmdBonusCollect + "=1" +
		"&id=" + fmt.Sprint(id) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	if res != "ok" {
		return errors.New(cmdBonusCollect + ": invalid response")
	}

	return nil
}

// Command: get goal types
func (thg *Thg) GetGoalTypes() (GoalTypes, error) {
	var ret GoalTypes
	req := cmdGetGoalTypes +
		"&app_version=" + fmt.Sprint(thg.AppVersion)

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return ret, err
	}

	data := ParseDataDog(res)
	ret.RawData = res

	ret.Gtypes = make(map[byte]*GoalType, len(data[0]))
	for i := range data[0] {
		if v, e := strconv.Atoi(data[0][i][0]); e == nil {
			id := byte(v)
			ret.Gtypes[id] = &GoalType{Name: data[0][i][1]}
			if v, e := strconv.Atoi(data[0][i][2]); e == nil {
				ret.Gtypes[id].Max = byte(v)
			}
		}
	}

	return ret, nil
}

// Command: goal update
func (thg *Thg) GoalUpdate(id uint32, record byte) error {
	req := cmdGoalUpdate +
		"&id=" + fmt.Sprint(id) +
		"&record=" + fmt.Sprint(record) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	data := ParseDataDog(res)
	if data[0][0][0] != "finished" {
		return errors.New(cmdGoalUpdate + ": invalid response")
	}

	return nil
}

// Command: get chat messages
func (thg *Thg) ChatDisplay(room uint16, lastMessage string) (ChatData, error) {
	var ret ChatData
	if len(lastMessage) == 0 {
		lastMessage = "0000-00-00 00:00:00.0000"
	}
	req := cmdChatDisplay +
		"&room=" + fmt.Sprint(room) +
		"&last_message=" + lastMessage +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return ret, err
	}

	data := ParseDataDog(res)
	ret.RawData = res

	ret.Messages = make([]ChatMessage, len(data[0]))
	for i := range data[0] {
		ir := len(data[0]) - i - 1
		ret.Messages[ir].DateTime = data[0][i][0]
		ret.Messages[ir].Name = data[0][i][1]
		ret.Messages[ir].Message = GetNormalSymbols(data[0][i][2])
		if v, e := strconv.Atoi(data[0][i][3]); e == nil {
			ret.Messages[ir].Id = uint32(v)
		}
		if v, e := strconv.Atoi(data[0][i][4]); e == nil {
			ret.Messages[ir].Experience = uint32(v)
		}
		if v, e := strconv.Atoi(data[0][i][6]); e == nil {
			ret.Messages[ir].Country = uint16(v)
		}
	}

	return ret, nil
}

// Command: send message to the chat
func (thg *Thg) ChatSend(room uint16, message string, lastMessage string) (ChatData, error) {
	var ret ChatData
	if len(lastMessage) == 0 {
		lastMessage = "0000-00-00 00:00:00.0000"
	}
	req := cmdChatSend +
		"&room=" + fmt.Sprint(room) +
		"&last_message=" + lastMessage +
		"&message=" + message +
		"&id_player=" + fmt.Sprint(thg.IdPlayer) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return ret, err
	}

	data := ParseDataDog(res)
	ret.RawData = res

	ret.Messages = make([]ChatMessage, len(data[0]))
	for i := range data[0] {
		ir := len(data[0]) - i - 1
		ret.Messages[ir].DateTime = data[0][i][0]
		ret.Messages[ir].Name = data[0][i][1]
		ret.Messages[ir].Message = GetNormalSymbols(data[0][i][2])
		if v, e := strconv.Atoi(data[0][i][3]); e == nil {
			ret.Messages[ir].Id = uint32(v)
		}
		if v, e := strconv.Atoi(data[0][i][4]); e == nil {
			ret.Messages[ir].Experience = uint32(v)
		}
		if v, e := strconv.Atoi(data[0][i][6]); e == nil {
			ret.Messages[ir].Country = uint16(v)
		}
	}

	return ret, nil
}

// Command: update network topology
func (thg *Thg) NetUpdate(net string) error {
	req := cmdNetUpdate +
		"&id_player=" + fmt.Sprint(thg.IdPlayer) +
		"&net=" + net +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}
	if res != "ok" {
		return errors.New(cmdNetUpdate + ": invalid response")
	}

	return nil
}

func (thg *Thg) NodeCreate(ntype byte, net string) error {
	req := cmdNodeCreate + "=1" +
		"&id_player=" + fmt.Sprint(thg.IdPlayer) +
		"&id_node=" + fmt.Sprint(ntype) +
		"&net=" + net +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	data := ParseDataDog(GetNormalSymbols(res))
	if data[0][0][1] != "ok" {
		return errors.New(cmdNodeCreate + ": invalid response")
	}

	return nil
}

func (thg *Thg) NodeDelete(id uint32, net string) error {
	req := cmdNodeDelete + "=1" +
		"&id_player=" + fmt.Sprint(thg.IdPlayer) +
		"&id=" + fmt.Sprint(id) +
		"&net=" + net +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	if res != "ok" {
		return errors.New(cmdNodeDelete + ": invalid response")
	}

	return nil
}

func (thg *Thg) NodeUpgrade(id uint32) error {
	req := cmdNodeUpgrade + "=1" +
		"&id=" + fmt.Sprint(id) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	if res != "ok" {
		return errors.New(cmdNodeUpgrade + ": invalid response")
	}

	return nil
}

func (thg *Thg) NodeSetBuilders(id uint32, builders byte) error {
	req := cmdNodeSetBuilders + "=1" +
		"&id_node=" + fmt.Sprint(id) +
		"&builders=" + fmt.Sprint(builders) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	if res != "ok" {
		return errors.New(cmdNodeSetBuilders + ": invalid response")
	}

	return nil
}

func (thg *Thg) NodeCollect(id uint32) error {
	req := cmdNodeCollect + "=1" +
		"&id_node=" + fmt.Sprint(id) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	_, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	return nil
}

func (thg *Thg) ProgramUpgrade(id uint32) error {
	req := cmdProgramUpgrade + "=1" +
		"&id=" + fmt.Sprint(id) +
		"&app_version=" + fmt.Sprint(thg.AppVersion) +
		"&session_id=" + thg.SessionId

	res, err := thg.HttpGet(thg.GetFullUrl(req))
	if err != nil {
		return err
	}

	if res != "ok" {
		return errors.New(cmdProgramUpgrade + ": invalid response")
	}

	return nil
}
