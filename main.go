package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-audio/wav"
	"github.com/hajimehoshi/oto"
)

type Params struct {
	AccentPhrases      []AccentPhrases `json:"accent_phrases"`
	SpeedScale         float64         `json:"speedScale"`
	PitchScale         float64         `json:"pitchScale"`
	IntonationScale    float64         `json:"intonationScale"`
	VolumeScale        float64         `json:"volumeScale"`
	PrePhonemeLength   float64         `json:"prePhonemeLength"`
	PostPhonemeLength  float64         `json:"postPhonemeLength"`
	OutputSamplingRate int             `json:"outputSamplingRate"`
	OutputStereo       bool            `json:"outputStereo"`
	Kana               string          `json:"kana"`
}

type Mora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant"`
	ConsonantLength *float64 `json:"consonant_length"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowel_length"`
	Pitch           float64  `json:"pitch"`
}

type AccentPhrases struct {
	Moras           []Mora `json:"moras"`
	Accent          int    `json:"accent"`
	PauseMora       *Mora  `json:"pause_mora"`
	IsInterrogative bool   `json:"is_interrogative"`
}

type Speaker struct {
	Name        string   `json:"name"`
	SpeakerUUID string   `json:"speaker_uuid"`
	Styles      []Styles `json:"styles"`
	Version     string   `json:"version"`
}

type Styles struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var speakers = []Speaker{}
var url = "http://localhost:50021"
var script = ""
var list = false
var play = false
var version = "vx.x.x"
var commit = ""

func getSpeakers() {
	resp, err := http.Get(url + "/speakers")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&speakers); err != nil {
		log.Fatal(err)
	}
}

func getQuery(id int, text string) (*Params, error) {
	req, err := http.NewRequest("POST", url+"/audio_query", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))
	q.Add("text", text)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var params *Params
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return nil, err
	}
	return params, nil
}

func synth(id int, params *Params) ([]byte, error) {
	b, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url+"/synthesis", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "audio/wav")
	req.Header.Add("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("speaker", strconv.Itoa(id))
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buff := bytes.NewBuffer(nil)
	if _, err := io.Copy(buff, resp.Body); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func playback(params *Params, b []byte) error {
	ch := 1
	if params.OutputStereo {
		ch = 2
	}
	ctx, err := oto.NewContext(params.OutputSamplingRate, ch, 2, 3200)
	if err != nil {
		return err
	}
	defer ctx.Close()
	p := ctx.NewPlayer()
	if _, err := io.Copy(p, bytes.NewReader(b)); err != nil {
		return err
	}
	if err := p.Close(); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.StringVar(&url, "url", "http://localhost:50021", "api url")
	flag.StringVar(&script, "s", "", "input script(txt or pptx")
	flag.BoolVar(&list, "l", false, "list speaker")
	flag.BoolVar(&play, "p", false, "play")
	flag.Parse()
	getSpeakers()
	if list {
		showSpeakers()
		return
	}
	log.Printf("version=%s(%s)", version, commit)
	st := time.Now()
	playScript(script)
	log.Printf("time=%v", time.Since(st))
}

type config struct {
	speaker    int
	style      int
	speed      float64
	intonation float64
	volume     float64
	pitch      float64
}

func playScript(file string) {
	scripts, err := readScript(file)
	if err != nil {
		log.Fatalf("readScript err=%v", err)
	}
	var outwav *wav.Encoder
	var out *os.File

	cfg := getConfig("")
	for slide, lines := range scripts {
		for i, l := range lines {
			l = strings.TrimSpace(l)
			log.Printf("%d:%d %s\n", slide, i+1, l)
			if strings.HasPrefix(l, "#") {
				cfg = getConfig(l)
			} else if l != "" {
				if b, err := speak(cfg, l); err != nil {
					log.Fatalln(err)
				} else {
					if !play {
						r := bytes.NewReader(b)
						d := wav.NewDecoder(r)
						buf, err := d.FullPCMBuffer()
						if err != nil {
							log.Fatalln(err)
						}
						if outwav == nil {
							wavfile := getWavFileName(file, slide)
							out, err := os.Create(wavfile)
							if err != nil {
								log.Fatalln(err)
							}
							outwav = wav.NewEncoder(out, buf.Format.SampleRate, int(d.BitDepth), buf.Format.NumChannels, int(d.WavAudioFormat))
						}
						if err := outwav.Write(buf); err != nil {
							log.Fatalln(err)
						}
					}
				}
			}
		}
		if outwav != nil {
			outwav.Close()
			outwav = nil
		}
		if out != nil {
			out.Close()
			out = nil
		}
	}
}

func speak(cfg config, l string) ([]byte, error) {
	spk := speakers[cfg.speaker]
	spkID := spk.Styles[cfg.style].ID
	params, err := getQuery(spkID, l)
	if err != nil {
		log.Fatal(err)
	}
	params.SpeedScale = cfg.speed
	params.PitchScale = cfg.pitch
	params.IntonationScale = cfg.intonation
	params.VolumeScale = cfg.volume
	b, err := synth(spkID, params)
	if err != nil {
		return b, err
	}
	if play {
		return b, playback(params, b[44:])
	}
	return b, nil
}

func getConfig(l string) config {
	ret := config{
		speaker:    0,
		style:      0,
		speed:      1.0,
		intonation: 1.0,
		volume:     1.0,
		pitch:      0.0,
	}
	l = strings.ReplaceAll(l, "#", "")
	p := strings.Split(l, ",")
	if len(p) < 2 {
		return ret
	}
	speaker, style, err := findSpeaker(strings.TrimSpace(p[0]), strings.TrimSpace(p[1]))
	if err != nil {
		log.Println(err)
		return ret
	}
	ret.speaker = speaker
	ret.style = style
	if len(p) < 6 {
		return ret
	}
	if v, err := strconv.ParseFloat(p[2], 64); err == nil {
		ret.speed = v
	}
	if v, err := strconv.ParseFloat(p[3], 64); err == nil {
		ret.intonation = v
	}
	if v, err := strconv.ParseFloat(p[4], 64); err == nil {
		ret.volume = v
	}
	if v, err := strconv.ParseFloat(p[5], 64); err == nil {
		ret.pitch = v
	}
	return ret
}

func showSpeakers() {
	for _, s := range speakers {
		for _, t := range s.Styles {
			fmt.Printf("%s,%s\n", s.Name, t.Name)
		}
	}
}

func findSpeaker(name, style string) (int, int, error) {
	for i, s := range speakers {
		if name == s.Name {
			for j, t := range s.Styles {
				if style == t.Name {
					return i, j, nil
				}
			}
		}
	}
	return -1, -1, fmt.Errorf("speaker not found name=%s style=%s", name, style)
}

func readScript(file string) (map[int][]string, error) {
	ext := filepath.Ext(file)
	if ext == ".pptx" {
		return readScriptFromPPTX(file)
	}
	return readScriptFromTxt(file)
}

func readScriptFromTxt(file string) (map[int][]string, error) {
	ret := make(map[int][]string)
	b, err := os.ReadFile(file)
	if err != nil {
		return ret, err
	}
	a := strings.Split(string(b), "$\n")
	for i, l := range a {
		ret[i+1] = strings.Split(l, "\n")
	}
	return ret, nil
}

var noteReg = regexp.MustCompile(`ppt/notesSlides/notesSlide(\d+).xml`)
var textReg = regexp.MustCompile(`<a:t>([^<]+)</a:t>`)
var numOnlyReg = regexp.MustCompile(`^(/d+)$`)

func readScriptFromPPTX(file string) (map[int][]string, error) {
	ret := make(map[int][]string)
	reader, err := zip.OpenReader(file)
	if err != nil {
		return ret, err
	}
	for _, file := range reader.File {
		fm := noteReg.FindAllStringSubmatch(file.Name, 1)
		if len(fm) > 0 && len(fm[0]) > 1 {
			slide, err := strconv.Atoi(fm[0][1])
			if err != nil {
				continue
			}
			log.Printf("%s:%d\n", file.Name, slide)
			r, err := file.Open()
			if err != nil {
				return ret, err
			}
			defer r.Close()
			b, err := io.ReadAll(r)
			if err != nil {
				return ret, err
			}
			m := textReg.FindAllStringSubmatch(string(b), -1)
			for _, e := range m {
				if len(e) > 1 && !numOnlyReg.MatchString(e[1]) {
					ret[slide] = append(ret[slide], e[1])
				}
			}
		}
	}
	return ret, nil
}

func getWavFileName(file string, slide int) string {
	ext := filepath.Ext(file)
	base := filepath.Base(file)
	dir := filepath.Dir(file)
	add := fmt.Sprintf("-%d.wav", slide)
	return filepath.Join(dir, strings.Replace(base, ext, add, 1))
}
