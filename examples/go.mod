module github.com/smartwalle/mdns/examples

go 1.18

require (
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/smartwalle/log4go v1.0.4 // indirect
	github.com/smartwalle/mail4go v1.0.0 // indirect
	github.com/smartwalle/mdns v0.0.0
	golang.org/x/net v0.0.0-20220822230855-b0a4917ee28c // indirect
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10 // indirect
)

replace (
	github.com/smartwalle/mdns => ../
)
