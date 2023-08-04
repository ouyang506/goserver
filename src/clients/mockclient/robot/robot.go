package robot

import (
	"github.com/looplab/fsm"
)

type Robot struct {
	fsm *fsm.FSM
}

func NewRobot() *Robot {
	robot := &Robot{}
	robot.fsm = fsm.NewFSM()
	return robot
}
