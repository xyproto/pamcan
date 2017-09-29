// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"syscall"
	"unsafe"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// getWidth returns the width of the current terminal, thanks https://stackoverflow.com/a/16576712/131264
func getWidth() uint {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return uint(ws.Col)
}

const (
	// 140 characters, 70 pixels wide image of a Pam Can
	image140 = `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`

	// 80 characters, 40 pixels wide image of a Pam Can
	image80 = `H4sIABgazlkCA+2aTXbrIAyF595CJl0CEuLHp0vpGrr/aYfp8VUiWQ87zTPTe8Ak8CEuQrcv6Z/lk+r3x8ftS9bv2xSm8PeEZU7I9QTOBZR1q0gCJdsKfifnyesU9uHJ2yYdYGRAj5qtYC+WrVIKjN4BfPiy8GPwJ7RXpJi3zFLroGBwbLbSgFlKIfZh9E6T2TeJiXiwasd6N+NUNylVSM74+5qprBBYM9kbAhU4DH4py45DZSIWQEwciKFPhLEEepXtqjaHt9z+GhLkZ9uLEjmO8+SYHTY3oI9w2dIbgfVVgPsmhgfBZke4soZ4bCGlOmyguSws8K8KIFrY3kLKphLzT5TVMe2yh+tlMBkHIjcKQkesQjIULsU4nLGFAo9DKSOu6crUKAQiydmmtBS7F5CMgX2VwIrTfTWXJ1h4FA71GoVgd6R0OKTkUJuYkuxDth6ELjvQZcO1KI4kN9OjMJNty5Uvg5kXxVtIIOD7GKw2g65EIw1S+DxFMaflFIemrAzAw8whhWwsuziyHI47pid7R0pgXh0L0QdhGVEUO0qnKcUzF+XPXCTqiSCnIdi6UnrPknzLkxhDYyD1bJGj0IYmxTG0YmXKsa9Kx+aBKQAodXA5aTstbTtzVcwtlW2Aoc9jV708uZbwQQH1QHo9WyfC5lvnHmEhFRC7KYBN3kJVbTepGWcMv7xnOZaHXrocFmtjijJWn3DaAlIE+BKBhy3J8VbWHA8TNUCs4o6V21f2WIgYoH1Qr+a4IF4eUFMo3XYUiusAGFK26azJNq8Z2lSbcsjE3R8Dl120vpZNOSah/Z+/K4KSEE5YMyo2wPi2hyiiMgbg+7ZcTkAxBucMpUc8b5LY4dfBs1bqM4bebLGL9ZS+q46HwjFczjB6iIDOUgG1hkIx286BxcZbQvC20wLmRPVFAr7lIrpYzOsoE9ZML9voKsDzjufn5R+5nCH0jSwEotvt+xp+B+oklIwDWuLYDQ7aaH53hsuLkYzcKm6hmtXprip3DLmKf0iO2vg2pjBtwvC+ApgJ09Ay0RiMldyC2J4EK9onotcSYlkzTx5NyV9IIKwjx78r2X8ASH7HgwE/AAA=`

	// 50 characters, 25 pixels wide image of a Pam Can
	image50 = `H4sIAC8gzlkCA+2ZTY7DMAiF97lCNz2CAeMf9Sg9Q++/nV0U+dmBuo06ncmyT8Gy+QCDe7nHctMbpcf1ernH+ricgktgqY1StbWR0hpFASXAwmAl68rL6XwvHvgkV1sRgm9iuzJlU6ES/jIxFh6H6DCwKbRWFB1WEawk22klgA2sKOUx7GXnqLYy50JUxKGoXT/AhdxxPDo12et09tOuY/1mJVCeqokrPEF4MuVSTzSpeXC8ITiSGbnF+O27QtSOP4FEg0MqfFN2Kt26Dljl9hBaxzm97DjUoxQHhuAAw1MKOfYTj7jemNmhwPY4mLB6SrUvvJ53wkGKejItOIInvqG0e+iJSYYK5HCAUtoaJWxBYTMCxJGcp8TUw5RPYZm5tREKEeS3OmoktieZTZib/SzDVOh4uEwp6ddymSFXsbEHo+AYGVKylczjIF+e4pSnqMjXcLIFLSalDjfs9T3cwLnKnjblPUi+GRKNHbdmm5r5t2npX6K2IbKcSJ4pmtEk0qGGQ8IUxzo1jP/7dy82G75OwaQwUzA7wZBtZudTpSUgsU6xnFE603rv1eQk9mpPgtMbTuO7U/T4QVro/DPgiOcTMoHBuA0jBOLaNIw/L8w0kWsbAAA=`

	// 40 characters, 20 pixels wide image of a Pam Can
	image40 = `H4sIAOkazlkCA+WYTXLDMAiF976CNzmCQCChyVFyhtx/213b4UmBpJ3Jj5dhsDCfntBz9ovYWc/UrqfTfpFx3Q8Q4Ko+ItU/VC3MYSmw8net7RgkxacU8pHuSQ6FFECLtY1eBC1XDjHwRD4EEZQh1hqxVAWqi8XvU9t6F7YbrWYiNQEo06pf2e78jW1PdrP7iEJDCg1pgV3wOUOWut/WxSdcLEGqJOTGiQjFeGC/E2cIV2GozVCbSzhLqA8YFAUwTxot/xLRcc+d8BeCYQDkRgZzlnxO98wVAfuX/TkeN/iiRFMRS8CrT5n6CMbaWnlrdcJJQMATF5Fj1WOjMxnN/CpeAjiUEUbQOlAtsYQtM4k1cYvx+5o51N3kIVBrw9neEnTlAR/xxsYYhjGTrg3tWssT2vSISfvwzzccFBQPgZx1+GiUlLhdwUWgUZuRpFi4R/vXISPA+PME77dfhuwLRaaUW9ARAAA=`

	// 30 characters, 15 pixels wide image of a Pam Can
	image30 = `H4sIACogzlkCA81WORLDMAjs/QU1foKEQMf4KX6D/9+mSpzR4qB47ETtjmDYZQG5lcsiS0jbPLuV6+YGByhyg1RpY2JpgzhCmvxEph/XT4BgcRhV28Q+AO1sKwNBmVsdtArpFAsBBBsTAOGO5iU7j0CF4i19dmD60Ipk0pLawRMRbyqo2AmeoFxE0PX6DfISg6mj5nrmjeIVumA3KFMS8E2GMWkzS0EtvO1umMhreJ4AwOoU0FugQ4I0yrbo4RjHOB5SLlFBcYOy/WjYCwqcczw+C8e67MdkGp50z65AJ6M7lDXJ//lI3PMZQRFCEVuot9/UA7PgQH5SCgAA`
)

// Decompress text that has first been gzipped and then base64 encoded
func decompressImage(asciigfx string) string {
	unbasedBytes, err := base64.StdEncoding.DecodeString(asciigfx)
	if err != nil {
		panic("Could not decode base64: " + err.Error())
	}
	buf := bytes.NewBuffer(unbasedBytes)
	decompressorReader, err := gzip.NewReader(buf)
	if err != nil {
		panic("Could not read buffer: " + err.Error())
	}
	decompressedBytes, err := ioutil.ReadAll(decompressorReader)
	decompressorReader.Close()
	if err != nil {
		panic("Could not decompress: " + err.Error())
	}
	return string(decompressedBytes)
}

func main() {
	if getWidth() >= 140 {
		fmt.Println(decompressImage(image140))
	} else if getWidth() >= 80 {
		fmt.Println(decompressImage(image80))
	} else if getWidth() >= 50 {
		fmt.Println(decompressImage(image50))
	} else if getWidth() >= 40 {
		fmt.Println(decompressImage(image40))
	} else if getWidth() >= 30 {
		fmt.Println(decompressImage(image30))
	} else {
		fmt.Println("PAM CAN!")
	}
}
