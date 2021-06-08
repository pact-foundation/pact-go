// Sugar exposes all of the matchers and main interfaces so that you can dot import them for better readability
package sugar

import (
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/message"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/version"
)

// Structures
type MapMatcher = matchers.MapMatcher
type Matcher = matchers.Matcher
type Map = matchers.MapMatcher
type String = matchers.String
type S = matchers.S

// HTTP
var NewV2Pact = consumer.NewV2Pact
var NewV3Pact = consumer.NewV3Pact

type HTTPVerifier = provider.HTTPVerifier

// Message
var NewMessagePactV3 = message.NewMessagePactV3

type MessageVerifier = message.MessageVerifier
type MessageHandlers = message.MessageHandlers
type VerifyMessageRequest = message.VerifyMessageRequest

// Configs
type MessageConfig = message.MessageConfig
type MockHTTPProviderConfig = consumer.MockHTTPProviderConfig
type MockServerConfig = consumer.MockServerConfig
type ProviderStateV3 = models.ProviderStateV3
type ProviderState = models.ProviderState
type VerifyRequest = provider.VerifyRequest
type AsynchronousMessage = message.AsynchronousMessage
type ProviderStateV3Response = models.ProviderStateV3Response
type StateHandlers = models.StateHandlers

// V2
var Like = matchers.Like
var _ = matchers.Like
var EachLike = matchers.EachLike
var Term = matchers.Term
var Regex = matchers.Regex
var HexValue = matchers.HexValue
var Identifier = matchers.Identifier
var IPAddress = matchers.IPAddress
var IPv6Address = matchers.IPv6Address
var Timestamp = matchers.Timestamp
var Date = matchers.Date
var Time = matchers.Time
var UUID = matchers.UUID
var ArrayMinLike = matchers.ArrayMinLike

// V3
var Decimal = matchers.Decimal
var Integer = matchers.Integer
var Equality = matchers.Equality
var Includes = matchers.Includes
var FromProviderState = matchers.FromProviderState
var EachKeyLike = matchers.EachKeyLike
var ArrayContaining = matchers.ArrayContaining
var ArrayMinMaxLike = matchers.ArrayMinMaxLike
var ArrayMaxLike = matchers.ArrayMaxLike
var DateGenerated = matchers.DateGenerated
var TimeGenerated = matchers.TimeGenerated
var DateTimeGenerated = matchers.DateTimeGenerated

// Helpers
var SetLogLevel = log.SetLogLevel
var CheckVersion = version.CheckVersion
