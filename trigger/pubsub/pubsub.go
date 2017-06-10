package pubsub

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"

	"github.com/rusenask/keel/provider"
	"github.com/rusenask/keel/types"
	"github.com/rusenask/keel/util/version"

	log "github.com/Sirupsen/logrus"
)

type Subscriber struct {
	providers map[string]provider.Provider

	project      string
	topic        string
	subscription string

	client *pubsub.Client
}

type Opts struct {
	Project      string
	Topic        string
	Subscription string
	Providers    map[string]provider.Provider
}

func NewSubscriber(opts *Opts) (*Subscriber, error) {
	client, err := pubsub.NewClient(context.Background(), opts.Project)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		project:      opts.Project,
		topic:        opts.Topic,
		subscription: opts.Subscription,
		providers:    opts.Providers,
		client:       client,
	}, nil
}

// Message - expected message from gcr
type Message struct {
	Action string `json:"action,omitempty"`
	Tag    string `json:"tag,omitempty"`
}

// Subscribe - initiate subscriber
func (s *Subscriber) Subscribe(ctx context.Context) error {
	sub := s.client.Subscription(s.subscription)
	log.Info("trigger.pubsub: subscribing for events...")
	// err := sub.Receive(ctx, s.callback)
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.WithFields(log.Fields{
			"payload":     msg.Data,
			"attributeds": msg.Attributes,
		}).Info("trigger.pubsub: received message")
		defer msg.Ack()
		var decoded Message
		err := json.Unmarshal(msg.Data, &decoded)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("trigger.pubsub: failed to decode message")
			return
		}

		if decoded.Tag == "" {
			return
		}

		parsedVersion, err := version.GetVersionFromImageName(decoded.Tag)
		if err != nil {
			return
		}

		// sending event to the providers
		log.WithFields(log.Fields{
			"action":  decoded.Action,
			"tag":     decoded.Tag,
			"version": parsedVersion.String(),
		}).Info("trigger.pubsub: got message")
		event := types.Event{
			Repository: types.Repository{Name: decoded.Tag, Tag: parsedVersion.String()},
			CreatedAt:  time.Now(),
		}
		for _, p := range s.providers {
			err = p.Submit(event)
			if err != nil {
				log.WithFields(log.Fields{
					"error":    err,
					"provider": p.GetName(),
					"version":  parsedVersion.String(),
					"image":    decoded.Tag,
				}).Error("trigger.pubsub: got error while submitting event")
			}
		}

	})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("trigger.pubsub: got error while subscribing")
		return err
	}
	return nil
}

func (s *Subscriber) callback(ctx context.Context, m *pubsub.Message) {
	m.Ack()
}