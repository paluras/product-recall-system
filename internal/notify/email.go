package notify

import (
	"bytes"
	"context"
	"html/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/paluras/product-recall-system/internal/models"
)

type AWSConfig struct {
	Region    string
	FromEmail string
}

type EmailService struct {
	sesClient *ses.Client
	config    AWSConfig
	db        *models.DB
}

func NewEmailService(awsConfig AWSConfig, db *models.DB) (*EmailService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsConfig.Region),
	)
	if err != nil {
		return nil, err
	}

	return &EmailService{
		sesClient: ses.NewFromConfig(cfg),
		config:    awsConfig,
		db:        db,
	}, nil
}

func (s *EmailService) SendBatchNotification(recipients []string, items []models.ScrapedItem) error {

	tokenMap := make(map[string]string)
	for _, recipient := range recipients {
		token, err := s.db.CreateUnsubscribeToken(recipient)
		if err != nil {
			return err
		}
		tokenMap[recipient] = token
	}

	emailTemplate := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Alerte Retragere Produse</title>
	</head>
	<body style="margin: 0; padding: 20px; background-color: #f5f5f5; font-family: monospace;">
		<div style="max-width: 600px; margin: 0 auto; background-color: #fff; border: 3px solid #000; padding: 20px; box-sizing: border-box;">
			<!-- Logo and Header -->
			<div style="margin-bottom: 30px; text-align: center;">
				<div style="width: 60px; height: 60px; background: #000; position: relative; margin: 0 auto 20px;">
					<div style="position: absolute; color: #fff; font-size: 40px; font-weight: bold; top: 50%; left: 50%; transform: translate(-50%, -50%);">!</div>
				</div>
				<h1 style="margin: 0; font-size: clamp(20px, 5vw, 28px); text-transform: uppercase; border-bottom: 3px solid #000; padding-bottom: 20px;">Retrageri Noi de Produse</h1>
			</div>

			<!-- Product Recalls -->
			{{range .Items}}
			<div style="margin-bottom: 30px; padding: 15px; border: 3px solid #000; background-color: #fff;">
				<h2 style="margin: 0 0 15px 0; font-family: monospace; font-size: clamp(16px, 4vw, 20px); line-height: 1.4; word-break: break-word;">
					<a href="{{.Link}}" style="color: #000; text-decoration: none; border-bottom: 2px solid #ff0000; display: inline-block;">
						{{.Title}}
					</a>
				</h2>
				<div style="font-family: monospace; color: #666; font-size: 14px; text-transform: uppercase;">
					Data Publicării: {{.Date.Format "02/01/2006"}}
				</div>
			</div>
			{{end}}

			<!-- Footer -->
			 <div style="margin-top: 30px; padding-top: 20px; border-top: 3px solid #000; font-size: 14px; color: #666; text-align: center;">
        		<p style="margin: 0 0 10px 0;">Primiți acest email deoarece v-ați abonat la alertele noastre despre retragerile de produse.</p>
        		<p style="margin: 0;">
           			 <a href="http://produseretrase.eu/unsubscribe?token={{.UnsubscribeToken}}"
               style="color: #ff0000; text-decoration: none; display: inline-block; border: 2px solid #ff0000; padding: 10px 20px; margin-top: 10px;">
               Dezabonare
           			 </a>
        </p>
    </div>
		</div>
	</body>
	</html>`

	textTemplate := `ALERTE RETRAGERI PRODUSE
------------------------
{{range .Items}}
{{.Title}}
Link: {{.Link}}
Data: {{.Date.Format "02/01/2006"}}

{{end}}

Pentru dezabonare, accesați: http://produseretrase/unsubscribe?token={{.UnsubscribeToken}}`

	for recipient, token := range tokenMap {
		data := struct {
			Items            []models.ScrapedItem
			UnsubscribeToken string
		}{
			Items:            items,
			UnsubscribeToken: token,
		}

		htmlTmpl, err := template.New("email").Parse(emailTemplate)
		if err != nil {
			return err
		}
		var htmlBuffer bytes.Buffer
		err = htmlTmpl.Execute(&htmlBuffer, data)
		if err != nil {
			return err
		}
		htmlBody := htmlBuffer.String()

		textTmpl, err := template.New("email").Parse(textTemplate)
		if err != nil {
			return err
		}
		var textBuffer bytes.Buffer
		err = textTmpl.Execute(&textBuffer, data)
		if err != nil {
			return err
		}
		textBody := textBuffer.String()

		input := &ses.SendEmailInput{
			Destination: &types.Destination{
				BccAddresses: []string{recipient},
			},
			Message: &types.Message{
				Body: &types.Body{
					Html: &types.Content{
						Data: &htmlBody,
					},
					Text: &types.Content{
						Data: &textBody,
					},
				},
				Subject: &types.Content{
					Data: aws.String("Alerte Noi Retrageri de Produse"),
				},
			},
			Source: &s.config.FromEmail,
		}

		_, err = s.sesClient.SendEmail(context.TODO(), input)
		if err != nil {
			return err
		}
	}

	return nil
}
