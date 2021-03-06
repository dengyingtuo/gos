package player

import (
	"fmt"
	"github.com/kataras/iris/core/errors"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/scene_utils"
	"goslib/session_utils"
)

const SERVER = "__player_manager_server__"

/*
   GenServer Callbacks
*/
type PlayerManager struct {
}

func StartPlayerManager() {
	gen_server.Start(SERVER, new(PlayerManager))
}

func StartPlayer(accountId string) error {
	session, err := session_utils.Find(accountId)
	if err != nil {
		logger.ERR("StartPlayer failed: ", err)
		return err
	}
	scene, err := scene_utils.FindScene(session.SceneId)
	if err != nil {
		logger.ERR("StartPlayer failed: ", err)
		return err
	}
	if scene == nil {
		err = errors.New(fmt.Sprintf("scene: %s not found", session.SceneId))
		logger.ERR("StartPlayer failed: ", err)
		return err
	}
	if scene.GameAppId != CurrentGameAppId {
		err = errors.New("player not belongs to this server!")
		logger.ERR("StartPlayer failed: ", err)
		return err
	}
	_, err = gen_server.Call(SERVER, "StartPlayer", accountId)
	return err
}

func (self *PlayerManager) Init(args []interface{}) (err error) {
	return nil
}

func (self *PlayerManager) HandleCast(args []interface{}) {
}

func (self *PlayerManager) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "StartPlayer" {
		accountId := args[1].(string)
		if !gen_server.Exists(accountId) {
			gen_server.Start(accountId, new(Player), accountId)
		}
	}
	return nil, nil
}

func (self *PlayerManager) Terminate(reason string) (err error) {
	return nil
}
