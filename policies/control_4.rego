package deployment

import rego.v1

violation contains finding if {
	control_no := "control_4"

	input.privileged

	finding := {
		"control_no": control_no,
	}
}
