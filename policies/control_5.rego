package deployment

import rego.v1

violation contains finding if {
	control_no := "control_5"

	lower(input.namespace) == "production"
	input.replicas < 2

	finding := {
		"control_no": control_no,
	}
}
