package robotmgr

import (
	"sync"
	"time"
)

var (
	once                = sync.Once{}
	robotmMgr *RobotMgr = nil
)

// singleton
func Instance() *RobotMgr {
	once.Do(func() {
		robotmMgr = newRobotMgr()
	})
	return robotmMgr
}

type RobotMgr struct {
	robot *Robot
}

func newRobotMgr() *RobotMgr {
	mgr := &RobotMgr{
		robot: newRobot(),
	}
	return mgr
}

func (mgr *RobotMgr) Start() {
	go mgr.loop()
}

func (mgr *RobotMgr) loop() {
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()

	for range t.C {
		mgr.robot.Update()
	}
}
