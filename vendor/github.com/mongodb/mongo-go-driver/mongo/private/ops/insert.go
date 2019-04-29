// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package ops

import (
	"context"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo/internal"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/writeconcern"
)

// Insert executes an insert command for the given set of  documents.
func Insert(ctx context.Context, s *SelectedServer, ns Namespace, writeConcern *writeconcern.WriteConcern,
	docs []*bson.Document, options ...options.InsertOptioner) (rdr bson.Reader, err error) {

	if err := ns.validate(); err != nil {
		return nil, err
	}

	command := bson.NewDocument()
	command.Append(bson.C.String("insert", ns.Collection))
	vals := make([]*bson.Value, 0, len(docs))
	for _, doc := range docs {
		vals = append(vals, bson.AC.Document(doc))
	}
	command.Append(bson.C.ArrayFromElements("documents", vals...))

	for _, option := range options {
		if option == nil {
			continue
		}
		option.Option(command)
	}

	if writeConcern != nil {
		elem, err := writeConcern.MarshalBSONElement()
		if err != nil {
			return nil, err
		}
		command.Append(elem)
	}

	rdr, err = runMustUsePrimary(ctx, s, ns.DB, command)
	if err != nil {
		return nil, internal.WrapError(err, "failed to execute insert")
	}

	return rdr, err
}
