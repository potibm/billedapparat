package cmd

type Code int

const (
	Exit_OK Code = 0

	Exit_Software Code = 70 // EX_SOFTWARE (sysexits)
	Exit_Usage    Code = 64 // EX_USAGE
	Exit_Config   Code = 78 // EX_CONFIG

	Exit_Unavailable Code = 69 // EX_UNAVAILABLE
	Exit_TempFail    Code = 75 // EX_TEMPFAIL
)
