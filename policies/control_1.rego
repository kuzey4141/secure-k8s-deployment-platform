package deployment

import rego.v1

violation contains finding if {
	control_no := "control_1"

	input.cpu_limit == ""

	finding := {
		"control_no": control_no,
	}
}
