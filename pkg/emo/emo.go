package emo

const (
	Truthy = `✅`
	Falsey = `❌`
	Bot    = `🤖`
)

func Bool(b bool) string {
	if b {
		return Truthy
	}
	return Falsey
}
