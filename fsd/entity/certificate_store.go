package entity

const (
	// Local: using certificates from the local directory.
	Local string = "local"

	// Remote: using certificates from the remote bucket's.
	Remote string = "remote"
)

// key definition for certificateStores.
const (
	RootCert          string = "RootCertificate"
	ServerCert        string = "ServerCertificate"
	ServerCertPrivKey string = "ServerCertificatePrivateKey"
)
