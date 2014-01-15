package controllers

import (
	"code.google.com/p/go.net/websocket"
	"github.com/robfig/revel"
	"github.com/robfig/revel/samples/chat/app/chatroom"

    "fmt"
    "html"
)

type WebSocket struct {
	*revel.Controller
}

func (c WebSocket) Room() revel.Result {
    user, ok := c.Session["screenName"]
    fmt.Println("入室時セッションチェック", user, ok)
    if ! ok {
        c.Redirect("/")
    }
	return c.Render(user)
}

func (c WebSocket) RoomSocket(ws *websocket.Conn) revel.Result {

    user, ok := c.Session["screenName"]
    fmt.Println("ソケットコネクション時セッションチェック", user, ok)
    if ! ok {
        c.Redirect("/")
    }

	// Join the room.
	subscription := chatroom.Subscribe()
	defer subscription.Cancel()

	chatroom.Join(user)
	defer chatroom.Leave(user)

	// Send down the archive.
	for _, event := range subscription.Archive {
		if websocket.JSON.Send(ws, &event) != nil {
			// They disconnected
			return nil
		}
	}

	// In order to select between websocket messages and subscription events, we
	// need to stuff websocket events into a channel.
	newMessages := make(chan string)
	go func() {
		var msg string
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				close(newMessages)
				return
			}
            user, ok = c.Session["screenName"]
            fmt.Println("メッセージごとのセッションチェック", user, ok, msg)
			newMessages <- msg
		}
	}()

	// Now listen for new events from either the websocket or the chatroom.
	for {
		select {
		case event := <-subscription.New:
			if websocket.JSON.Send(ws, &event) != nil {
				// They disconnected.
				return nil
			}
		case msg, ok := <-newMessages:
			// If the channel is closed, they disconnected.
			if !ok {
				return nil
			}

			// Otherwise, say something.
            // chatroom.Sayでやるべきな気がするけれど
            escaped := html.EscapeString(msg)
			chatroom.Say(user, escaped)
		}
	}
	return nil
}
