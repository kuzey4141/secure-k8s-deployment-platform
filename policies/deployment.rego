package deployment

import rego.v1

deny contains "CPU limit is required" if {
	input.cpu_limit == ""
}

deny contains "Image tag 'latest' cannot be used" if {
	endswith(lower(input.image), ":latest")
}

deny contains "Memory limit is required" if {
	input.memory_limit == ""
}

deny contains "Privileged containers are not allowed" if {
	input.privileged
}

deny contains "Production deployments must have at least 2 replicas" if {
	lower(input.namespace) == "production"
	input.replicas < 2
}

decision := {
	"allow": count(deny) == 0,
	"deny": sort([reason | reason := deny[_]]),
}
