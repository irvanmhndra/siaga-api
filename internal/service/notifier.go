package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/config"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
)

// Notifier fans an incident out to a user across their enabled channels and
// records every delivery (including in-app) in the notifications table.
//
// External channels are gated by config: when SMTP/Twilio aren't configured the
// send is skipped and only the in-app row is written — same env-driven pattern
// as the rest of the stack, so local dev needs no credentials.
type Notifier struct {
	cfg   *config.Config
	repo  repository.NotificationRepository
	httpc *http.Client
}

func NewNotifier(cfg *config.Config, repo repository.NotificationRepository) *Notifier {
	return &Notifier{cfg: cfg, repo: repo, httpc: &http.Client{Timeout: 8 * time.Second}}
}

// NotifyUser delivers an incident alert to a single user. In-app is always
// recorded; email/SMS/voice depend on the user's prefs and provider config.
func (n *Notifier) NotifyUser(ctx context.Context, u *model.User, inc *model.Incident, voice bool) {
	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(inc.Urgency), inc.Title)
	body := n.incidentBody(inc)

	n.record(ctx, inc, &u.ID, model.ChannelInApp, "sent", inc.Title)

	if u.NotifyEmail && n.cfg.SMTP.Enabled() {
		status := "sent"
		if err := n.sendEmail(u.Email, subject, body); err != nil {
			slog.Error("email notify failed", "user", u.ID, "error", err)
			status = "failed"
		}
		n.record(ctx, inc, &u.ID, model.ChannelEmail, status, subject)
	}

	if u.NotifySMS && u.Phone != nil && *u.Phone != "" && n.cfg.Twilio.Enabled() {
		status := "sent"
		if err := n.sendSMS(*u.Phone, subject+" — "+n.link(inc)); err != nil {
			slog.Error("sms notify failed", "user", u.ID, "error", err)
			status = "failed"
		}
		n.record(ctx, inc, &u.ID, model.ChannelSMS, status, subject)

		if voice {
			vstatus := "sent"
			if err := n.placeCall(*u.Phone, inc.Title); err != nil {
				slog.Error("voice notify failed", "user", u.ID, "error", err)
				vstatus = "failed"
			}
			n.record(ctx, inc, &u.ID, model.ChannelVoice, vstatus, inc.Title)
		}
	}
}

// NotifySlack posts an incident update to a channel webhook. webhookURL falls
// back to the org default when empty; disabled entirely when neither is set.
func (n *Notifier) NotifySlack(ctx context.Context, inc *model.Incident, webhookURL, text string) {
	if webhookURL == "" {
		webhookURL = n.cfg.Slack.DefaultWebhookURL
	}
	if webhookURL == "" {
		return
	}
	status := "sent"
	if err := n.postJSON(ctx, webhookURL, map[string]string{"text": text}); err != nil {
		slog.Error("slack notify failed", "error", err)
		status = "failed"
	}
	n.record(ctx, inc, nil, model.ChannelSlack, status, text)
}

func (n *Notifier) record(ctx context.Context, inc *model.Incident, userID *string, channel, status, message string) {
	rec := &model.Notification{
		ID:             uuid.NewString(),
		OrganizationID: inc.OrganizationID,
		IncidentID:     &inc.ID,
		UserID:         userID,
		Channel:        channel,
		Status:         status,
		Message:        message,
	}
	if err := n.repo.Create(ctx, rec); err != nil {
		slog.Error("failed to record notification", "channel", channel, "error", err)
	}
}

func (n *Notifier) incidentBody(inc *model.Incident) string {
	return fmt.Sprintf("Incident: %s\nUrgency: %s\nStatus: %s\nTriggered: %s\n\n%s",
		inc.Title, inc.Urgency, inc.Status, inc.TriggeredAt.Format(time.RFC1123), n.link(inc))
}

func (n *Notifier) link(inc *model.Incident) string {
	return strings.TrimRight(n.cfg.App.BaseURL, "/") + "/incidents/" + inc.ID
}

func (n *Notifier) sendEmail(to, subject, body string) error {
	addr := n.cfg.SMTP.Host + ":" + n.cfg.SMTP.Port
	msg := "From: " + n.cfg.SMTP.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body
	var auth smtp.Auth
	if n.cfg.SMTP.Username != "" {
		auth = smtp.PlainAuth("", n.cfg.SMTP.Username, n.cfg.SMTP.Password, n.cfg.SMTP.Host)
	}
	return smtp.SendMail(addr, auth, n.cfg.SMTP.From, []string{to}, []byte(msg))
}

func (n *Notifier) sendSMS(to, body string) error {
	form := url.Values{}
	form.Set("To", to)
	form.Set("From", n.cfg.Twilio.FromNumber)
	form.Set("Body", body)
	return n.twilioPost("Messages.json", form)
}

func (n *Notifier) placeCall(to, title string) error {
	twiml := "<Response><Say>Siaga alert. " + escapeXML(title) + ". Please acknowledge.</Say></Response>"
	form := url.Values{}
	form.Set("To", to)
	form.Set("From", n.cfg.Twilio.FromNumber)
	form.Set("Twiml", twiml)
	return n.twilioPost("Calls.json", form)
}

func (n *Notifier) twilioPost(resource string, form url.Values) error {
	endpoint := "https://api.twilio.com/2010-04-01/Accounts/" + n.cfg.Twilio.AccountSID + "/" + resource
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(n.cfg.Twilio.AccountSID, n.cfg.Twilio.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := n.httpc.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("twilio %s status %d", resource, resp.StatusCode)
	}
	return nil
}

func (n *Notifier) postJSON(ctx context.Context, endpoint string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.httpc.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}

func escapeXML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
