package repository

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Counter is a collection of documents (shards)
// to realize counter with high frequency.
type Counter struct {
	numShards      int
	collectionName string
	client         *firestore.Client
}

// Shard is a single counter, which is used in a group
// of other shards within Counter.
type Shard struct {
	Count int
}

// initCounter creates a given number of shards as
// subcollection of specified document.
func (c *Counter) initCounter(ctx context.Context) error {
	colRef := c.client.Collection(c.collectionName)

	// Initialize each shard with count=0
	for num := 0; num < c.numShards; num++ {
		shard := Shard{0}

		if _, err := colRef.Doc(strconv.Itoa(num)).Set(ctx, shard); err != nil {
			return fmt.Errorf("Set: %v", err)
		}
	}
	return nil
}

// incrementCounter increments a randomly picked shard.
func (c *Counter) incrementCounter(ctx context.Context) (*firestore.WriteResult, error) {
	docID := strconv.Itoa(rand.Intn(c.numShards))

	shardRef := c.client.Collection(c.collectionName).Doc(docID)
	return shardRef.Update(ctx, []firestore.Update{
		{Path: "Count", Value: firestore.Increment(1)},
	})
}

// getCount returns a total count across all shards.
func (c *Counter) getCount(ctx context.Context) (int64, error) {
	var total int64
	shards := c.client.Collection(c.collectionName).Documents(ctx)
	for {
		doc, err := shards.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("Next: %v", err)
		}

		vTotal := doc.Data()["Count"]
		shardCount, ok := vTotal.(int64)
		if !ok {
			return 0, fmt.Errorf("firestore: invalid dataType %T, want int64", vTotal)
		}
		total += shardCount
	}
	return total, nil
}

func (c *Counter) counterExists(ctx context.Context) bool {
	_, err := c.client.Collection(c.collectionName).Doc("0").Get(ctx)
	return err == nil
}
