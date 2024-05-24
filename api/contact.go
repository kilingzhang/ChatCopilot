package api

import (
	"github.com/labstack/echo/v4"
	"github.com/lw396/WeComCopilot/internal/errors"
)

func (a *Api) getContactPerson(c echo.Context) (err error) {
	nickname := c.QueryParam("nickname")
	if nickname == "" {
		return errors.New(errors.CodeInvalidParam, "请输入联系人昵称")
	}

	result, err := a.service.GetContactPersonByNickname(c.Request().Context(), nickname)
	if err != nil {
		return
	}
	return OK(c, result)
}

func (a *Api) saveContactPerson(c echo.Context) (err error) {
	var req ReqSaveMessage
	if err = c.Bind(&req); err != nil {
		return
	}
	if err = c.Validate(&req); err != nil {
		return
	}

	message, err := a.service.ScanMessage(c.Request().Context(), req.Usrname)
	if err != nil {
		return
	}

	contact, err := a.service.GetContactPersonByUsrname(c.Request().Context(), req.Usrname)
	if err != nil {
		return
	}

	contact.DBName = message.DBName
	err = a.service.SaveContactPerson(c.Request().Context(), contact)
	if err != nil {
		return
	}

	return Created(c, "")
}
