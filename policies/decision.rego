package deployment

import rego.v1

decision := {
	"allow": count(violation) == 0,
	"violations": [finding | finding := violation[_]],
}
