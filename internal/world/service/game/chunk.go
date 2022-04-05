package game

import (
	"errors"
	"fmt"
	"math"

	"github.com/Roukii/pock_multiplayer/internal/world/entity"
	"github.com/Roukii/pock_multiplayer/internal/world/entity/player"
	"github.com/Roukii/pock_multiplayer/internal/world/entity/universe"
	"github.com/Roukii/pock_multiplayer/internal/world/service/procedural_generation"
	"github.com/gocql/gocql"
)

func (g *GameService) loadWorldChunks(world *universe.World) error {
	chunks, err := g.ChunkDao.LoadWorldChunk(world.UUID)
	if err != nil {
		fmt.Println("Couldn't load chunks : ", err)
		return err
	}
	fmt.Println("chunks length : ", len(chunks))
	if len(chunks) == 0 {
		return g.generateAndSaveWorldChunks(world)
	}
	world.Chunks = make(map[int]map[int]universe.Chunk)
	for _, chunk := range chunks {
		if world.Chunks[chunk.PositionX] == nil {
			world.Chunks[chunk.PositionX] = make(map[int]universe.Chunk)
		}
		world.Chunks[chunk.PositionX][chunk.PositionY] = *chunk
	}
	return err
}

func (g *GameService) generateAndSaveWorldChunks(world *universe.World) (err error) {
	generator := procedural_generation.NewWorldGenerator(world)
	g.WorldGenerators[world.UUID] = &generator

	fmt.Println("Start generate and save chunks")
	world.Chunks = make(map[int]map[int]universe.Chunk)
	for x := 0; x < world.Length; x++ {
		world.Chunks[x] = make(map[int]universe.Chunk)
		for y := 0; y < world.Width; y++ {
			chunk, err := g.generateChunk(world, entity.Vector2{x, y})
			world.Chunks[x][y] = *chunk
			if err != nil {
				fmt.Println("Error generating chunk : ", x, "/", y, " with error : ", err)
				return err
			}
		}
	}
	fmt.Println("Save chunks to database")
	err = g.saveWorldChunks(world)
	if err != nil {
		fmt.Println("Error saving chunks : ", err)
		return err
	}
	return nil
}

func (g *GameService) GetChunksFromSpawnSpoint(spawnPoint player.SpawnPoint, viewDistance int) ([]*universe.Chunk, error) {
	var chunks []*universe.Chunk
	world, ok := g.Universe.Worlds[spawnPoint.WorldUUID]
	if !ok {
		return nil, errors.New("world not found : " + spawnPoint.WorldUUID)
	}
	fmt.Println("chunks : ", world.Chunks)
	if chunks, ok := world.Chunks[0]; ok {
		fmt.Println(len(chunks))
	}
	if len(world.Chunks) == 0 {
		err := g.loadWorldChunks(&world)
		if err != nil {
			return nil, errors.New("world not found : " + spawnPoint.WorldUUID)
		}
	}
	spawnChunkPosX := int(math.Floor(float64(spawnPoint.Coordinate.Position.X / 25)))
	spawnChunkPosY := int(math.Floor(float64(spawnPoint.Coordinate.Position.Y / 25)))
	for x := -viewDistance; x <= viewDistance; x++ {
		for y := -viewDistance; y <= viewDistance; y++ {
			if chunk, ok := world.Chunks[spawnChunkPosX+x][spawnChunkPosY+y]; ok {
				fmt.Println("add chunk from struct : ", chunk.PositionX, "/", chunk.PositionY)
				chunks = append(chunks, &chunk)
			} else {
				fmt.Println("Couldn't load chunk from pos : ", spawnChunkPosX+x, "/", spawnChunkPosY+y)
			}
		}
	}
	return chunks, nil
}

func (g *GameService) saveWorldChunks(world *universe.World) (err error) {
	for _, chunks := range world.Chunks {
		for _, chunk := range chunks {
			err = g.ChunkDao.Insert(world.UUID, &chunk)
			if err != nil {
				fmt.Println("can't save chunk : ", err)
			}
			fmt.Println("insert chunk to db : ", chunk.PositionX, "/", chunk.PositionY)
		}
	}
	return
}

func (g *GameService) generateChunk(world *universe.World, position entity.Vector2) (*universe.Chunk, error) {
	generator := g.WorldGenerators[world.UUID]
	chunk, err := generator.GenerateChunk(position.X, position.Y)
	if err != nil {
		fmt.Println("error generating chunk : ", err)
		return nil, err
	}
	chunk.UUID = gocql.TimeUUID().String()
	g.ChunkDao.Insert(world.UUID, chunk)
	return chunk, nil
}
