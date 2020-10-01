package main

import (
	"github.com/tadeuszjt/data"
	"github.com/tadeuszjt/geom/32"
	"github.com/tadeuszjt/gfx"
    "github.com/tadeuszjt/neuralnetwork"
    "sync"
)

const (
	botsStart       = 100
	botsStartFed    = 100.
	botsChildFed    = 20.
	botsMaxFed      = 100.
	botsEatPlus     = 1.6
	botsFedBleed    = 0.3
	botsEatRadius   = 30
	botsSightRadius = 200
	botsSightWidth  = 2
	botsSpeed       = 1
	botsBreedOdds   = 1400
	botsNumSensors  = 8
)

var (
	arena = geom.RectCentred(2000, 2000)

	botsT = data.Table{
		&bots.ori,
		&bots.dir,
		&bots.fed,
		&bots.col,
		&bots.id,
		&bots.brain,
	}
	bots struct {
		ori   geom.SliceOri2
		dir   geom.SliceVec2
		fed   data.SliceFloat32
		col   gfx.SliceColour
		id    data.SliceInt
		brain SliceBrain
	}

	botIds    = 0
	botHeld   = false
	botHeldId = 0
	botsPause = true
)

func addBot(ori geom.Ori2, dir geom.Vec2, fed float32, col gfx.Colour) {
    nn := nn.MakeNeuralNetwork(botsNumSensors, 2, 3, 20)
    nn.RandomiseWeights()

	botsT.Append(ori, dir, fed, col, botIds, BotBrain{
        network: nn,
    })
	botIds++
}

func spawnPred() {
	spawnPos := geom.Vec2Rand(arena)
	spawnOri := geom.Ori2{ spawnPos.X, spawnPos.Y, geom.AngleRand(), }
	spawnDir := geom.Vec2RandNormal()
	addBot( spawnOri, spawnDir, float32(botsStartFed), gfx.Colour{1, 1, 1, 1},)
}

func start() {
	for i := 0; i < botsStart; i++ {
		spawnPred()
	}
}

func update() {
	for i := range bots.ori {
		for j := range bots.ori {
			if i == j {
				continue
			}

			delta := bots.ori[j].Vec2().Minus(bots.ori[i].Vec2())

			if delta.Len2() > botsSightRadius*botsSightRadius {
				continue
			}

			bearing := delta.Theta() - bots.ori[i].Theta

			if bearing > botsSightWidth || -bearing > botsSightWidth {
				continue
			}

			sensorIdx := int(botsNumSensors/2*bearing/botsSightWidth) + botsNumSensors/2
			activation := (botsSightRadius - delta.Len()) / botsSightRadius

            inputs := bots.brain[i].network.Inputs()
			if activation > bots.brain[i].sensors[sensorIdx] {
			    inputs[sensorIdx] = activation
			}
		}
	}

    var wg sync.WaitGroup
    wg.Add(len(bots.brain))
    for i := range bots.brain {
        go func(){
            bots.brain[i].network.Process()
            bots.col[i].R = bots.brain[i].network.Outputs()[0]
            wg.Done()
        }()
    }
    wg.Wait()

	if botsPause {
		return
	}

	for i := range bots.ori {
		bots.ori[i].PlusEquals(bots.dir[i].ScaledBy(botsSpeed).Ori2())
		bots.ori[i].Theta = bots.dir[i].Theta()
	}
}

func drawBots(w *gfx.WinDraw, tex gfx.TexID, mat geom.Mat3) {
	texCoords := [4]geom.Vec2{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	botsSize := float32(40)
	botsData := make([]float32, 0, 6*8*len(bots.ori))

	for i, ori := range bots.ori {
		botsCol := bots.col[i]
		verts := geom.RectCentred(botsSize, botsSize).Verts()

		for j := range verts {
			verts[j] = ori.Mat3Transform().TimesVec2(verts[j], 1).Vec2()
		}

		for _, j := range [6]int{0, 1, 2, 0, 2, 3} {
			botsData = append(
                botsData,
				verts[j].X, verts[j].Y,
				texCoords[j].X, texCoords[j].Y,
				botsCol.R, botsCol.G, botsCol.B, botsCol.A,
			)
		}
	}

	w.DrawVertexData(botsData, &tex, &mat)
}
