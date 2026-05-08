package deployment

import rego.v1

violation contains finding if {
	control_no := "control_3"

	endswith(lower(input.image), ":latest")

	finding := {
		"control_no": control_no,
	}
}
