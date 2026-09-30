package entity

// Bit16 is the bit length, and QueueSize defines the size of the queue.
const (
	Bit16                      = 16
	QueueOutboundSize          = 1024
	QueueInboundSize           = 1024
	PreallocatedBuffersPerPool = 0 // Disable and allow for infinite memory growth
)

// MaxMessageSize is 65535=(2^16)-1.
const MaxMessageSize = (1 << Bit16) - 1
