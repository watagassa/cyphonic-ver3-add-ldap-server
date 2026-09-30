package entity

type RegistrationResponse struct {
	BaseHeader    BaseHeader
	KeepAliveFlag KeepAliveFlagClass
	Padding       [3]byte
	HMAC          HMAC
}

type KeepAliveFlagClass uint8

const (
	KeepAliveFlagClassInactive KeepAliveFlagClass = iota
	KeepAliveFlagClassActive
)

const (
	KeepAliveFlagSize               = 1
	RegistrationResponsePaddingSize = 3
)

const BaseRegistrationResponseLength = BaseHeaderLength + KeepAliveFlagSize + RegistrationResponsePaddingSize + HMACLength

func UnmarshalRegistrationResponse(buffer []byte) (*RegistrationResponse, error) {
	if len(buffer) < BaseRegistrationResponseLength {
		return nil, ErrInvalidBufferLength
	}

	baseHeader := UnmarshalBaseHeader(buffer[:BaseHeaderLength])

	keepAliveFlagStart := BaseHeaderLength

	// paddingStart := keepAliveFlagStart + KeepAliveFlagSize
	// hmacEnd := paddingStart + RegistrationResponsePaddingSize + HMACLength

	// TODO: HMACに関する検証を行う必要がある
	return &RegistrationResponse{
		BaseHeader:    *baseHeader,
		KeepAliveFlag: KeepAliveFlagClass(buffer[keepAliveFlagStart]),
		Padding:       [3]byte{},
		HMAC: HMAC{
			HMAC: make([]byte, HMACLength),
		},
	}, nil
}
