// Sugar exposes all of the matchers so that you can dot import them for better readability
package sugar

import v3 "github.com/pact-foundation/pact-go/v3"

// Structures
type MapMatcher = v3.MapMatcher
type Map = v3.MapMatcher
type String = v3.String
type S = v3.S

// Builders
var NewMessagePactV3 = v3.NewMessagePactV3
var NewV2Pact = v3.NewV2Pact
var NewV34Pact = v3.NewV3Pact

// Configs
type MessageConfig = v3.MessageConfig
type MockHTTPProviderConfig = v3.MockHTTPProviderConfig
type MockServerConfig = v3.MockServerConfig
type ProviderStateV3 = v3.ProviderStateV3
type ProviderState = v3.ProviderState
type AsynchronousMessage = v3.AsynchronousMessage

// V2
var Like = v3.Like
var _ = v3.Like
var EachLike = v3.EachLike
var Term = v3.Term
var Regex = v3.Regex
var HexValue = v3.HexValue
var Identifier = v3.Identifier
var IPAddress = v3.IPAddress
var IPv6Address = v3.IPv6Address
var Timestamp = v3.Timestamp
var Date = v3.Date
var Time = v3.Time
var UUID = v3.UUID
var ArrayMinLike = v3.ArrayMinLike

// V3
var Decimal = v3.Decimal
var Integer = v3.Integer
var Equality = v3.Equality
var Includes = v3.Includes
var FromProviderState = v3.FromProviderState
var EachKeyLike = v3.EachKeyLike
var ArrayContaining = v3.ArrayContaining
var ArrayMinMaxLike = v3.ArrayMinMaxLike
var ArrayMaxLike = v3.ArrayMaxLike
var DateGenerated = v3.DateGenerated
var TimeGenerated = v3.TimeGenerated
var DateTimeGenerated = v3.DateTimeGenerated
