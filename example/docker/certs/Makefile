all:
	go install -v github.com/square/certstrap@latest
	certstrap --depot-path . init --curve P-256 --passphrase "" --common-name "Uptrace CA"
	certstrap --depot-path . request-cert --ou "Uptrace" --curve P-256 --passphrase "" --common-name server --domain *.uptrace.dev
	certstrap --depot-path . sign --expires "99 years" --CA "Uptrace CA" server
