package params

// HaloBootnodes are the enode URLs of the P2P bootstrap nodes running
// reliably and availably on the Halo network.
var HaloBootnodes = []string{
	// TODO: Add bootnode enodes here after setting them up
	// Format: "enode://[node-id]@[ip]:[port]"
	// Example:
	// "enode://abcd1234...@203.0.113.1:30303",
	// "enode://efgh5678...@203.0.113.2:30303",
	"enode://67b03fc4c109abd319ee7eb156dd0ce25c61ccb07848810fa4142cee50d28e505812f68b7ee5a5260f6904bb0fe7edb37ce0a5af2671206fc1930ae24dbdc22b@51.20.84.193:30301"
}

// HaloDNSNetwork is the DNS discovery configuration for Halo network
// TODO: Set up DNS discovery and configure this value
// var HaloDNSNetwork = "enrtree://[public-key]@halo.example.com"
