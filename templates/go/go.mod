module ad-plugin-template

go 1.25.0

require (
	github.com/phi42/ad-enforcement-tool v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.11
)

// Replace with the local source until ad-enforcement-tool has a published release.
// Remove this directive and pin a real version once the module is released.
replace github.com/phi42/ad-enforcement-tool => ../../
