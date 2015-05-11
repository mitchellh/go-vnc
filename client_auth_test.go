package vnc

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestClientAuthNone_Impl(t *testing.T) {
	var raw interface{}
	raw = new(ClientAuthNone)
	if _, ok := raw.(ClientAuth); !ok {
		t.Fatal("ClientAuthNone doesn't implement ClientAuth")
	}
}

func TestClientAuthVNC_Impl(t *testing.T) {
	var raw interface{}
	raw = new(ClientAuthVNC)
	if _, ok := raw.(ClientAuth); !ok {
		t.Fatal("ClientAuthVNC doesn't implement ClientAuth")
	}
}

// wiresharkToChallenge converts VNC authentication challenge and response
// values captured with Wireshark (https://www.wireshark.org) into usable byte
// streams.
func wiresharkToChallenge(h string) [challengeSize]byte {
	var c [challengeSize]byte
	r := strings.NewReplacer(":", "")
	b, err := hex.DecodeString(r.Replace(h))
	if err != nil {
		return c
	}
	copy(c[:], b)
	return c
}

func TestClientAuthVNCEncode(t *testing.T) {
	tests := []struct {
		pw                  string
		challenge, response string
	}{
		{".", "7f:e2:e1:3d:a4:ae:10:9c:54:c5:5f:52:74:aa:db:31", "1d:86:92:71:1f:00:24:35:02:d3:91:ef:e9:bc:c5:d5"},
		{"12345678", "13:8e:a4:2e:0e:66:f3:ad:2d:f3:08:c3:04:cd:c4:2a", "5b:e1:56:fa:49:49:ef:56:d3:f8:44:97:73:27:95:9f"},
		{"abc123", "c6:30:45:d2:57:9e:e7:f2:f9:0c:62:3e:52:40:86:c6", "a3:63:59:e4:28:c8:7f:b3:45:2c:d7:e0:ca:d6:70:3e"},
	}

	for _, tt := range tests {
		challenge := wiresharkToChallenge(tt.challenge)
		a := ClientAuthVNC{tt.pw}
		if err := a.encode(&challenge); err != nil {
			t.Errorf("ClientAuthVNC.encode() failed: key=%v, err=%v", tt.pw, err)
		}
		response := wiresharkToChallenge(tt.response)
		if challenge != response {
			t.Errorf("ClientAuthVNC.encode() failed: key=%v got=%v, want=%v", tt.pw, challenge, response)
		}
	}
}
