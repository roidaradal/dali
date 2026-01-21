package dali

const currentVersion string = "0.1.4"

var updateNotes = map[string][]string{
	"0.1.4": {
		"`reset` command",
		"`update` notes",
		"Fix discovery on multiple network interfaces (WLAN, Ethernet)",
	},
	"0.1.3": {
		"Add logs for sending / receiving",
		"`logs` command",
		"`logs` filters (date, action, from, to, file)",
		"Prompt name and timeout on first use",
		"Filesize string (KB, MB instead of bytes)",
		"Append _N to destination path if file exists",
		"`open` overwrite",
		"`send` auto=1",
		"`send` wait",
		"Select network if multiple IP address",
		"Finish early in detection if search param found",
	},
	"0.1.2": {
		"Add colors to CLI",
		"Changed default timeout from 5s to 3s",
		"`help` command",
		"`update` command",
		"`find` name={NAME}",
		"`find` ip={IP}",
		"`open` accept=auto",
	},
	"0.1.1": {
		"Removed AI-generated code",
		"`set` command",
		"`open` command",
		"`find` command",
		"`send` command",
		"`version` command",
	},
	"0.1.0": {
		"Vibe coded in Rust",
		"Vibe coded conversion of Rust to Go",
	},
}
