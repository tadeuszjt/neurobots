package main

import "github.com/tadeuszjt/neuralnetwork"

type BotBrain struct {
	sensors [botsNumSensors]float32
    network nn.NeuralNetwork
}

type SliceBrain []BotBrain

func (s *SliceBrain) Len() int {
	return len(*s)
}

func (s *SliceBrain) Swap(i, j int) {
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

func (s *SliceBrain) Delete(i int) {
	end := s.Len() - 1
	if i < end {
		s.Swap(i, end)
	}

	*s = (*s)[:end]
}

func (s *SliceBrain) Append(t interface{}) {
	i, ok := t.(BotBrain)
	if !ok {
		panic("wrong type")
	}

	*s = append(*s, i)
}
