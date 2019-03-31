package mango

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	*mongo.Collection
}

func (coll *Collection) Clone(opts ...*options.CollectionOptions) (*Collection, error) {
	cc, err := coll.Collection.Clone(opts...)
	if err != nil {
		return nil, err
	}
	return &Collection{Collection: cc}, err
}

func (coll *Collection) Database() *Database {
	db := coll.Collection.Database()
	return &Database{Database: db}
}

func (coll *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel,
		opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.BulkWrite", modulestr))
	defer span.end(ctx)

	bwres, err := coll.Collection.BulkWrite(ctx, models, opts...)
	if err != nil {
		span.setError(err)
	}
	return bwres, err
}

func (coll *Collection) InsertOne(ctx context.Context, doc interface{},
		opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.InsertOne", modulestr))
	defer span.end(ctx)

	ior, err := coll.Collection.InsertOne(ctx, doc, opts...)
	if err != nil {
		span.setError(err)
	}
	return ior, err
}

func (coll *Collection) InsertMany(ctx context.Context, doc []interface{},
		opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.InsertMany", modulestr))
	defer span.end(ctx)

	imr, err := coll.Collection.InsertMany(ctx, doc, opts...)
	if err != nil {
		span.setError(err)
	}
	return imr, err
}

func (coll *Collection) DeleteOne(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.DeleteOne", modulestr))
	defer span.end(ctx)

	dr, err := coll.Collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return dr, err
}

// DeleteMany deletes multiple documents from the collection.
func (coll *Collection) DeleteMany(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.DeleteMany", modulestr))
	defer span.end(ctx)

	dr, err := coll.Collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return dr, err
}

// UpdateOne updates a single document in the collection.
func (coll *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.UpdateOne", modulestr))
	defer span.end(ctx)

	ur, err := coll.Collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		span.setError(err)
	}
	return ur, err
}

// UpdateMany updates multiple documents in the collection.
func (coll *Collection) UpdateMany(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.UpdateMany", modulestr))
	defer span.end(ctx)

	ur, err := coll.Collection.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		span.setError(err)
	}
	return ur, err
}

// ReplaceOne replaces a single document in the collection.
func (coll *Collection) ReplaceOne(ctx context.Context, filter interface{},
		replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.ReplaceOne", modulestr))
	defer span.end(ctx)

	ur, err := coll.Collection.ReplaceOne(ctx, filter, replacement, opts...)
	if err != nil {
		span.setError(err)
	}
	return ur, err
}

// Aggregate runs an aggregation framework pipeline.
//
// See https://docs.mongodb.com/manual/aggregation/.
func (coll *Collection) Aggregate(ctx context.Context, pipeline interface{},
		opts ...*options.AggregateOptions) (*mongo.Cursor, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.Aggregate", modulestr))
	defer span.end(ctx)

	cursor, err := coll.Collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		span.setError(err)
	}
	return cursor, err
}

// CountDocuments gets the number of documents matching the filter.
func (coll *Collection) CountDocuments(ctx context.Context, filter interface{},
		opts ...*options.CountOptions) (int64, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.CountDocuments", modulestr))
	defer span.end(ctx)

	count, err := coll.Collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return count, err
}

// EstimatedDocumentCount gets an estimate of the count of documents in a collection using collection metadata.
func (coll *Collection) EstimatedDocumentCount(ctx context.Context,
		opts ...*options.EstimatedDocumentCountOptions) (int64, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.EstimatedDocumentCount", modulestr))
	defer span.end(ctx)

	count, err := coll.Collection.EstimatedDocumentCount(ctx, opts...)
	if err != nil {
		span.setError(err)
	}
	return count, err
}

// Distinct finds the distinct values for a specified field across a single
// collection.
func (coll *Collection) Distinct(ctx context.Context, fieldName string, filter interface{},
		opts ...*options.DistinctOptions) ([]interface{}, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.Distinct", modulestr))
	defer span.end(ctx)

	ret, err := coll.Collection.Distinct(ctx, fieldName, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return ret, err
}

// Find finds the documents matching a model.
func (coll *Collection) Find(ctx context.Context, filter interface{},
		opts ...*options.FindOptions) (*mongo.Cursor, error) {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.Find", modulestr))
	defer span.end(ctx)

	cursor, err := coll.Collection.Find(ctx, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return cursor, err
}

// FindOne returns up to one document that matches the model.
func (coll *Collection) FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) *mongo.SingleResult {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.FindOne", modulestr))
	defer span.end(ctx)

	return coll.Collection.FindOne(ctx, filter, opts...)
}

// FindOneAndDelete find a single document and deletes it, returning the
// original in result.
func (coll *Collection) FindOneAndDelete(ctx context.Context, filter interface{},
		opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.FindOneAndDelete", modulestr))
	defer span.end(ctx)

	return coll.Collection.FindOneAndDelete(ctx, filter, opts...)
}

// FindOneAndReplace finds a single document and replaces it, returning either
// the original or the replaced document.
func (coll *Collection) FindOneAndReplace(ctx context.Context, filter interface{},
		replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.FindOneAndReplace", modulestr))
	defer span.end(ctx)

	return coll.Collection.FindOneAndReplace(ctx, filter, replacement, opts...)
}

// FindOneAndUpdate finds a single document and updates it, returning either
// the original or the updated.
func (coll *Collection) FindOneAndUpdate(ctx context.Context, filter interface{},
		update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {

	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.FindOneAndUpdate", modulestr))
	defer span.end(ctx)

	return coll.Collection.FindOneAndUpdate(ctx, filter, update, opts...)
}

// Drop drops this collection from database.
func (coll *Collection) Drop(ctx context.Context) error {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Collection.Drop", modulestr))
	defer span.end(ctx)

	return coll.Collection.Drop(ctx)
}
