package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func showMacVoices() {
	result, err := exec.Command("say", "-v", "?").Output()
	if err != nil {
		log.Fatalf("showMacVoices err=%v", err)
	}
	fmt.Println(string(result))
}

func speakMac(cfg config, l string) ([]byte, error) {
	cmd := exec.Command("say", "-v", cfg.voice, "-o", "/tmp/twna.wav", "--data-format=LEF32@32000", "-r",
		fmt.Sprintf("%d", int(170*cfg.speed)), l)
	_, err := cmd.Output()
	if err != nil {
		log.Fatalf("speakMac err=%v", err)
	}
	defer os.Remove("/tmp/twna.wav")
	return os.ReadFile("/tmp/twna.wav")
}

func getConfigMac(l string) config {
	ret := config{
		speaker:    0,
		style:      0,
		speed:      1.0,
		intonation: 1.0,
		volume:     1.0,
		pitch:      0.0,
		voice:      "Alex",
	}
	l = strings.ReplaceAll(l, "#", "")
	p := strings.Split(l, ",")
	if len(p) < 2 {
		return ret
	}
	ret.voice = p[0]
	if len(p) > 2 {
		if v, err := strconv.ParseFloat(p[2], 64); err == nil {
			ret.speed = v
		}
	}
	return ret

}
