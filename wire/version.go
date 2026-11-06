package wire

import "github.com/easyssl-io/easyssl-contract/plugincontract"

// ContractVersion is this module's one version, the same constant
// plugincontract declares: `wire`, `plugincontract` and `l10n` are released
// together under one tag, so an Agent and a Server compare one number.
const ContractVersion = plugincontract.ContractVersion

// Every request an Agent makes to an Agent route (the RouteAgent constants)
// carries these two headers, so the Server learns which build is speaking
// before it reads a body: the Agent's own version in User-Agent, as
// UserAgentPrefix followed by it, and the contract version it was built
// against in HeaderContractVersion. A Server answers an Agent whose contract
// version is not plugincontract.Compatible with CodeContractUnsupported on
// every such route, so an Agent too old or too new for it is told so rather
// than failing on a field one side does not know. The Marketplace routes are
// not Agent routes: a browser reads them too.
const (
	HeaderContractVersion = "Easyssl-Contract-Version"
	UserAgentPrefix       = "easyssl-agent/"
)
