package main

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/kaushikc92/chess"
	"github.com/kaushikc92/chess/uci"
)

const DatabaseUrl string = "mongodb://mongo-0.mongo,mongo-1.mongo:27017"

type UpdatePosition struct {
	Fen string `bson:"fen,omitempty"`
	Depth int `bson:"depth,omitempty"`
}

type RepertoirePosition struct {
	Fen string `bson:"fen,omitempty"`
	BestMove string `bson:"bestMove,omitempty"`
	Depth int `bson:"depth,omitempty"`
	MyMove string `bson:"myMove,omitempty"`
}

func main() {
	for {
		upos, err := popUpdatePosition()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			} else {
				panic(err)
			}
		}
		rpos, err := getRepertoirePosition(upos)
		if err != nil {
			panic(err)
		}
		err = upsertRepertoirePosition(rpos)
		if err != nil {
			panic(err)
		}
	}
}

func popUpdatePosition() (*UpdatePosition, error){
	client, err := mongo.NewClient(options.Client().ApplyURI(DatabaseUrl))
	if err != nil {
		return &UpdatePosition{}, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return &UpdatePosition{}, err
	}
	defer client.Disconnect(ctx)

	collection := client.Database("cagliostro").Collection("updates")
	var upos UpdatePosition
	err = collection.FindOneAndDelete(ctx, bson.M{}).Decode(&upos)
	if err != nil {
		return &UpdatePosition{}, err
	} else {
		return &upos, nil
	}
}

func upsertRepertoirePosition(rpos *RepertoirePosition) error {
	client, err := mongo.NewClient(options.Client().ApplyURI(DatabaseUrl))
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	collection := client.Database("cagliostro").Collection("repertoire")
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"fen": rpos.Fen}
	update := bson.D{{"$set", rpos}}
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getRepertoirePosition(upos *UpdatePosition) (*RepertoirePosition, error) {
	newRpos := RepertoirePosition {
		Fen: upos.Fen,
		Depth: upos.Depth,
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(DatabaseUrl))
	if err != nil {
		return &RepertoirePosition{}, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return &RepertoirePosition{}, err
	}
	defer client.Disconnect(ctx)
	collection := client.Database("cagliostro").Collection("repertoire")
	var rpos RepertoirePosition
	err = collection.FindOne(ctx, bson.M{"fen": upos.Fen}).Decode(&rpos)
	if err == nil {
		newRpos.MyMove = rpos.MyMove
		if newRpos.Depth > rpos.Depth {
			move, err := getMove(newRpos.Fen, newRpos.Depth)
			if err != nil {
				return &RepertoirePosition{}, err
			}
			newRpos.BestMove = move
		} else {
			newRpos.Depth = rpos.Depth
			newRpos.BestMove = rpos.BestMove
		}
	} else {
		move, err := getMove(newRpos.Fen, newRpos.Depth)
		if err != nil {
			return &RepertoirePosition{}, err
		} else {
			newRpos.BestMove = move
			newRpos.MyMove = move
		}
	}
	return &newRpos, nil
}

func getMove(fenString string, depth int) (string, error) {
	fen, err := chess.FEN(fenString)
	if err != nil {
		return "", err
	}
	game := chess.NewGame(fen)
	position := game.Position()
	eng, err := uci.New("stockfish")
	if err != nil {
		return "", err
	}
	defer eng.Close()
	if err != nil {
		return "", err
	}
	setPos := uci.CmdPosition{Position: position}
	setGo := uci.CmdGo{Depth: depth}
	if err := eng.Run(uci.CmdUCINewGame, setPos, setGo); err != nil {
		return "", err
	}
	bestMove := eng.SearchResults().BestMove
	moveStr := bestMove.String()

	return moveStr, nil
}
