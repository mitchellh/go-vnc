/*
client_events.go provides constants for making VNC keyboard and mouse events.

Sample usage:

// Move mouse to x=100, y=200.
x, y := 100, 200
conn.PointerEvent(vnc.MouseNone, x, y)
// Give mouse some time to "settle."
time.Sleep(10*time.Millisecond)
// Left click.
conn.PointerEvent(vnc.MouseLeft, x, y)
conn.PointerEvent(vnc.MouseNone, x, y)

// Press return key
conn.KeyEvent(vnc.KeyReturn, true)
// Release the key.
conn.KeyEvent(vnc.KeyReturn, false)
*/

package vnc

// Latin 1 (byte 3 = 0)
// ISO/IEC 8859-1 = Unicode U+0020..U+00FF
const (
	KeySpace = iota + 0x0020
	KeyExclam
	KeyQuoteDbl
	KeyNumberSign
	KeyDollar
	KeyPercent
	KeyAmpersand
	KeyApostrophe
	KeyParenLeft
	KeyParenRight
	KeyAsterisk
	KeyPlus
	KeyComma
	KeyMinus
	KeyPeriod
	KeySlash
	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	KeyColon
	KeySemicolon
	KeyLess
	KeyEqual
	KeyGreater
	KeyQuestion
	KeyAt
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyBracketLeft
	KeyBackslash
	KeyBracketRight
	KeyAsciiCircum
	KeyUnderscore
	KeyGrave
	Keya
	Keyb
	Keyc
	Keyd
	Keye
	Keyf
	Keyg
	Keyh
	Keyi
	Keyj
	Keyk
	Keyl
	Keym
	Keyn
	Keyo
	Keyp
	Keyq
	Keyr
	Keys
	Keyt
	Keyu
	Keyv
	Keyw
	Keyx
	Keyy
	Keyz
	KeyBraceLeft
	KeyBar
	KeyBraceRight
	KeyAsciiTilde
)
const (
	KeyBackspace = iota + 0xff08
	KeyTab
	KeyLinefeed
	KeyClear
	_
	KeyReturn
)
const (
	KeyPause      = 0xff13
	KeyScrollLock = 0xff14
	KeySysReq     = 0xff15
	KeyEscape     = 0xff1b
)
const (
	KeyF1 = iota + 0xffbe
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)
const (
	KeyShiftLeft = iota + 0xffe1
	KeyShiftRight
	KeyControlLeft
	KeyControlRight
	KeyCapsLock
	_
	_
	_
	KeyAltLeft
	KeyAltRight

	KeyDelete = 0xffff
)
const (
	// Mouse buttons
	MouseLeft = 1 << iota
	MouseMiddle
	MouseRight
	MouseNone = 0
)
