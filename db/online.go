package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"go-mahjong-server/db/model"
	"go-mahjong-server/pkg/errutil"
)

func InsertOnline(count int, deskCount int) {
	o := model.Online{
		Time:      time.Now().Unix(),
		UserCount: count,
		DeskCount: deskCount,
	}

	_, err := database.Insert(o)
	if err != nil {
		log.Errorf("统计在线人数失败: %s", err.Error())
	}
}

func OnlineStats(begin, end int64) ([]model.Online, error) {
	if begin > end {
		return nil, errutil.ErrIllegalParameter
	}

	list := []model.Online{}

	return list, database.Where("`time` BETWEEN ? AND ?", begin, end).Find(&list)
}
