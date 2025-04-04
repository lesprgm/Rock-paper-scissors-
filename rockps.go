package main

import (
	"bytes"
	"image"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/wav"
	"github.com/faiface/beep/speaker"
	"golang.org/x/image/draw"
)

func loadImage(path string, width, height int) fyne.Resource {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	resizedImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), draw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, resizedImg); err != nil {
		panic(err)
	}

	return fyne.NewStaticResource(filepath.Base(path), buf.Bytes())
}

type Game struct {
	resultLabel         *widget.Label
	playerChoiceLabel   *widget.Label
	computerChoiceLabel *widget.Label
	computerChoiceImage *canvas.Image
	rockButton          *widget.Button
	paperButton         *widget.Button
	scissorsButton      *widget.Button
	playAgainButton     *widget.Button
	options             []string
	random              *rand.Rand
	rockRes             fyne.Resource
	paperRes            fyne.Resource
	scissorsRes         fyne.Resource
	winSound            *beep.Buffer
	loseSound           *beep.Buffer
}

func (g *Game) handleChoice(playerChoice string) {
	g.rockButton.Disable()
	g.paperButton.Disable()
	g.scissorsButton.Disable()

	computerChoice := g.options[g.random.Intn(len(g.options))]

	g.playerChoiceLabel.SetText("Your choice: " + playerChoice)
	g.computerChoiceLabel.SetText("Computer's choice: " + computerChoice)

	switch computerChoice {
	case "rock":
		g.computerChoiceImage.Resource = g.rockRes
	case "paper":
		g.computerChoiceImage.Resource = g.paperRes
	case "scissors":
		g.computerChoiceImage.Resource = g.scissorsRes
	}
	g.computerChoiceImage.Refresh()

	result := determineResult(playerChoice, computerChoice)
	g.resultLabel.SetText("Result: " + result)

	if result == "You win!" {
		g.playSound(g.winSound)
	} else if result == "You lose!" {
		g.playSound(g.loseSound)
	}

	g.playAgainButton.Enable()
}

func determineResult(player, computer string) string {
	if player == computer {
		return "It's a tie!"
	}
	switch player {
	case "rock":
		if computer == "scissors" {
			return "You win!"
		}
	case "paper":
		if computer == "rock" {
			return "You win!"
		}
	case "scissors":
		if computer == "paper" {
			return "You win!"
		}
	}
	return "You lose!"
}

func (g *Game) resetGame() {
	g.rockButton.Enable()
	g.paperButton.Enable()
	g.scissorsButton.Enable()
	g.playAgainButton.Disable()
	g.playerChoiceLabel.SetText("Your choice: ")
	g.computerChoiceLabel.SetText("Computer's choice: ")
	g.computerChoiceImage.Resource = nil
	g.computerChoiceImage.Refresh()
	g.resultLabel.SetText("Result: ")
}

func (g *Game) playSound(buffer *beep.Buffer) {
	sound := buffer.Streamer(0, buffer.Len())
	speaker.Play(sound)
}

func main() {
	a := app.New()
	w := a.NewWindow("Rock, Paper, Scissors")
	w.Resize(fyne.NewSize(600, 400))

	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, int(sampleRate/10)) 

	winSound, err := loadSound("win.wav")
	if err != nil {
		panic(err)
	}
	loseSound, err := loadSound("lose.wav")
	if err != nil {
		panic(err)
	}

	rockRes := loadImage("rock.png", 150, 150)
	paperRes := loadImage("paper.png", 150, 150)
	scissorsRes := loadImage("scissors.png", 150, 150)

	instructionLabel := widget.NewLabel("Choose Rock, Paper, or Scissors:")
	resultLabel := widget.NewLabel("Result: ")
	computerChoiceLabel := widget.NewLabel("Computer's choice: ")
	playerChoiceLabel := widget.NewLabel("Your choice: ")
	computerChoiceImage := canvas.NewImageFromResource(nil)
	computerChoiceImage.SetMinSize(fyne.NewSize(150, 150))

	rockButton := widget.NewButtonWithIcon("", rockRes, nil)
	paperButton := widget.NewButtonWithIcon("", paperRes, nil)
	scissorsButton := widget.NewButtonWithIcon("", scissorsRes, nil)

	rockContainer := container.NewGridWrap(fyne.NewSize(160, 160), rockButton)
	paperContainer := container.NewGridWrap(fyne.NewSize(160, 160), paperButton)
	scissorsContainer := container.NewGridWrap(fyne.NewSize(160, 160), scissorsButton)

	playAgainButton := widget.NewButton("Play Again", nil)
	playAgainButton.Disable()

	game := &Game{
		resultLabel:         resultLabel,
		playerChoiceLabel:   playerChoiceLabel,
		computerChoiceLabel: computerChoiceLabel,
		computerChoiceImage: computerChoiceImage,
		rockButton:          rockButton,
		paperButton:         paperButton,
		scissorsButton:      scissorsButton,
		playAgainButton:     playAgainButton,
		options:             []string{"rock", "paper", "scissors"},
		random:              rand.New(rand.NewSource(time.Now().UnixNano())),
		rockRes:             rockRes,
		paperRes:            paperRes,
		scissorsRes:         scissorsRes,
		winSound:            winSound,
		loseSound:           loseSound,
	}

	rockButton.OnTapped = func() { game.handleChoice("rock") }
	paperButton.OnTapped = func() { game.handleChoice("paper") }
	scissorsButton.OnTapped = func() { game.handleChoice("scissors") }
	playAgainButton.OnTapped = func() { game.resetGame() }

	top := container.NewVBox(instructionLabel)
	buttons := container.NewHBox(rockContainer, paperContainer, scissorsContainer)
	computerChoiceContainer := container.NewHBox(computerChoiceLabel, computerChoiceImage)
	bottom := container.NewVBox(
		playerChoiceLabel,
		computerChoiceContainer,
		resultLabel,
		playAgainButton,
	)

	content := container.NewBorder(top, bottom, nil, nil, buttons)
	w.SetContent(content)

	w.ShowAndRun()
}

func loadSound(path string) (*beep.Buffer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	streamer, format, err := wav.Decode(file)
	if err != nil {
		return nil, err
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	return buffer, nil
}