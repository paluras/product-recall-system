package email

const (
	NotificationHTMLTemplate = `
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

	NotificationTextTemplate = `ALERTE RETRAGERI PRODUSE
------------------------
{{range .Items}}
{{.Title}}
Link: {{.Link}}
Data: {{.Date.Format "02/01/2006"}}

{{end}}

Pentru dezabonare, accesați: http://produseretrase/unsubscribe?token={{.UnsubscribeToken}}`

	VerificationHTMLTemplate = `
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Confirmare Abonare</title>
    </head>
    <body style="margin: 0; padding: 20px; background-color: #f5f5f5; font-family: monospace;">
        <div style="max-width: 600px; margin: 0 auto; background-color: #fff; border: 3px solid #000; padding: 20px; box-sizing: border-box;">
            <div style="margin-bottom: 30px; text-align: center;">
                <div style="width: 60px; height: 60px; background: #000; position: relative; margin: 0 auto 20px;">
                    <div style="position: absolute; color: #fff; font-size: 40px; font-weight: bold; top: 50%; left: 50%; transform: translate(-50%, -50%);">!</div>
                </div>
                <h1 style="margin: 0; font-size: clamp(20px, 5vw, 28px); text-transform: uppercase; border-bottom: 3px solid #000; padding-bottom: 20px;">Confirmare Abonare</h1>
            </div>

            <div style="text-align: center; margin-bottom: 30px;">
                <p style="margin-bottom: 20px;">Vă mulțumim pentru abonare. Pentru a finaliza procesul, vă rugăm să confirmați adresa de email.</p>
                <a href="http://produseretrase.eu/confirm?token={{.Token}}"
                   style="display: inline-block; background-color: #000; color: #fff; padding: 15px 30px; text-decoration: none; font-weight: bold;">
                    Confirmă Abonarea
                </a>
            </div>

            <div style="font-size: 14px; color: #666; text-align: center; margin-top: 30px; padding-top: 20px; border-top: 3px solid #000;">
                <p>Dacă nu ați solicitat această abonare, puteți ignora acest email.</p>
            </div>
        </div>
    </body>
    </html>`

	VerificationTextTemplate = `Confirmare Abonare

Vă mulțumim pentru abonare. Pentru a finaliza procesul, vă rugăm să accesați următorul link:

http://produseretrase.eu/confirm?token={{.Token}}

Dacă nu ați solicitat această abonare, puteți ignora acest email.`
)
