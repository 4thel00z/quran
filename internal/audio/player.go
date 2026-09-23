package audio

import (
	"bytes"
	"math"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

const (
	outputRate = beep.SampleRate(44100)
	// bufferTime is the speaker buffer; Position subtracts it so the
	// reported time matches what is audible.
	bufferTime = 60 * time.Millisecond
)

// Player plays one ayah at a time through the default output device.
type Player struct {
	initOnce sync.Once
	initErr  error

	stream beep.StreamSeekCloser
	format beep.Format
	ctrl   *beep.Ctrl
	volume *effects.Volume
	level  int
	// finish closes the channel Play returned; Stop calls it too, so
	// waiters are released when playback is cut short.
	finish func()
}

// MaxLevel is full volume; each step below halves loudness by a quarter octave.
const MaxLevel = 10

func NewPlayer() *Player {
	return &Player{level: MaxLevel}
}

func (p *Player) init() error {
	p.initOnce.Do(func() {
		p.initErr = speaker.Init(outputRate, outputRate.N(bufferTime))
	})
	return p.initErr
}

// Play starts data and returns a channel closed when it finishes playing or
// is stopped.
func (p *Player) Play(data []byte) (<-chan struct{}, error) {
	if err := p.init(); err != nil {
		return nil, err
	}
	stream, format, err := mp3.Decode(memoryFile{bytes.NewReader(data)})
	if err != nil {
		return nil, err
	}
	p.Stop()
	var source beep.Streamer = stream
	if format.SampleRate != outputRate {
		source = beep.Resample(4, format.SampleRate, outputRate, stream)
	}
	done := make(chan struct{})
	finish := sync.OnceFunc(func() { close(done) })
	speaker.Lock()
	p.stream, p.format = stream, format
	p.volume = &effects.Volume{Streamer: source, Base: 2}
	p.applyLevel()
	p.ctrl = &beep.Ctrl{Streamer: p.volume}
	p.finish = finish
	speaker.Unlock()
	speaker.Play(beep.Seq(p.ctrl, beep.Callback(finish)))
	return done, nil
}

func (p *Player) Stop() {
	if p.stream == nil {
		return
	}
	speaker.Clear()
	speaker.Lock()
	stream, finish := p.stream, p.finish
	p.stream, p.ctrl, p.volume, p.finish = nil, nil, nil, nil
	speaker.Unlock()
	finish()
	_ = stream.Close()
}

// TogglePause reports whether playback is now paused.
func (p *Player) TogglePause() bool {
	if p.ctrl == nil {
		return false
	}
	speaker.Lock()
	defer speaker.Unlock()
	p.ctrl.Paused = !p.ctrl.Paused
	return p.ctrl.Paused
}

func (p *Player) Paused() bool {
	if p.ctrl == nil {
		return false
	}
	speaker.Lock()
	defer speaker.Unlock()
	return p.ctrl.Paused
}

// Position is the audible offset into the current ayah.
func (p *Player) Position() time.Duration {
	if p.stream == nil {
		return 0
	}
	speaker.Lock()
	position := p.format.SampleRate.D(p.stream.Position())
	speaker.Unlock()
	return max(0, position-bufferTime)
}

func (p *Player) Duration() time.Duration {
	if p.stream == nil {
		return 0
	}
	speaker.Lock()
	defer speaker.Unlock()
	return p.format.SampleRate.D(p.stream.Len())
}

func (p *Player) Level() int {
	return p.level
}

// SetLevel sets volume from 0 (silent) to MaxLevel.
func (p *Player) SetLevel(level int) {
	p.level = min(MaxLevel, max(0, level))
	if p.volume == nil {
		return
	}
	speaker.Lock()
	p.applyLevel()
	speaker.Unlock()
}

func (p *Player) applyLevel() {
	p.volume.Silent = p.level == 0
	p.volume.Volume = -float64(MaxLevel-p.level) / 4
}

// Percent is the volume level as a percentage of full loudness.
func (p *Player) Percent() int {
	if p.level == 0 {
		return 0
	}
	return int(math.Round(100 * math.Pow(2, -float64(MaxLevel-p.level)/4)))
}


// memoryFile lets the decoder seek, which it needs to report Len.
type memoryFile struct {
	*bytes.Reader
}

func (memoryFile) Close() error {
	return nil
}
