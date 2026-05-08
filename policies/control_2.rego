package deployment

import rego.v1

violation contains finding if {
	control_no := "control_2"

	input.memory_limit == ""

	finding := {
		"control_no": control_no,
	}
}
