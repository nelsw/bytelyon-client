package logs

// ROYGBIV is a common acronym representing the seven colors of the rainbow in order:
// Red, Orange, Yellow, Green, Blue, Indigo, and Violet.
// Coined by Isaac Newton to connect light spectrum colors with musical notes,
// it is a mnemonic device for remembering spectral order, often expanded as
// "Richard Of York Gave Battle In Vain".
const (
	Reset   = "\033[0m"
	Default = Reset
	Black   = "\033[0;30m"
	Red     = "\033[0;31m"
	Green   = "\033[0;32m"
	Yellow  = "\033[0;33m"
	Blue    = "\033[0;34m"
	Purple  = "\033[0;35m"
	Cyan    = "\033[0;36m"
	White   = "\033[0;37m"

	BlackBold  = "\033[1;30m"
	RedBold    = "\033[1;31m"
	GreenBold  = "\033[1;32m"
	YellowBold = "\033[1;33m"
	BlueBold   = "\033[1;34m"
	PurpleBold = "\033[1;35m"
	CyanBold   = "\033[1;36m"
	WhiteBold  = "\033[1;37m"

	BlackUnderline  = "\033[4;30m"
	RedUnderline    = "\033[4;31m"
	GreenUnderline  = "\033[4;32m"
	YellowUnderline = "\033[4;33m"
	BlueUnderline   = "\033[4;34m"
	PurpleUnderline = "\033[4;35m"
	CyanUnderline   = "\033[4;36m"
	WhiteUnderline  = "\033[4;37m"

	BlackBackground  = "\033[40m"
	RedBackground    = "\033[41m"
	GreenBackground  = "\033[42m"
	YellowBackground = "\033[43m"
	BlueBackground   = "\033[44m"
	PurpleBackground = "\033[45m"
	CyanBackground   = "\033[46m"
	WhiteBackground  = "\033[47m"

	BlackIntense  = "\033[0;90m"
	RedIntense    = "\033[0;91m"
	GreenIntense  = "\033[0;92m"
	YellowIntense = "\033[0;93m"
	BlueIntense   = "\033[0;94m"
	PurpleIntense = "\033[0;95m"
	CyanIntense   = "\033[0;96m"
	WhiteIntense  = "\033[0;97m"

	BlackBoldIntense  = "\033[1;90m"
	RedBoldIntense    = "\033[1;91m"
	GreenBoldIntense  = "\033[1;92m"
	YellowBoldIntense = "\033[1;93m"
	BlueBoldIntense   = "\033[1;94m"
	PurpleBoldIntense = "\033[1;95m"
	CyanBoldIntense   = "\033[1;96m"
	WhiteBoldIntense  = "\033[1;97m"

	BlackBackgroundIntense  = "\033[0;100m"
	RedBackgroundIntense    = "\033[0;101m"
	OrangeBackgroundIntense = "\033[48:5:208m"
	YellowBackgroundIntense = "\033[0;103m"
	GreenBackgroundIntense  = "\033[0;102m"
	BlueBackgroundIntense   = "\033[0;104m"
	PurpleBackgroundIntense = "\033[0;105m"
	CyanBackgroundIntense   = "\033[0;106m"
	WhiteBackgroundIntense  = "\033[0;107m"
)

// PreviewROYGBIV previews the colors in the order of the ROYGBIV acronym as Isaac Newton intended it.
// Queue Harry Chapin "Flowers are Red" (https://open.spotify.com/track/6m6Wf11DSIy25RSdK4WMZ6?si=6b373b9d71144579)
func PreviewROYGBIV() {

	/*
		red
	*/
	println(Color(Red, "red"))
	println(Color(RedBackground, "red background"))
	println(Color(RedBackgroundIntense, "red background intense"))
	println(Color(RedBold, "red bold"))
	println(Color(RedBoldIntense, "red bold intense"))
	println(Color(RedIntense, "red intense"))
	println(Color(RedUnderline, "red underline"))

	/*
		orange
	*/
	println(Color(OrangeBackgroundIntense, "orange background intense"))
	/*
		yellow
	*/
	println(Color(Yellow, "yellow"))
	println(Color(YellowBackground, "yellow background"))
	println(Color(YellowBackgroundIntense, "yellow background intense"))
	println(Color(YellowBold, "yellow bold"))
	println(Color(YellowBoldIntense, "yellow bold intense"))
	println(Color(YellowIntense, "yellow intense"))
	println(Color(YellowUnderline, "yellow underline"))

	println(Color(Green, "green"))
	println(Color(GreenBackground, "green background"))
	println(Color(GreenBackgroundIntense, "green background intense"))
	println(Color(GreenBold, "green bold"))
	println(Color(GreenBoldIntense, "green bold intense"))
	println(Color(GreenIntense, "green intense"))
	println(Color(GreenUnderline, "green underline"))

	println(Color(Blue, "blue"))
	println(Color(BlueBackground, "blue background"))
	println(Color(BlueBackgroundIntense, "blue background intense"))
	println(Color(BlueBold, "blue bold"))
	println(Color(BlueBoldIntense, "blue bold intense"))
	println(Color(BlueIntense, "blue intense"))
	println(Color(BlueUnderline, "blue underline"))

	println(Color(Purple, "purple"))
	println(Color(PurpleBackground, "purple background"))
	println(Color(PurpleBackgroundIntense, "purple background intense"))
	println(Color(PurpleBold, "purple bold"))
	println(Color(PurpleBoldIntense, "purple bold intense"))
	println(Color(PurpleIntense, "purple intense"))
	println(Color(PurpleUnderline, "purple underline"))
}

func PreviewByColor() {
	println(Color(Black, "black"))
	println(Color(BlackBackground, "black background"))
	println(Color(BlackBackgroundIntense, "black background intense"))
	println(Color(BlackBold, "black bold"))
	println(Color(BlackBoldIntense, "black bold intense"))
	println(Color(BlackIntense, "black intense"))
	println(Color(BlackUnderline, "black underline"))

	println(Color(Blue, "blue"))
	println(Color(BlueBackground, "blue background"))
	println(Color(BlueBackgroundIntense, "blue background intense"))
	println(Color(BlueBold, "blue bold"))
	println(Color(BlueBoldIntense, "blue bold intense"))
	println(Color(BlueIntense, "blue intense"))
	println(Color(BlueUnderline, "blue underline"))

	println(Color(Cyan, "cyan"))
	println(Color(CyanBackground, "cyan background"))
	println(Color(CyanBackgroundIntense, "cyan background intense"))
	println(Color(CyanBold, "cyan bold"))
	println(Color(CyanBoldIntense, "cyan bold intense"))
	println(Color(CyanIntense, "cyan intense"))
	println(Color(CyanUnderline, "cyan underline"))

	println(Color(Green, "green"))
	println(Color(GreenBackground, "green background"))
	println(Color(GreenBackgroundIntense, "green background intense"))
	println(Color(GreenBold, "green bold"))
	println(Color(GreenBoldIntense, "green bold intense"))
	println(Color(GreenIntense, "green intense"))
	println(Color(GreenUnderline, "green underline"))

	println(Color(Purple, "purple"))
	println(Color(PurpleBackground, "purple background"))
	println(Color(PurpleBackgroundIntense, "purple background intense"))
	println(Color(PurpleBold, "purple bold"))
	println(Color(PurpleBoldIntense, "purple bold intense"))
	println(Color(PurpleIntense, "purple intense"))
	println(Color(PurpleUnderline, "purple underline"))

	println(Color(Red, "red"))
	println(Color(RedBackground, "red background"))
	println(Color(RedBackgroundIntense, "red background intense"))
	println(Color(RedBold, "red bold"))
	println(Color(RedBoldIntense, "red bold intense"))
	println(Color(RedIntense, "red intense"))
	println(Color(RedUnderline, "red underline"))

	println(Color(White, "white"))
	println(Color(WhiteBackground, "white background"))
	println(Color(WhiteBackgroundIntense, "white background intense"))
	println(Color(WhiteBold, "white bold"))
	println(Color(WhiteBoldIntense, "white bold intense"))
	println(Color(WhiteIntense, "white intense"))
	println(Color(WhiteUnderline, "white underline"))

	println(Color(Yellow, "yellow"))
	println(Color(YellowBackground, "yellow background"))
	println(Color(YellowBackgroundIntense, "yellow background intense"))
	println(Color(YellowBold, "yellow bold"))
	println(Color(YellowBoldIntense, "yellow bold intense"))
	println(Color(YellowIntense, "yellow intense"))
	println(Color(YellowUnderline, "yellow underline"))

	println(Reset)
}

func PreviewByCode() {
	println(Color(Black, "black"))
	println(Color(Red, "red"))
	println(Color(Green, "green"))
	println(Color(Yellow, "yellow"))
	println(Color(Blue, "blue"))
	println(Color(Purple, "purple"))
	println(Color(Cyan, "cyan"))
	println(Color(White, "white"))
	println(Color(BlackBold, "black bold"))
	println(Color(RedBold, "red bold"))
	println(Color(GreenBold, "green bold"))
	println(Color(YellowBold, "yellow bold"))
	println(Color(BlueBold, "blue bold"))
	println(Color(PurpleBold, "purple bold"))
	println(Color(CyanBold, "cyan bold"))
	println(Color(WhiteBold, "white bold"))
	println(Color(BlackUnderline, "black underline"))
	println(Color(RedUnderline, "red underline"))
	println(Color(GreenUnderline, "green underline"))
	println(Color(YellowUnderline, "yellow underline"))
	println(Color(BlueUnderline, "blue underline"))
	println(Color(PurpleUnderline, "purple underline"))
	println(Color(CyanUnderline, "cyan underline"))
	println(Color(WhiteUnderline, "white underline"))
	println(Color(BlackBackground, "black background"))
	println(Color(RedBackground, "red background"))
	println(Color(GreenBackground, "green background"))
	println(Color(YellowBackground, "yellow background"))
	println(Color(BlueBackground, "blue background"))
	println(Color(PurpleBackground, "purple background"))
	println(Color(CyanBackground, "cyan background"))
	println(Color(WhiteBackground, "white background"))
	println(Color(BlackIntense, "black intense"))
	println(Color(RedIntense, "red intense"))
	println(Color(GreenIntense, "green intense"))
	println(Color(YellowIntense, "yellow intense"))
	println(Color(BlueIntense, "blue intense"))
	println(Color(PurpleIntense, "purple intense"))
	println(Color(CyanIntense, "cyan intense"))
	println(Color(WhiteIntense, "white intense"))
	println(Color(BlackBoldIntense, "black bold intense"))
	println(Color(RedBoldIntense, "red bold intense"))
	println(Color(GreenBoldIntense, "green bold intense"))
	println(Color(YellowBoldIntense, "yellow bold intense"))
	println(Color(BlueBoldIntense, "blue bold intense"))
	println(Color(PurpleBoldIntense, "purple bold intense"))
	println(Color(CyanBoldIntense, "cyan bold intense"))
	println(Color(WhiteBoldIntense, "white bold intense"))
	println(Color(BlackBackgroundIntense, "black background intense"))
	println(Color(RedBackgroundIntense, "red background intense"))
	println(Color(GreenBackgroundIntense, "green background intense"))
	println(Color(YellowBackgroundIntense, "yellow background intense"))
	println(Color(BlueBackgroundIntense, "blue background intense"))
	println(Color(PurpleBackgroundIntense, "purple background intense"))
	println(Color(CyanBackgroundIntense, "cyan background intense"))
	println(Color(WhiteBackgroundIntense, "white background intense"))
	println(Reset)
}

func Color(color string, text string) string {
	return color + text + Reset
}
