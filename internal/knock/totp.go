package knock

import (
	"context"
	"encoding/base32"
	"time"

	"github.com/bzp2010/knockdoor/internal/config"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/pquerna/otp/totp"
	"golang.org/x/net/ipv4"
)

type totpKnock struct {
	cfg config.PortTOTP

	portSerial []string

	fsm  *fsm.FSM
	new  bool
	done bool
}

// NewTOTPKnock creates a new TOTP knock
func NewTOTPKnock(cfg config.PortTOTP, doneCallback func()) Knock {
	code, _ := totp.GenerateCode(
		base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(cfg.Secret)),
		time.Now(),
	)

	portSerial := []string{cfg.Prefix + code[0:1], cfg.Prefix + code[1:2], cfg.Prefix + code[2:3], cfg.Prefix + code[3:4], cfg.Prefix + code[4:5], cfg.Prefix + code[5:6]}

	return &totpKnock{
		cfg: cfg,
		fsm: fsm.NewFSM("new_visitor", fsm.Events{
			{Name: "knock", Src: []string{"new_visitor"}, Dst: "STAGE_" + portSerial[0]},
			{Name: "knock_" + portSerial[1], Src: []string{"STAGE_" + portSerial[0]}, Dst: "STAGE_" + portSerial[1]},
			{Name: "knock_" + portSerial[2], Src: []string{"STAGE_" + portSerial[1]}, Dst: "STAGE_" + portSerial[2]},
			{Name: "knock_" + portSerial[3], Src: []string{"STAGE_" + portSerial[2]}, Dst: "STAGE_" + portSerial[3]},
			{Name: "knock_" + portSerial[4], Src: []string{"STAGE_" + portSerial[3]}, Dst: "STAGE_" + portSerial[4]},
			{Name: "knock_" + portSerial[5], Src: []string{"STAGE_" + portSerial[4]}, Dst: "STAGE_OPEN_DOOR"},
		}, fsm.Callbacks{
			"enter_state": func(_ context.Context, e *fsm.Event) {
				if e.Dst == "STAGE_OPEN_DOOR" {
					doneCallback()
				}
			},
		}),
		new:        true,
		portSerial: portSerial,
	}
}

// Handle handles the TOTP knock
func (t *totpKnock) Handle(ipHeader *ipv4.Header, tcpPacket *layers.TCP) bool {
	if t.new {
		if portToString(tcpPacket.DstPort) != t.portSerial[0] {
			return true
		}
		t.new = false
		t.fsm.Event(context.Background(), "knock")
	} else {
		t.fsm.Event(context.Background(), "knock_"+portToString(tcpPacket.DstPort))
	}
	return t.done
}
