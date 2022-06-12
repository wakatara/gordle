package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	title  string = "Gordle"
	width  int    = 435
	height int    = 600
	rows   int    = 6
	cols   int    = 5
)

var (
	fontSize        float64 = 24.0
	mplusNormalFont font.Face

	bkg       = color.White
	lightgrey = color.RGBA{0xc2, 0xc5, 0xc6, 0xff}
	grey      = color.RGBA{0x77, 0x7c, 0x7e, 0xff}
	yellow    = color.RGBA{0xcd, 0xb3, 0x5d, 0xff}
	green     = color.RGBA{0x60, 0xa6, 0x65, 0xff}
	fontColor = color.Black

	edge = false

	// Very hacky - replace later
	alphabet = "qwertyuiopasdfghjklzxcvbnm"

	grid   [cols * rows]string
	dict   []string
	check  [cols * rows]int
	loc    int = 0
	won        = false
	answer string
)

type Game struct {
	runes []rune
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return width, height
}

func (g Game) Update() error {
	if !won {
		g.runes = ebiten.AppendInputChars(g.runes[:0])
		if strings.Contains(alphabet, string(g.runes)) && string(g.runes) != "" && loc >= 0 && loc < rows*cols {
			grid[loc] = string(g.runes[0:1])
			if !edge {
				loc++
			}
		}
	}
	edge = false

	if (loc+2)%cols == 1 && loc != 0 {
		edge = true
	}

	if edge == true && (repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyNumpadEnter)) && grid[loc] != "" {
		inp := ""
		for i := (loc - (cols - 1)); i < (loc + 1); i++ {
			inp += grid[i]
		}
		validWord := false
		for _, w := range dict {
			if w == inp {
				validWord = true
			}
		}
		if validWord {
			var checkWord [cols]bool
			for i, letter := range inp {
				for j, ans := range answer {
					if i == j && string(letter) == string(ans) {
						check[loc-cols+i+1] = 1
					} else {
						check[loc-cols+i+1] = 3
					}
				}
			}
			for i, letter := range inp {
				if strings.Contains(answer, string(letter)) {
					found := false
					for j, ans := range answer {
						if found == false && checkWord[j] == false {
							if string(letter) == string(ans) {
								checkWord[j] = true
								found = true
								check[loc-cols+i+1] = 2
							}
						}
					}
				}
			}
			if inp == answer {
				won = true
			}
			loc++
			edge = false
		}
	}
	if loc < 0 {
		loc = 0
	}
	if loc > rows*cols {
		loc = rows*cols - 1
	}

	return nil
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)&interval == 0 {
		return true
	}
	return false
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(bkg)
	if won {
		winner := "Good job!"
		for i := 0; i < len(winner); i++ {
			msg := fmt.Sprintf(strings.ToUpper(string([]rune(winner)[i])))
			fontColor = color.Black
			text.Draw(screen, msg, mplusNormalFont, i*85+40, rows*85+55, fontColor)
		}
	}
	for w := 0; w < cols; w++ {
		for h := 0; h < rows; h++ {
			rect := ebiten.NewImage(75, 75)
			rect.Fill(lightgrey)
			fontColor = color.Black
			if check[w+(h*cols)] != 0 {
				if check[w+(h*cols)] == 1 {
					rect.Fill(green)
				}
				if check[w+(h*cols)] == 2 {
					rect.Fill(yellow)
				}
				if check[w+(h*cols)] == 3 {
					rect.Fill(grey)
				}
				fontColor = color.White
			}
			if w+cols*h == loc && check[w+(h*cols)] == 0 {
				rect.Fill(grey)
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(w*85+10)+1, float64(h*85+10)+1)
			screen.DrawImage(rect, op)
			if check[w+(h*cols)] == 0 {
				rect2 := ebiten.NewImage(73, 73)
				rect2.Fill(color.White)
				op2 := &ebiten.DrawImageOptions{}
				op2.GeoM.Translate(float64(w*85+10), float64(h*85+10))
				screen.DrawImage(rect2, op2)
			}
			if grid[w+(h*cols)] != "" {
				msg := fmt.Sprintf(strings.ToUpper(grid[w+(h*cols)]))
				text.Draw(screen, msg, mplusNormalFont, w*85+38, h*85+55, fontColor)
			}
		}
	}
	if !won && check[len(check)-1] != 0 {
		for i := 0; i < len(answer); i++ {
			msg := fmt.Sprintf(strings.ToUpper(string([]rune(answer)[i])))
			fontColor = color.Black
			text.Draw(screen, msg, mplusNormalFont, i*85+38, rows*85+55, fontColor)
		}
	}
}

func main() {
	g := &Game{}
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle(title)
	content, err := ioutil.ReadFile("res/dict.txt")
	if err != nil {
		log.Fatal(err)
	} else {
		dict = strings.Split(string(content), "\n")
	}
	rand.Seed(time.Now().UnixNano())
	answer = dict[rand.Intn(len(dict))]
	fmt.Printf(answer)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
