package mailer

import "net/textproto"

// Mail defines the email headers/body/attachments.
type Mail struct {
	From, Sender, Subject, Template, Locale, HTML, Text string
	To, ReplyTo, Bcc, Cc, Attachments, ReadReceipt      []string
	Headers                                             textproto.MIMEHeader
	TemplateData                                        interface{}
}
