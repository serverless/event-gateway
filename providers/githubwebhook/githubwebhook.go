package githubwebhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/serverless/event-gateway/function"
	eventpkg "github.com/serverless/event-gateway/event"
	"go.uber.org/zap/zapcore"
	"gopkg.in/go-playground/webhooks.v3/github"
	validator "gopkg.in/go-playground/validator.v9"
)

// Type of provider.
const Type = function.ProviderType("githubwebhook")

func init() {
	function.RegisterProvider(Type, ProviderLoader{})
}

// GithubWebhook function implementation
type GithubWebhook struct {
	Secret             string `json:"secret,omitempty"`
}

// Call parses the event, emits it as a custom event into Event Gateway, and responds to the webhook.
func (g GithubWebhook) Call(payload []byte) ([]byte, error) {
    e := &eventpkg.Event{}
    err := json.Unmarshal(payload, e)
    
    if err != nil {
        return nil, err
    }
    data, ok := e.Data.(*eventpkg.HTTPEvent)
    if !ok {
        fmt.Printf("%+v", data)
        return nil, errors.New("github webhook must be used with HTTP subscription.")
    }

    req, err := makeGithubWebhookRequest(data)
    if err != nil {
		return nil, errors.New("invalid request format")
    }

    if g.Secret != "" {
        ok = validateGithubWebhookRequest([]byte(g.Secret), req.Signature, req.Payload)
        if !ok {
		    return nil, errors.New("invalid signature")
        }
    }

    event, err := makeGithubEvent(req)
    if err != nil {
		return nil, errors.New("internal error")
    }
    
    fmt.Printf("I would emit the following event back into the Gateway: %+v", event)

    return []byte(`{ "statusCode": 200, }`), nil
}

type githubWebhookRequest struct {
    Signature   string
    Event       string
    ID          string
    Payload     []byte
}

func makeGithubWebhookRequest(data *eventpkg.HTTPEvent) (*githubWebhookRequest, error) {
    req := githubWebhookRequest{}

    if req.Signature = data.Headers["x-hub-signature"]; len(req.Signature) == 0 {
        return nil, errors.New("invalid Github webhook request. Missing signature header")
    }

    if req.Event = data.Headers["x-github-event"]; len(req.Event) == 0 {
        return nil, errors.New("invalid Github webhook request. Missing event header")
    }

    if req.ID = data.Headers["x-github-delivery"]; len(req.ID) == 0 {
        return nil, errors.New("invalid Github webhook request. Missing event ID")
    }

    body, ok := data.Body.([]byte); 
    if !ok {
		return nil, errors.New("could not process payload body")
    }
    req.Payload = body

    return &req, nil
}

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

func validateGithubWebhookRequest(secret []byte, signature string, body []byte) bool {

	const signaturePrefix = "sha1="
	const signatureLength = 45

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		return false
	}

	actual := make([]byte, 20)
	hex.Decode(actual, []byte(signature[5:]))

	return hmac.Equal(signBody(secret, body), actual)
}

func makeGithubEvent(req *githubWebhookRequest) (*eventpkg.Event, error) {
    eventType := eventpkg.Type("github." + req.Event)
    event := eventpkg.New(eventType, "application/json", "")
    githubEvent := github.Event(req.Event)
	switch githubEvent {
	case github.CommitCommentEvent:
		var cc github.CommitCommentPayload
		json.Unmarshal([]byte(req.Payload), &cc)
		event.Data = cc
	case github.CreateEvent:
		var c github.CreatePayload
		json.Unmarshal([]byte(req.Payload), &c)
		event.Data = c
	case github.DeleteEvent:
		var d github.DeletePayload
		json.Unmarshal([]byte(req.Payload), &d)
		event.Data = d
	case github.DeploymentEvent:
		var d github.DeploymentPayload
		json.Unmarshal([]byte(req.Payload), &d)
		event.Data = d
	case github.DeploymentStatusEvent:
		var d github.DeploymentStatusPayload
		json.Unmarshal([]byte(req.Payload), &d)
		event.Data = d
	case github.ForkEvent:
		var f github.ForkPayload
		json.Unmarshal([]byte(req.Payload), &f)
		event.Data = f
	case github.GollumEvent:
		var g github.GollumPayload
		json.Unmarshal([]byte(req.Payload), &g)
		event.Data = g
	case github.InstallationEvent, github.IntegrationInstallationEvent:
		var i github.InstallationPayload
		json.Unmarshal([]byte(req.Payload), &i)
		event.Data = i
	case github.IssueCommentEvent:
		var i github.IssueCommentPayload
		json.Unmarshal([]byte(req.Payload), &i)
		event.Data = i
	case github.IssuesEvent:
		var i github.IssuesPayload
		json.Unmarshal([]byte(req.Payload), &i)
		event.Data = i
	case github.LabelEvent:
		var l github.LabelPayload
		json.Unmarshal([]byte(req.Payload), &l)
		event.Data = l
	case github.MemberEvent:
		var m github.MemberPayload
		json.Unmarshal([]byte(req.Payload), &m)
		event.Data = m
	case github.MembershipEvent:
		var m github.MembershipPayload
		json.Unmarshal([]byte(req.Payload), &m)
		event.Data = m
	case github.MilestoneEvent:
		var m github.MilestonePayload
		json.Unmarshal([]byte(req.Payload), &m)
		event.Data = m
	case github.OrganizationEvent:
		var o github.OrganizationPayload
		json.Unmarshal([]byte(req.Payload), &o)
		event.Data = o
	case github.OrgBlockEvent:
		var o github.OrgBlockPayload
		json.Unmarshal([]byte(req.Payload), &o)
		event.Data = o
	case github.PageBuildEvent:
		var p github.PageBuildPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.PingEvent:
		var p github.PingPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.ProjectCardEvent:
		var p github.ProjectCardPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.ProjectColumnEvent:
		var p github.ProjectColumnPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.ProjectEvent:
		var p github.ProjectPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.PublicEvent:
		var p github.PublicPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.PullRequestEvent:
		var p github.PullRequestPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.PullRequestReviewEvent:
		var p github.PullRequestReviewPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.PullRequestReviewCommentEvent:
		var p github.PullRequestReviewCommentPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.PushEvent:
		var p github.PushPayload
		json.Unmarshal([]byte(req.Payload), &p)
		event.Data = p
	case github.ReleaseEvent:
		var r github.ReleasePayload
		json.Unmarshal([]byte(req.Payload), &r)
		event.Data = r
	case github.RepositoryEvent:
		var r github.RepositoryPayload
		json.Unmarshal([]byte(req.Payload), &r)
		event.Data = r
	case github.StatusEvent:
		var s github.StatusPayload
		json.Unmarshal([]byte(req.Payload), &s)
		event.Data = s
	case github.TeamEvent:
		var t github.TeamPayload
		json.Unmarshal([]byte(req.Payload), &t)
		event.Data = t
	case github.TeamAddEvent:
		var t github.TeamAddPayload
		json.Unmarshal([]byte(req.Payload), &t)
		event.Data = t
	case github.WatchEvent:
		var w github.WatchPayload
		json.Unmarshal([]byte(req.Payload), &w)
		event.Data = w
	}

	return event, nil
}

// validate provider config.
func (g GithubWebhook) validate() error {
	validate := validator.New()
	err := validate.Struct(g)
	if err != nil {
		return err
	}
	return nil
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface.
func (g GithubWebhook) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("secret", g.Secret)
	return nil
}

// ProviderLoader implementation
type ProviderLoader struct{}

// Load decode JSON data as Config and return initialized Provider instance.
func (p ProviderLoader) Load(data []byte) (function.Provider, error) {
	provider := &GithubWebhook{}
	err := json.Unmarshal(data, provider)
	if err != nil {
		return nil, errors.New("unable to load function provider config: " + err.Error())
	}

	err = provider.validate()
	if err != nil {
		return nil, errors.New("missing required fields for Github Webhook function")
	}

	return provider, nil
}
