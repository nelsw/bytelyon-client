package logs

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

var bannerTopLines = []string{
	YellowBoldIntense + `  * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
	YellowBoldIntense + `  *                                                                           * `,
	YellowBoldIntense + `  *` + BlueBoldIntense + `    ██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + BlueBoldIntense + `    ██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + BlueBoldIntense + `    ██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + BlueBoldIntense + `    ██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + BlueBoldIntense + `    ██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + BlueBoldIntense + `    ╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + BlackBoldIntense + `    -------------------------------------------------------------------    ` + YellowBoldIntense + `* `,
}

var managerLines = []string{
	YellowBoldIntense + `  *` + Purple + `               ▄▄   ▄▄  ▄▄▄  ▄▄  ▄▄  ▄▄▄   ▄▄▄▄ ▄▄▄▄▄ ▄▄▄▄                 ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + Purple + `               ██▀▄▀██ ██▀██ ███▄██ ██▀██ ██ ▄▄ ██▄▄  ██▄█▄                ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + Purple + `               ██   ██ ██▀██ ██ ▀██ ██▀██ ▀███▀ ██▄▄▄ ██ ██                ` + YellowBoldIntense + `* `,
}

var workerLines = []string{
	YellowBoldIntense + `  *` + Purple + `                  ▄▄   ▄▄  ▄▄▄  ▄▄▄▄  ▄▄ ▄▄ ▄▄▄▄▄ ▄▄▄▄                     ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + Purple + `                  ██ ▄ ██ ██▀██ ██▄█▄ ██▄█▀ ██▄▄  ██▄█▄                    ` + YellowBoldIntense + `* `,
	YellowBoldIntense + `  *` + Purple + `                   ▀█▀█▀  ▀███▀ ██ ██ ██ ██ ██▄▄▄ ██ ██                    ` + YellowBoldIntense + `* `,
}

var bannerBottomLines = []string{
	YellowBoldIntense + `  *                                                                           * `,
	YellowBoldIntense + `  * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
}

func PrintBanner(opts ...map[string]any) {
	var lines []string
	lines = append(lines, "\n")
	lines = append(lines, bannerTopLines...)
	lines = append(lines, bannerBottomLines...)
	lines = append(lines, "\n")
	println(strings.Join(lines, "\n") + Reset)
}

func PrintWorkerBanner(opts ...map[string]any) {
	var lines []string
	lines = append(lines, "\n")
	lines = append(lines, bannerTopLines...)
	lines = append(lines, workerLines...)
	lines = append(lines, bannerBottomLines...)
	lines = append(lines, "\n")
	if len(opts) > 0 {
		for _, k := range slices.Sorted(maps.Keys(opts[0])) {
			v := fmt.Sprint(opts[0][k])
			if len(v) > 10 {
				v = fmt.Sprintf("...%s", v[len(v)-7:])
			}
			lines = append(lines, fmt.Sprintf(" %s%40s: %s%v", Reset, k, BlackIntense, v))
		}
		lines = append(lines, "\n")
	}
	println(strings.Join(lines, "\n") + Reset)
}

func PrintManagerBanner() {
	var lines []string
	lines = append(lines, bannerTopLines...)
	lines = append(lines, managerLines...)
	lines = append(lines, bannerBottomLines...)
	println(strings.Join(lines, "\n") + Reset)
}

// ████████████████████████████████████████████████████████████████████████████████
func PrintNyanCat(lines ...string) {
	println(strings.Join(append([]string{
		BlueIntense + `████████████████████████████████████████████████████████████████████████████████`,
		BlueIntense + `████████████████████████████████████████████████████████████████████████████████`,
		Red + `██████████████████` + BlueIntense + `████████████████` + Black + `██████████████████████████████` + BlueIntense + `████████████████`,
		Red + `████████████████████████████████` + Black + `██` + White + `██████████████████████████████` + Black + `██` + BlueIntense + `██████████████`,
		RedIntense + `████` + Red + `██████████████████████████` + Black + `██` + White + `██████` + Purple + `██████████████████████` + White + `██████` + Black + `██` + BlueIntense + `████████████`,
		RedIntense + `██████████████████████████████` + Black + `██` + White + `████` + Purple + `████████████████` + Black + `████` + Purple + `██████` + White + `████` + Black + `██` + BlueIntense + `██` + Black + `████` + BlueIntense + `██████`,
		RedIntense + `██████████████████████████████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `████` + Black + `██` + Purple + `██████` + White + `██` + Black + `████` + White + `████` + Black + `██` + BlueIntense + `████`,
		YellowIntense + `██████████████████` + RedIntense + `████████████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `██████` + Purple + `██████` + White + `██` + Black + `██` + White + `██████` + Black + `██` + BlueIntense + `████`,
		YellowIntense + `██████████████████████` + Black + `██` + YellowIntense + `██████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `██████` + Black + `████████` + White + `████████` + Black + `██` + BlueIntense + `████`,
		YellowIntense + `████████████████████` + Black + `██` + White + `██` + Black + `██` + YellowIntense + `████` + Black + `██` + White + `██` + Purple + `████████████████` + Black + `██` + White + `██████████████████████` + Black + `██` + BlueIntense + `████`,
		GreenIntense + `██████████████████` + YellowIntense + `██` + Black + `██` + White + `██` + Black + `████████` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██████████████████████████` + Black + `██` + BlueIntense + `██`,
		GreenIntense + `██████████████████████` + White + `████████` + Black + `██` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██████` + YellowIntense + `██` + White + `██████████` + YellowIntense + `██` + Black + `██` + White + `████` + Black + `██` + BlueIntense + `██`,
		GreenIntense + `██████████████████████` + Black + `████` + White + `████` + Black + `██` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██████` + Black + `██` + White + `██████` + Black + `██` + White + `██` + Black + `████` + White + `████` + Black + `██` + BlueIntense + `██`,
		Blue + `██████████████████` + GreenIntense + `████████` + Black + `██████` + White + `██` + Purple + `██████████████` + Black + `██` + White + `██` + Purple + `████` + White + `████████████████` + Purple + `████` + Black + `██` + BlueIntense + `██`,
		Blue + `██████████████████████████████` + Black + `██` + White + `████` + Purple + `██████████████` + Black + `██` + White + `██████` + Black + `████████████` + White + `████` + Black + `██` + BlueIntense + `████`,
		BlueIntense + `██████████████████` + Blue + `████` + Blue + `██████` + Black + `████` + White + `██████` + Purple + `██████████████` + Black + `██` + White + `██████████████████` + Black + `██` + BlueIntense + `██████`,
		BlueIntense + `██████████████████████████` + Black + `██` + White + `██` + Black + `████` + White + `████████████████████` + Black + `██████████████████` + BlueIntense + `████████`,
		BlueIntense + `████████████████████████` + Black + `██` + White + `██████` + Black + `████████████████████████████████` + White + `██` + Black + `██` + BlueIntense + `████████████`,
		BlueIntense + `████████████████████████` + Black + `██` + White + `████` + Black + `██` + BlueIntense + `██` + Black + `██` + White + `████` + BlueIntense + `████████████` + Black + `██` + White + `████` + Black + `████` + White + `████` + Black + `██` + BlueIntense + `████████████`,
		BlueIntense + `████████████████████████` + Black + `██████` + BlueIntense + `████` + Black + `██████` + BlueIntense + `████████████` + Black + `██████` + BlueIntense + `████` + Black + `██████` + BlueIntense + `████████████`,
		`████████████████████████████████████████████████████████████████████████████████` + Reset,
	}, lines...), "\n"))
}
