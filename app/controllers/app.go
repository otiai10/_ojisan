package controllers

import (
	"github.com/robfig/revel"
)

type Application struct {
	*revel.Controller
}

// さいしょのページレンダリングだけー
func (c Application) Index() revel.Result {
    if _, ok := c.Session["screenName"]; ok {
        c.Flash.Success("You already have session.")
        return c.Redirect(WebSocket.Room)
    }
	return c.Render()
}

func (c Application) CheckIn(user string) revel.Result {
    // ここで、oauthして、screenNameを取得して、セッションにぶち込む
    c.Session["screenName"] = user
    /// バリデーションせなアカン気がする
    return c.Redirect(WebSocket.Room)
}
