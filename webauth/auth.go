package webauth

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/abourget/slick"
	"github.com/gigawattio/web"
	"github.com/nlopes/slack"
	"golang.org/x/oauth2"
)

func init() {
	oauth2.RegisterBrokenAuthHeaderProvider("https://slack.com/")
	slick.RegisterPlugin(&OAuthPlugin{})
	gob.Register(&slack.User{})
}

type OAuthPlugin struct {
	config    OAuthConfig
	webserver slick.WebServer
}

type OAuthConfig struct {
	RedirectURL  string `json:"oauth_base_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (p *OAuthPlugin) InitWebServerAuth(bot *slick.Bot, webserver slick.WebServer) {
	p.webserver = webserver

	var config struct {
		WebAuthConfig OAuthConfig
	}
	bot.LoadConfig(&config)

	conf := config.WebAuthConfig
	webserver.SetAuthMiddleware(func(handler http.Handler) http.Handler {
		return &OAuthMiddleware{
			handler:   handler,
			plugin:    p,
			webserver: webserver,
			bot:       bot,
			oauthCfg: &oauth2.Config{
				ClientID:     conf.ClientID,
				ClientSecret: conf.ClientSecret,
				RedirectURL:  conf.RedirectURL + "/oauth2callback",
				Scopes:       []string{"identify"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://slack.com/oauth/authorize",
					TokenURL: "https://slack.com/api/oauth.access",
				},
			},
		}
	})
	webserver.SetAuthenticatedUserFunc(p.AuthenticatedUser)
}

func (p *OAuthPlugin) AuthenticatedUser(r *http.Request) (*slack.User, error) {
	sess := p.webserver.GetSession(r)

	rawProfile, ok := sess.Values["profile"]
	if ok == false {
		return nil, fmt.Errorf("Not authenticated")
	}
	profile, ok := rawProfile.(*slack.User)
	if ok == false {
		return nil, fmt.Errorf("Profile data unreadable")
	}
	return profile, nil
}

type OAuthMiddleware struct {
	handler   http.Handler
	plugin    *OAuthPlugin
	webserver slick.WebServer
	oauthCfg  *oauth2.Config
	bot       *slick.Bot
}

func (mw *OAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/oauth2callback" {
		mw.handleOAuth2Callback(w, r)
		return
	}

	if _, err := mw.plugin.AuthenticatedUser(r); err != nil {
		cookie := &http.Cookie{
			Name:     "return-to",
			Value:    r.URL.Path,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(5 * time.Minute),
		}
		http.SetCookie(w, cookie)
		log.Errorf("Not logged in: %s", err)
		url := mw.oauthCfg.AuthCodeURL("", oauth2.SetAuthURLParam("team", mw.bot.Config.TeamID))
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	// Check if session exists, yield a 403 unless we're on the main page
	mw.handler.ServeHTTP(w, r)
}

func (mw *OAuthMiddleware) handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	profile, err := mw.doOAuth2Roundtrip(w, r)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		// Mark logged in
		sess := mw.webserver.GetSession(r)
		sess.Values["profile"] = profile
		err := sess.Save(r, w)
		if err != nil {
			log.Errorf("Error saving cookie: %s", err)
			w.Write([]byte(err.Error()))
			return
		}

		if cookie, err := r.Cookie("return-to"); err == nil && cookie.Value != "" {
			// "Remove" the cookie by sending an "old" cookie with an expired date.
			oldCookie := &http.Cookie{
				Name:     "return-to",
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(-999 * time.Hour),
			}
			http.SetCookie(w, oldCookie)
			http.Redirect(w, r, cookie.Value, http.StatusFound)
		} else if err != nil {
			web.RespondWithText(w, http.StatusInternalServerError, err)
			return
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

func (mw *OAuthMiddleware) doOAuth2Roundtrip(w http.ResponseWriter, r *http.Request) (*slack.User, error) {
	code := r.FormValue("code")

	token, err := mw.oauthCfg.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Errorf("OAuth2 problem: %s", err)
		return nil, fmt.Errorf("Error processing token.")
	}
	client := slack.New(token.AccessToken)

	resp, err := client.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("User unauthenticated: %s", err)
	}

	expectedURL := fmt.Sprintf("https://%s.slack.com/", mw.bot.Config.TeamDomain)
	if resp.URL != expectedURL {
		return nil, fmt.Errorf("Authenticated for wrong domain: %q != %q", resp.URL, expectedURL)
	}

	return mw.bot.GetUser(resp.User), nil
}
