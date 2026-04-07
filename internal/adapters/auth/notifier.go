package authadapter

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/go-sum/auth/model"
	"github.com/go-sum/componentry/email"
	"github.com/go-sum/send"

	g "maragu.dev/gomponents"
)

type Notifier struct {
	sender   send.Sender
	sendFrom string
}

func NewNotifier(sender send.Sender, sendFrom string) *Notifier {
	return &Notifier{sender: sender, sendFrom: sendFrom}
}

func (n *Notifier) SendVerification(ctx context.Context, input model.DeliveryInput) error {
	subject := verificationSubject(input.Purpose)
	return n.sender.Send(ctx, send.Message{
		To:      input.Email,
		From:    n.sendFrom,
		Subject: subject,
		HTML:    renderEmailHTML(verificationBody(input)),
		Text:    verificationText(input),
	})
}

func verificationSubject(purpose model.FlowPurpose) string {
	switch purpose {
	case model.FlowPurposeSignup:
		return "Verify your signup code"
	case model.FlowPurposeEmailChange:
		return "Verify your email change"
	default:
		return "Verify your sign in"
	}
}

func renderEmailHTML(body g.Node) string {
	var buf bytes.Buffer
	_ = body.Render(&buf)
	return buf.String()
}

func verificationBody(input model.DeliveryInput) g.Node {
	title := verificationSubject(input.Purpose)
	return email.Layout(title, g.Group([]g.Node{
		email.H1(title),
		email.P("Use this 6-digit code to continue: " + input.Code),
		email.P("Or open this secure verification link: " + input.VerifyURL),
		email.P("This code expires at " + input.ExpiresAt.UTC().Format(time.RFC1123) + "."),
	}))
}

func verificationText(input model.DeliveryInput) string {
	return strings.Join([]string{
		verificationSubject(input.Purpose),
		"",
		"Code: " + input.Code,
		"",
		"Verify link: " + input.VerifyURL,
		"",
		"Expires at: " + input.ExpiresAt.UTC().Format(time.RFC1123),
	}, "\r\n")
}
