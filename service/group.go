package service

import (
	"context"

	"github.com/lw396/WeComCopilot/internal/repository/sqlite"
)

type GroupContact struct {
	UsrName         string `json:"usr_name"`
	Nickname        string `json:"nickname"`
	HeadImgUrl      string `json:"head_img_url"`
	ChatRoomMemList string `json:"member_list"`
	DBName          string `json:"db_name"`
}

func (a *Service) GetGroupContactByNickname(ctx context.Context, nickname string) (result []*GroupContact, err error) {
	if err = a.ConnectDB(ctx, sqlite.GroupDB); err != nil {
		return
	}
	contact, err := a.sqlite.GetGroupContactByNickname(ctx, nickname)
	if err != nil {
		return
	}
	for _, row := range contact {
		result = append(result, &GroupContact{
			UsrName:         row.UsrName,
			Nickname:        row.Nickname,
			HeadImgUrl:      row.HeadImgUrl,
			ChatRoomMemList: row.ChatRoomMemList,
		})
	}
	return
}

func (a *Service) GetGroupContactByUsrname(ctx context.Context, usrname string) (result *GroupContact, err error) {
	return
}

func (a *Service) ConnectDB(ctx context.Context, dbName string) (err error) {
	tx, err := a.sqlite.OpenDB(ctx, dbName)
	if err != nil {
		return
	}
	if tx != nil {
		a.sqlite.BindDB(ctx, tx, dbName)
	}
	return
}
