package mongodump

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/shelman/mongo-tools-proto/common/db"
	"github.com/shelman/mongo-tools-proto/common/log"
	commonopts "github.com/shelman/mongo-tools-proto/common/options"
	"github.com/shelman/mongo-tools-proto/common/util"
	"github.com/shelman/mongo-tools-proto/mongodump/options"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

type MongoDump struct {
	// basic mongo tool options
	ToolOptions *commonopts.ToolOptions

	InputOptions  *options.InputOptions
	OutputOptions *options.OutputOptions

	SessionProvider *db.SessionProvider

	// useful internals that we don't directly expose as options
	useStdout bool
}

func (dmp *MongoDump) ValidateOptions() error {
	switch {
	case dmp.InputOptions.Query != "" && dmp.ToolOptions.Collection != "":
		return fmt.Errorf("cannot dump using a query without a specific collection")
	}
	return nil
}

func (dmp *MongoDump) Dump() error {
	//TODO -- call proper things, track changes

	//TODO move this outside of this file
	if err := dmp.ValidateOptions(); err != nil {
		return err
	}

	dmp.DumpCollection(dmp.ToolOptions.DB, dmp.ToolOptions.Collection)

	log.Logf(1, "done")

	return nil
}

func (dmp *MongoDump) DumpCollection(db, c string) {
	if dmp.useStdout {
		//XXX
	} else {
		dbFolder := filepath.Join(dmp.OutputOptions.Out, db)
		err := os.MkdirAll(dbFolder, 0666) //TODO const?
		if err != nil {
			util.Exitf(1, "error creating directory `%v`: %v", dbFolder, err)
		}

		outFilepath := filepath.Join(dbFolder, fmt.Sprintf("%v.bson", c))
		out, err := os.Create(outFilepath)
		if err != nil {
			util.Exitf(1, "error creating bson file `%v`: %v", outFilepath, err)
		}
		defer out.Close()

		log.Logf(0, "writing %v.%v to %v", db, c, outFilepath)
		dmp.dumpCollectionToWriter(db, c, bufio.NewWriter(out))

		metadataFilepath := filepath.Join(dbFolder, fmt.Sprintf("%v.metadata.json", c))
		metaOut, err := os.Create(metadataFilepath)
		if err != nil {
			util.Exitf(1, "error creating metadata.json file `%v`: %v", outFilepath, err)
		}
		defer metaOut.Close()

		log.Logf(0, "writing %v.%v metadata to %v", db, c, metadataFilepath)
		dmp.dumpMetadataToWriter(db, c, bufio.NewWriter(metaOut))
	}
}

type Metadata struct {
	Options bson.M   `json:"options,omitempty"`
	Indexes []bson.M `json:"indexes"` //FIXME, order is really important :(
}

func (dmp *MongoDump) dumpMetadataToWriter(db, c string, w *bufio.Writer) {
	session := dmp.SessionProvider.GetSession()

	nsID := fmt.Sprintf("%v.%v", db, c)
	meta := Metadata{}

	// get options
	log.Logf(3, "\treading options for `%v`", nsID)
	namespaceDoc := bson.M{}
	collection := session.DB(db).C("system.namespaces")
	err := collection.Find(bson.M{"name": nsID}).One(&namespaceDoc)
	if err != nil {
		util.Exitf(2, "error finding metadata for collection `%v`: %v", nsID, err)
	}
	if opts, ok := namespaceDoc["options"]; ok {
		meta.Options = opts.(bson.M)
	}

	// get indexes
	log.Logf(3, "\treading indexes for `%v`", nsID)
	collection = session.DB(db).C("system.indexes")
	cursor := collection.Find(bson.M{"ns": nsID}).Iter()
	indexDoc := bson.M{}
	//TODO figure out the best way to represent indexes internally FIXME
	for cursor.Next(&indexDoc) {
		newIndexDoc := bson.M{}
		for k, v := range indexDoc {
			newIndexDoc[k] = v
		}
		meta.Indexes = append(meta.Indexes, newIndexDoc)
	}
	if err := cursor.Err(); err != nil {
		util.Exitf(2, "error finding index data for collection `%v`: %v", nsID, err)
	}

	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		util.Exitf(2, "error writing metadata for collection `%v`: %v", nsID, err)
	}

	w.Write(jsonBytes)
	w.Flush()
}

func (dmp *MongoDump) dumpCollectionToWriter(db, c string, w *bufio.Writer) {
	session := dmp.SessionProvider.GetSession()

	collection := session.DB(db).C(c)

	cursor := collection.Find(bson.M{}).Iter()
	defer cursor.Close()

	buffChan := make(chan []byte)
	go func() {
		for {
			raw := &bson.Raw{}
			if err := cursor.Err(); err != nil {
				log.Logf(0, "Error reading from %v.%v: %v", db, c, err)
			}
			next := cursor.Next(raw)
			if !next {
				close(buffChan)
				return
			}
			buffChan <- raw.Data
		}
	}()

	for {
		//TODO make better use of Next and Error
		buff, alive := <-buffChan
		if alive == false {
			break
		}
		_, err := w.Write(buff)
		if err != nil {
			log.Logf(0, "Error writing to file: %v", err)
		}
	}
	err := w.Flush()
	if err != nil {
		log.Logf(0, "Error flushing buffered writer...")
	}
}
