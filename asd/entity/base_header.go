package entity

// cyphonicVersion is mesuare version of CYPHONIC
const cyphonicVersion = 3

// BaseHeader is the header of a CYPHONIC packet.
// CYPHONIC expects to send multiple signaling messages at one time.
// Example: RegistrationRequest and NotificationRequest send at one time.
// Count the number of signaling messages and add to the BaseHeader struct's Count.
type BaseHeader struct {
	TransactionID  uint32
	Version        uint8
	Type           uint8
	Status         uint8
	Count          uint8
	SequenceNumber uint32
	MessageLength  uint16
	Token          uint8
	NextOpt        uint8
	ID             ID
}

// BaseHeaderLen lengths (bytes).
const BaseHeaderLen = 32

// ID is a CYPHONIC ID, a slice of bytes.
// length of the byte slice: a 16-byte slice.
type ID []byte

// IDlen lengths (bytes).
const IDlen = 16

// TypeClass defines the class associated with a CYPHONIC packet type.
// classes can be thought of as an array of parallel namespace trees.
type TypeClass uint8

// TypeClass known values.
const (
	TypeClassDefault                     TypeClass = 0
	TypeClassAck                         TypeClass = 1
	TypeClassNack                        TypeClass = 2
	TypeClassHolePunching                TypeClass = 3
	TypeClassLoginRequest                TypeClass = 4
	TypeClassLoginResponse               TypeClass = 5
	TypeClassProvisionRequest            TypeClass = 6
	TypeClassProvisionResponse           TypeClass = 7
	TypeClassConnectionRequest           TypeClass = 8
	TypeClassConnectionResponse          TypeClass = 9
	TypeClassRegistrationRequest         TypeClass = 10
	TypeClassRegistrationResponse        TypeClass = 11
	TypeClassDirectionRequest            TypeClass = 12
	TypeClassRouteDirectionToResponder   TypeClass = 13
	TypeClassRouteDirectionConfirmation  TypeClass = 14
	TypeClassRouteDirectionToInitiator   TypeClass = 15
	TypeClassTunnelRequestForUDP         TypeClass = 16
	TypeClassTunnelResponseForUDP        TypeClass = 17
	TypeClassTunnelRequestForQUIC        TypeClass = 18
	TypeClassTunnelResponseForQUIC       TypeClass = 19
	TypeClassCapsuleMessageFromInitiator TypeClass = 20
	TypeClassCapsuleMessageFromResponder TypeClass = 21
	TypeClassHolePunchingForOptimazation TypeClass = 22
	TypeClassAckForOptimazation          TypeClass = 23
	TypeClassFinalizationRequest         TypeClass = 24
	TypeClassKeepAlive                   TypeClass = 25
	TypeClassKeepAliveAck                TypeClass = 26
	TypeClassMultiCastRequest            TypeClass = 27
)

// StatusClass defines the class associated with a CYPHONIC packet status.
// classes can be thought of as an array of parallel namespace trees.
type StatusClass uint8

// StatusClass known values.
const (
	StatusClassSuccess                    StatusClass = 0
	StatusClassInternalServerError        StatusClass = 1
	StatusClassFalsificationDetected      StatusClass = 2
	StatusClassDecryptionFailed           StatusClass = 3
	StatusClassAuthenticationFailed       StatusClass = 4
	StatusClassForbidden                  StatusClass = 5
	StatusClassProvisionFailed            StatusClass = 6
	StatusClassProvisionAliaseFQDNFailed  StatusClass = 7
	StatusClassConnectionResolutionFailed StatusClass = 8
	StatusClassRegistrationFailed         StatusClass = 9
	StatusClassDestinationNotFound        StatusClass = 10
	StatusClassExecutionModeMismatch      StatusClass = 11
	StatusClassProfileMismatch            StatusClass = 12
	StatusClassTunnelConstructionFailed   StatusClass = 13
	StatusClassFinalizationFailed         StatusClass = 14
	StatusClassGeneralFailure             StatusClass = 15
)

// serializeTransactionID is to serialize BaseHeader's TransactionID.
func (b *BaseHeader) serializeTransactionID(prevBaseHeader BaseHeader) {
	b.TransactionID = prevBaseHeader.TransactionID
}

// serializeVersion is to serialize BaseHeader's version.
func (b *BaseHeader) serializeVersion() {
	b.Version = cyphonicVersion
}

// serializeType is to serialize BaseHeader's type.
func (b *BaseHeader) serializeType(typ TypeClass) {
	b.Type = uint8(typ)
}

func (b *BaseHeader) serializeStatus(st StatusClass) {
	b.Status = uint8(st)
}

func (b *BaseHeader) serializeCount() {
	b.Count = 0
}

// serializeSequenceNumber is to serialize BaseHeader's SequenceNumber.
func (b *BaseHeader) serializeSequenceNumber(prevBaseHeader BaseHeader) {
	b.SequenceNumber = prevBaseHeader.SequenceNumber + 1
}

func (b *BaseHeader) serializeMessageLength(length uint16) {
	b.MessageLength = length
}

func (b *BaseHeader) serializeToken() {
	b.Token = 0
}

func (b *BaseHeader) serializeNextOpt() {
	b.NextOpt = 0
}

func (b *BaseHeader) serializeID(id []byte) {
	copy(b.ID, id)
}
