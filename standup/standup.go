package standup

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/abourget/slick"
	"github.com/boltdb/bolt"
)

type Standup struct {
	bot            *slick.Bot
	sectionUpdates chan sectionUpdate
}

const (
	TODAY   = 0
	WEEKAGO = -6 // [0,-6] == 7 days

	bucketKey = "standup"
)

func init() {
	slick.RegisterPlugin(&Standup{})
}

func (standup *Standup) InitPlugin(bot *slick.Bot) {
	standup.bot = bot
	standup.sectionUpdates = make(chan sectionUpdate, 15)

	go standup.manageUpdatesInteraction()

	bot.Listen(&slick.Listener{
		MessageHandlerFunc: standup.ChatHandler,
	})
}

func (standup *Standup) ChatHandler(listen *slick.Listener, msg *slick.Message) {
	if res := sectionRegexp.FindAllStringSubmatchIndex(msg.Text, -1); res != nil {
		for _, section := range extractSectionAndText(msg.Text, res) {
			standup.TriggerReminders(msg, section.name)
			if err := standup.storeUpdate(msg, section); err != nil {
				log.Errorf("Problem storing update: %s", err)
			} else {
				log.Info("ok!")
			}
		}
	}
}

func (standup *Standup) storeUpdate(msg *slick.Message, section sectionMatch) error {
	log.Infof("Storing update for user-id=%s name=%s section=%+v", msg.FromUser.ID, msg.FromUser.Name, section)
	err := standup.bot.DB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucketKey)); err != nil {
			return err
		}
		var (
			today   = getStandupDate(0).String()
			bucket  = tx.Bucket([]byte(bucketKey))
			srcData = bucket.Get([]byte(bucketKey))
			sm      = standupMap{}
		)
		if srcData != nil {
			if err := json.Unmarshal(srcData, &sm); err != nil {
				return err
			}
		}
		var found bool
		if sus, ok := sm[today]; ok {
			for i, _ := range sus {
				su := &sus[i]
				if su.User.ID == msg.FromUser.ID {
					if err := su.Data.Update(section); err != nil {
						return err
					}
					found = true
					break
				}
			}
		} else {
			sm[today] = standupUsers{}
		}
		if !found {
			su := standupUser{
				User: msg.FromUser,
				Data: standupData{},
			}
			if err := su.Data.Update(section); err != nil {
				return err
			}
			sm[today] = append(sm[today], su)
		}
		if dstData, err := json.Marshal(sm); err != nil {
			return err
		} else if saveErr := bucket.Put([]byte(bucketKey), dstData); saveErr != nil {
			return saveErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
