/*
client_events.go provides constants for making VNC keyboard and mouse events.

Sample usage:

// Move mouse to x=100, y=200.
x, y := 100, 200
conn.PointerEvent(vnc.Mouse_none, x, y)
// Give mouse some time to "settle."
time.Sleep(10*time.Millisecond)
// Left click.
conn.PointerEvent(vnc.Mouse_left, x, y)
conn.PointerEvent(vnc.Mouse_none, x, y)

// Press return key
conn.KeyEvent(vnc.Key_return, true)
// Release the key.
conn.KeyEvent(vnc.Key_return, false)
*/

package vnc

// Latin 1 (byte 3 = 0)
// ISO/IEC 8859-1 = Unicode U+0020..U+00FF
const (
	Key_space = iota + 0x0020
	Key_exclam
	Key_quotedbl
	Key_numbersign
	Key_dollar
	Key_percent
	Key_ampersand
	Key_apostrophe
	Key_parenleft
	Key_parenright
	Key_asterisk
	Key_plus
	Key_comma
	Key_minus
	Key_period
	Key_slash
	Key_0
	Key_1
	Key_2
	Key_3
	Key_4
	Key_5
	Key_6
	Key_7
	Key_8
	Key_9
	Key_colon
	Key_semicolon
	Key_less
	Key_equal
	Key_greater
	Key_question
	Key_at
	Key_A
	Key_B
	Key_C
	Key_D
	Key_E
	Key_F
	Key_G
	Key_H
	Key_I
	Key_J
	Key_K
	Key_L
	Key_M
	Key_N
	Key_O
	Key_P
	Key_Q
	Key_R
	Key_S
	Key_T
	Key_U
	Key_V
	Key_W
	Key_X
	Key_Y
	Key_Z
	Key_bracketleft
	Key_backslash
	Key_bracketright
	Key_asciicircum
	Key_underscore
	Key_grave
	Key_a
	Key_b
	Key_c
	Key_d
	Key_e
	Key_f
	Key_g
	Key_h
	Key_i
	Key_j
	Key_k
	Key_l
	Key_m
	Key_n
	Key_o
	Key_p
	Key_q
	Key_r
	Key_s
	Key_t
	Key_u
	Key_v
	Key_w
	Key_x
	Key_y
	Key_z
	Key_braceleft
	Key_bar
	Key_braceright
	Key_asciitilde
)
const (
	Key_backspace = iota + 0xff08
	Key_tab
	Key_linefeed
	Key_clear
	_
	Key_return
)
const (
	Key_pause       = 0xff13
	Key_scroll_lock = 0xff14
	Key_sys_req     = 0xff15
	Key_escape      = 0xff1b
)
const (
	Key_f1 = iota + 0xffbe
	Key_f2
	Key_f3
	Key_f4
	Key_f5
	Key_f6
	Key_f7
	Key_f8
	Key_f9
	Key_f10
	Key_f11
	Key_f12
)
const (
	Key_shift_l = iota + 0xffe1
	Key_shift_r
	Key_control_l
	Key_control_r
	Key_caps_lock
	_
	_
	_
	Key_alt_l
	Key_alt_r

	Key_delete = 0xffff
)
const (
	// Mouse buttons
	Mouse_left = 1 << iota
	Mouse_middle
	Mouse_right
	Mouse_none = 0
)
