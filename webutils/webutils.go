package webutils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/abourget/slick"
	"github.com/gigawattio/web"
	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
)

type Utils struct {
	bot *slick.Bot
}

func init() {
	slick.RegisterPlugin(&Utils{})
}

func (utils *Utils) InitWebPlugin(bot *slick.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	utils.bot = bot
	privRouter.HandleFunc("/slack/channels", utils.handleGetChannels)
	privRouter.HandleFunc("/slack/users", utils.handleGetUsers)
}

func (utils *Utils) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	out := struct {
		Users map[string]slack.User `json:"users"`
	}{
		Users: utils.bot.Users,
	}

	err := enc.Encode(out)
	if err != nil {
		web.RespondWithJson(w, http.StatusInternalServerError, fmt.Errorf("Error encoding JSON: %s", err))
		return
	}
	return
}

func (utils *Utils) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	out := struct {
		Channels map[string]slick.Channel `json:"channels"`
	}{
		Channels: utils.bot.Channels,
	}

	err := enc.Encode(out)
	if err != nil {
		web.RespondWithJson(w, http.StatusInternalServerError, fmt.Errorf("Error encoding JSON: %s", err))
		return
	}
	return
}
