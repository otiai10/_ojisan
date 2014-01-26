package controllers

import (
	"github.com/robfig/revel"
    "github.com/mrjones/oauth"

    "encoding/json"

    "_ojisan/conf/my"

    "fmt"
)

// コンシューマの定義とプロバイダの定義を含んだ
// *oauth.Consumerをつくる
var TWITTER = oauth.NewConsumer(
    // コンシューマの定義
    my.AppTwitterConsumerKey,
    my.AppTwitterConsumerSecret,
    // プロバイダの定義
    oauth.ServiceProvider{
        AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
        RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
        AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
    },
)

type Application struct {
	*revel.Controller
}

// さいしょのページレンダリングだけー
func (c Application) Index() revel.Result {
    if _, ok := c.Session["screenName"]; ok {
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

func (c Application) Authenticate(oauth_verifier string) revel.Result {

    if _, nameExists := c.Session["screenName"]; nameExists {
        // 既にセッションを持っているのでルームにリダイレクトする
        return c.Redirect(WebSocket.Room)
    }

    if oauth_verifier != "" {
        // oauth_verifierがあるということは、
        // リクエストトークンをプロバイダに送ったあとのリダイレクトである
        // TODO: これをアクションとして切り出す
        // これとリクエストトークンを合わせてサイドプロバイダに問い合わせ
        // このユーザのaccess_tokenを獲得する

        // まずはRequestTokenの復元
        requestToken := &oauth.RequestToken{
            c.Session["requestToken"],
            c.Session["requestSecret"],
        }

        // これと、oauth_verifierを用いてaccess_tokenを獲得する
        accessToken, err := TWITTER.AuthorizeToken(requestToken, oauth_verifier)
        if err == nil {
            // {{{ TODO: この辺を domain - infra 構造に押し込める
            // 成功したので、これを用いてユーザ情報を取得する
            resp, _ := TWITTER.Get(
                //"https://api.twitter.com/1.1/statuses/mentions_timeline.json",
                "https://api.twitter.com/1.1/account/verify_credentials.json",
                map[string]string{},
                accessToken,
            )
            defer resp.Body.Close()
            account := struct {
                Name            string `json:"name"`
                ProfileImageUrl string `json:"profile_image_url"`
                ScreenName      string `json:"screen_name"`
            }{}
            _ = json.NewDecoder(resp.Body).Decode(&account)
            fmt.Printf("アカウント情報じゃあああい\n%+v", account)
            // }}}
            // セッションに格納する
            c.Session["name"] = account.Name
            c.Session["screenName"] = account.ScreenName
            c.Session["profileImageUrl"] = account.ProfileImageUrl
        } else {
            // 失敗したので、エラーを吐く
            revel.ERROR.Println("requestTokenとoauth_verifierを用いてaccessTokenを得たかったけど失敗したの図:\t", err)
        }

        // TODO: 何が起きたか関わらず、トップに戻す
        return c.Redirect(Application.Index)
    }

    // ここからは、oauth_verifierが無い状態でこのURLを叩いたとき
    // つまり、ユーザの最初のAuthenticateへのアクセスである

    // まずはverifier獲得した状態でリダイレクトするように促す
    // このアプリケーションのコンシューマキーとコンシューマシークレットを用いて
    // 一時的に使えるrequestTokenの取得を試みる
    requestToken, url, err := TWITTER.GetRequestTokenAndUrl("http://127.0.0.1:9000/Application/Authenticate")
    if err == nil {
        // 一時的に使えるrequestTokenが取得できたので、サーバ側で一次保存しておく
        c.Session["requestToken"] = requestToken.Token
        c.Session["requestSecret"] = requestToken.Secret
        // あとは、ユーザの問題
        // oauth_verifierを取ってきてもらう
        fmt.Println(
            "あとはユーザに取ってきてもらうところ",
            "いってらっしゃい",
            url,
        )
        return c.Redirect(url)
    } else {
        revel.ERROR.Println(
            "そもそもコンシューマキーを用いてリクエストトークン取得できなかったで御座る",
            err,
        )
    }

    // 何が起きてもとりあえずトップへ飛ばす
    return c.Redirect(Application.Index)
}

func init() {
    // revel.Controller.*が実行されるときに必ず呼べる？
    fmt.Println("initだよお")
    // TWITTER.Debug(true)
}
